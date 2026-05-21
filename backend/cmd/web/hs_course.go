package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/courseman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
	"github.com/google/uuid"
)

func getUserCourses(w http.ResponseWriter, r *http.Request) {
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("size"))

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 25
	}

	filter := new(courseman.Filter)
	filter.UserID = loggedUser.ID
	filter.Keyword = q.Get("keyword")
	filter.OrderBy = q.Get("order_by")

	course, total, err := app.Courses.GetAll(filter, page, size, "Notes")
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, map[string]interface{}{
		"items": course,
		"total": total,
	})
}

type Section struct {
	SectionName string `json:"section_name"`
	StartPage   int    `json:"start_page"`
	EndPage     int    `json:"end_page"`
}

func saveCourse(w http.ResponseWriter, r *http.Request) {
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	if !parseLimitedMultipartForm(w, r) {
		return
	}

	course := new(courseman.Course)

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		oapi.CustomError(w, http.StatusBadRequest, "Title is required")
		return
	}

	description := strings.TrimSpace(r.FormValue("description"))
	if description == "" {
		oapi.CustomError(w, http.StatusBadRequest, "Description is required")
		return
	}

	icon := strings.TrimSpace(r.FormValue("icon"))
	if icon == "" {
		icon = "BookOpen"
	}

	course.Title = title
	course.Description = description
	course.Icon = icon
	course.UserID = loggedUser.ID
	course.Status = courseman.STATUS_IN_PROGRESS

	courseDir := filepath.Join(app.Config.FilePath, uuid.NewString())
	courseDir, _ = filepath.Abs(courseDir)
	course.ContainerPath = courseDir

	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		bookDir := filepath.Join(courseDir, "book")
		if err := os.MkdirAll(bookDir, 0755); err != nil {
			oapi.ServerError(w, err)
			return
		}

		path, err := validateAndSaveFileToDisk(header, bookDir)
		if err != nil {
			_ = os.RemoveAll(courseDir)
			oapi.CustomError(w, http.StatusBadRequest, err.Error())
			return
		}

		course.FilePath = path
		course.HasBook = true
		var sections []*courseman.Section
		sectionsRaw := strings.TrimSpace(r.FormValue("sections"))
		if sectionsRaw != "" {
			if err := json.Unmarshal([]byte(sectionsRaw), &sections); err != nil {
				_ = os.RemoveAll(courseDir)
				oapi.CustomError(w, http.StatusBadRequest, "Invalid sections format")
				return
			}
		}

		course.Sections = sections
	}

	savedCourse, err := app.Courses.Save(course)
	if err != nil {
		_ = os.RemoveAll(courseDir)
		oapi.ServerError(w, err)
		return
	}

	var note *noteman.Note
	if strings.TrimSpace(savedCourse.FilePath) != "" {
		if len(course.Sections) > 0 {
			for _, s := range course.Sections {
				note = &noteman.Note{
					CourseID:      savedCourse.ID,
					Title:         s.SectionName,
					IsFromBook:    true,
					StartPage:     s.StartPage,
					EndPage:       s.EndPage,
					FilePath:      savedCourse.FilePath,
					Status:        noteman.STATUS_IN_PROGRESS,
					ProcessStatus: noteman.PROCESS_STATUS_PROCESSING,
				}

				savedNote, err := app.Notes.Save(note)
				if err != nil {
					oapi.ServerError(w, err)
					return
				}

				go processNote(savedNote)
			}
		} else {
			note = &noteman.Note{
				CourseID:      savedCourse.ID,
				Title:         savedCourse.Title,
				IsFromBook:    false,
				FilePath:      savedCourse.FilePath,
				Status:        noteman.STATUS_IN_PROGRESS,
				ProcessStatus: noteman.PROCESS_STATUS_PROCESSING,
			}

			savedNote, err := app.Notes.Save(note)
			if err != nil {
				oapi.ServerError(w, err)
				return
			}

			go processNote(savedNote)
		}
	}

	oapi.SendResp(w, savedCourse)
}

