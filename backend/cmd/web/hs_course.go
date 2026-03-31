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
	filter.OrderBy = q.Get("order_by")

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

	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		fileName := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
		course.FilePath = fileName
	}

	_, err = app.Courses.Save(course)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, course)
}
