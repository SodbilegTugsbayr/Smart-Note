package main

import (
	"encoding/json"
	"errors"
	"io"
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

type createNotePayload struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

type updateNotePayload struct {
	Title   *string `json:"title"`
	Summary *string `json:"summary"`
}

func createCourseNote(w http.ResponseWriter, r *http.Request) {
	chosenCourse := r.Context().Value(app.ContextKeyChosenCourse).(*courseman.Course)
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)
	if !canAccessCourse(loggedUser, chosenCourse) {
		oapi.Forbidden(w)
		return
	}

	var data createNotePayload
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil && !errors.Is(err, io.EOF) {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	title := strings.TrimSpace(data.Title)
	if title == "" {
		title = "Шинэ тэмдэглэл"
	}

	note := &noteman.Note{
		CourseID:   chosenCourse.ID,
		Title:      title,
		IsFromBook: false,
		Summary:    strings.TrimSpace(data.Summary),
		Status:     noteman.STATUS_IN_PROGRESS,
	}

	savedNote, err := app.Notes.Save(note)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	savedNote.PrepareResponse()
	oapi.SendResp(w, savedNote)
}

func updateNote(w http.ResponseWriter, r *http.Request) {
	chosenNote := r.Context().Value(app.ContextKeyChosenNote).(*noteman.Note)
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	course, err := app.Courses.GetByID(chosenNote.CourseID)
	if err != nil {
		if errors.Is(err, courseman.ErrNotFound) {
			oapi.NotFound(w)
			return
		}
		oapi.ServerError(w, err)
		return
	}
	if !canAccessCourse(loggedUser, course) {
		oapi.Forbidden(w)
		return
	}

	var data updateNotePayload
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
		chosenNote.Title = title
	}
	if data.Summary != nil {
		chosenNote.Summary = strings.TrimSpace(*data.Summary)
	}

	savedNote, err := app.Notes.Save(chosenNote)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	savedNote.PrepareResponse()
	oapi.SendResp(w, savedNote)
}

func uploadNoteFile(w http.ResponseWriter, r *http.Request) {
	chosenNote := r.Context().Value(app.ContextKeyChosenNote).(*noteman.Note)
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	course, err := app.Courses.GetByID(chosenNote.CourseID)
	if err != nil {
		if errors.Is(err, courseman.ErrNotFound) {
			oapi.NotFound(w)
			return
		}
		oapi.ServerError(w, err)
		return
	}
	if !canAccessCourse(loggedUser, course) {
		oapi.Forbidden(w)
		return
	}
	if chosenNote.IsFromBook {
		oapi.CustomError(w, http.StatusBadRequest, "Book notes cannot have another file attached")
		return
	}
	if strings.TrimSpace(chosenNote.FilePath) != "" {
		oapi.CustomError(w, http.StatusBadRequest, "Note already has an attached file")
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		oapi.CustomError(w, http.StatusBadRequest, "File is required")
		return
	}
	_ = file.Close()

	courseDir := course.ContainerPath
	if strings.TrimSpace(courseDir) == "" {
		courseDir = filepath.Join(app.Config.FilePath, uuid.NewString())
		courseDir, _ = filepath.Abs(courseDir)
		course.ContainerPath = courseDir
		if _, err := app.Courses.Save(course); err != nil {
			oapi.ServerError(w, err)
			return
		}
	}

	uploadDir := filepath.Join(courseDir, "notes", strconv.Itoa(chosenNote.ID))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		oapi.ServerError(w, err)
		return
	}

	path, err := validateAndSaveFileToDisk(header, uploadDir)
	if err != nil {
		_ = os.RemoveAll(uploadDir)
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	chosenNote.FilePath = path
	chosenNote.IsFromBook = false
	chosenNote.Summary = ""
	chosenNote.RawContent = ""
	chosenNote.KeyConcepts = nil
	chosenNote.FlashCards = nil
	chosenNote.Status = noteman.STATUS_IN_PROGRESS
	chosenNote.ProcessStatus = noteman.PROCESS_STATUS_PROCESSING

	savedNote, err := app.Notes.Save(chosenNote)
	if err != nil {
		_ = os.RemoveAll(uploadDir)
		oapi.ServerError(w, err)
		return
	}

	processNote(savedNote)
	savedNote.PrepareResponse()
	oapi.SendResp(w, savedNote)
}

func courseForRequest(w http.ResponseWriter, r *http.Request, courseID int) (*courseman.Course, bool) {
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)
	course, err := app.Courses.GetByID(courseID, "Notes")
	if err != nil {
		if errors.Is(err, courseman.ErrNotFound) {
			oapi.NotFound(w)
			return nil, false
		}
		oapi.ServerError(w, err)
		return nil, false
	}
	if !canAccessCourse(loggedUser, course) {
		oapi.Forbidden(w)
		return nil, false
	}

	return course, true
}