func getCourse(w http.ResponseWriter, r *http.Request) {
	chosenCourse := r.Context().Value(app.ContextKeyChosenCourse).(*courseman.Course)
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)
	if !canAccessCourse(loggedUser, chosenCourse) {
		oapi.Forbidden(w)
		return
	}

	oapi.SendResp(w, chosenCourse)
}

type updateCoursePayload struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Progress    *int    `json:"progress"`
	IsPublic    *bool   `json:"is_public"`
	Icon        *string `json:"icon"`
}

func updateCourse(w http.ResponseWriter, r *http.Request) {
	chosenCourse := r.Context().Value(app.ContextKeyChosenCourse).(*courseman.Course)
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)
	if !canAccessCourse(loggedUser, chosenCourse) {
		oapi.Forbidden(w)
		return
	}

	var data updateCoursePayload
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	if data.Title != nil {
		title := strings.TrimSpace(*data.Title)
		if title == "" {
			oapi.CustomError(w, http.StatusBadRequest, "Title is required")
			return
		}
		chosenCourse.Title = title
	}
	if data.Description != nil {
		chosenCourse.Description = strings.TrimSpace(*data.Description)
	}
	if data.Status != nil {
		status := strings.TrimSpace(*data.Status)
		if status != courseman.STATUS_IN_PROGRESS && status != courseman.STATUS_COMPLETED {
			oapi.CustomError(w, http.StatusBadRequest, "Invalid status")
			return
		}
		chosenCourse.Status = status
	}
	if data.Progress != nil {
		progress := *data.Progress
		if progress < 0 {
			progress = 0
		}
		if progress > 100 {
			progress = 100
		}
		chosenCourse.Progress = progress
	}
	if data.IsPublic != nil {
		chosenCourse.IsPublic = *data.IsPublic
	}
	if data.Icon != nil {
		icon := strings.TrimSpace(*data.Icon)
		if icon != "" {
			chosenCourse.Icon = icon
		}
	}

	savedCourse, err := app.Courses.Save(chosenCourse)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, savedCourse)
}

func deleteCourse(w http.ResponseWriter, r *http.Request) {
	chosenCourse := r.Context().Value(app.ContextKeyChosenCourse).(*courseman.Course)
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)
	if !canAccessCourse(loggedUser, chosenCourse) {
		oapi.Forbidden(w)
		return
	}

	if err := app.Courses.Delete(chosenCourse.ID); err != nil {
		oapi.ServerError(w, err)
		return
	}
	if chosenCourse.ContainerPath != "" {
		if err := os.RemoveAll(chosenCourse.ContainerPath); err != nil {
			app.ErrorLog.Println("failed to remove course files: ", err)
		}
	}

	oapi.SendResp(w, chosenCourse)
}

func canAccessCourse(user *userman.User, course *courseman.Course) bool {
	return user.Role == userman.ROLE_ADMIN || course.UserID == user.ID
}

func validateAndSaveFileToDisk(fh *multipart.FileHeader, dir string) (string, error) {
	if fh.Size > maxUploadFileSize {
		return "", fmt.Errorf("file too large: max 50MB allowed")
	}

	file, err := fh.Open()
	if err != nil {
		app.ErrorLog.Println(err)
		return "", err
	}
	defer file.Close()

	buff := make([]byte, 512)
	n, err := file.Read(buff)
	if err != nil && err != io.EOF {
		return "", err
	}

	mimeType := http.DetectContentType(buff[:n])

	allowed := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}

	if !allowed[mimeType] {
		return "", fmt.Errorf("unsupported file type: %s", mimeType)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	dstPath := filepath.Join(dir, fmt.Sprintf("%v%v", 0, filepath.Ext(fh.Filename)))

	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = dstFile.Close()
	}()

	if _, err := io.Copy(dstFile, file); err != nil {
		return "", err
	}

	return dstPath, nil
}
