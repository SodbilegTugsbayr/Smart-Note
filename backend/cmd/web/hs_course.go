package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/courseman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

func getCourse(w http.ResponseWriter, r *http.Request) {
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

	course, total, err := app.Courses.GetAll(filter, page, size)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, map[string]interface{}{
		"items": course,
		"total": total,
	})
}

func saveCourse(w http.ResponseWriter, r *http.Request) {
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	body := new(courseman.Course)

	body.ID, _ = strconv.Atoi(r.FormValue("id"))
	body.Title = strings.TrimSpace(r.FormValue("title"))
	body.Status = r.FormValue("status")

	body.UserID = loggedUser.ID

	if body.Title == "" {
		oapi.CustomError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if body.Status == "" {
		body.Status = "in_progress"
	}

	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()

		fileName := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
		body.FilePath = fileName
	}

	if body.ID == 0 {
		_, err = app.Courses.Save(body)
	} else {
		existing, fetchErr := app.Courses.GetByID(body.ID)
		if fetchErr != nil {
			oapi.CustomError(w, http.StatusNotFound, "Course not found")
			return
		}
		if existing.UserID != loggedUser.ID {
			oapi.Forbidden(w)
			return
		}

		if body.FilePath == "" {
			body.FilePath = existing.FilePath
		}

		_, err = app.Courses.Save(body)
	}

	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, body)
}
