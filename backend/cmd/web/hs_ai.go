package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/courseman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

type courseIDPayload struct {
	CourseID int `json:"course_id"`
}

type chatPayload struct {
	CourseID int    `json:"course_id"`
	Question string `json:"question"`
}

type flashcardPayload struct {
	CourseID   int    `json:"course_id"`
	NoteID     int    `json:"note_id"`
	Term       string `json:"term"`
	Definition string `json:"definition"`
}

func processCourseNotes(w http.ResponseWriter, r *http.Request) {
	var data courseIDPayload
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	course, ok := courseForRequest(w, r, data.CourseID)
	if !ok {
		return
	}
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	for _, note := range course.Notes {
		if noteHasSourceFile(note) && (note.ProcessStatus != noteman.PROCESS_STATUS_COMPLETED || strings.TrimSpace(note.Summary) == "") {
			if err := enqueueNoteProcessing(note.ID, loggedUser.ID); err != nil {
				app.ErrorLog.Println("failed to enqueue note processing: ", err)
			}
		}
	}

	refreshed, err := app.Courses.GetByID(course.ID, "Notes")
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, map[string]interface{}{
		"summary": buildCourseContext(refreshed),
		"course":  refreshed,
	})
}

func generateCourseFlashcards(w http.ResponseWriter, r *http.Request) {
	var data courseIDPayload
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	course, ok := courseForRequest(w, r, data.CourseID)
	if !ok {
		return
	}
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	for _, note := range course.Notes {
		if noteHasSourceFile(note) && len(note.FlashCards) == 0 {
			if err := enqueueNoteProcessing(note.ID, loggedUser.ID); err != nil {
				app.ErrorLog.Println("failed to enqueue note processing: ", err)
			}
		}
	}

	refreshed, err := app.Courses.GetByID(course.ID, "Notes")
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, refreshed)
}

func addFlashcard(w http.ResponseWriter, r *http.Request) {
	var data flashcardPayload
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	term := strings.TrimSpace(data.Term)
	definition := strings.TrimSpace(data.Definition)
	if term == "" || definition == "" {
		oapi.CustomError(w, http.StatusBadRequest, "Term and definition are required")
		return
	}

	course, ok := courseForRequest(w, r, data.CourseID)
	if !ok {
		return
	}
	if len(course.Notes) == 0 {
		oapi.CustomError(w, http.StatusBadRequest, "Course has no notes")
		return
	}

	target := course.Notes[0]
	if data.NoteID > 0 {
		for _, note := range course.Notes {
			if note.ID == data.NoteID {
				target = note
				break
			}
		}
	}

	target.FlashCards = append(target.FlashCards, &noteman.FlashCard{
		Question: term,
		Answer:   definition,
	})

	savedNote, err := app.Notes.Save(target)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, savedNote)
}

func askCourseChat(w http.ResponseWriter, r *http.Request) {
	var data chatPayload
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}

	question := strings.TrimSpace(data.Question)
	if question == "" {
		oapi.CustomError(w, http.StatusBadRequest, "Question is required")
		return
	}

	course, ok := courseForRequest(w, r, data.CourseID)
	if !ok {
		return
	}

	context := buildCourseContext(course)
	if strings.TrimSpace(context) == "" {
		oapi.CustomError(w, http.StatusBadRequest, "Processed note content is required before chat")
		return
	}

	answer, err := app.EguneService.AnswerQuestion(context, question)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, map[string]string{"answer": answer})
}

func buildCourseContext(course *courseman.Course) string {
	var builder strings.Builder
	const maxContextLength = 16000

	for _, note := range course.Notes {
		if strings.TrimSpace(note.Title) != "" {
			builder.WriteString(fmt.Sprintf("## %s\n", note.Title))
		}
		if strings.TrimSpace(note.Summary) != "" {
			builder.WriteString(note.Summary)
			builder.WriteString("\n")
		}
		for _, concept := range note.KeyConcepts {
			builder.WriteString(fmt.Sprintf("- %s: %s\n", concept.Concept, concept.Definition))
		}
		if strings.TrimSpace(note.RawContent) != "" && builder.Len() < maxContextLength/2 {
			builder.WriteString(note.RawContent)
			builder.WriteString("\n")
		}
		if builder.Len() >= maxContextLength {
			return truncateRunes(builder.String(), maxContextLength)
		}
	}

	return builder.String()
}

func noteHasSourceFile(note *noteman.Note) bool {
	return strings.TrimSpace(note.FilePath) != ""
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
