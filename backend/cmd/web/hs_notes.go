package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/courseman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

type updateNotePayload struct {
	Title   *string `json:"title"`
	Summary *string `json:"summary"`
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
