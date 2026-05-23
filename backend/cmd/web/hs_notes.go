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
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/quizman"
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

type quizSubmissionAnswer struct {
	QuizID int    `json:"quiz_id"`
	Answer string `json:"answer"`
}

type quizSubmissionPayload struct {
	Answers []quizSubmissionAnswer `json:"answers"`
}

type quizSubmissionResponse struct {
	Score        int                 `json:"score"`
	Total        int                 `json:"total"`
	Percentage   int                 `json:"percentage"`
	Passed       bool                `json:"passed"`
	Regenerating bool                `json:"regenerating"`
	Message      string              `json:"message"`
	Result       *quizman.QuizResult `json:"result,omitempty"`
	Note         *noteman.Note       `json:"note,omitempty"`
	Course       *courseman.Course   `json:"course,omitempty"`
}

const quizPassPercent = 90

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

	course, ok := noteCourseForRequest(w, chosenNote)
	if !ok {
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

func deleteNote(w http.ResponseWriter, r *http.Request) {
	chosenNote := r.Context().Value(app.ContextKeyChosenNote).(*noteman.Note)
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	course, ok := noteCourseForRequest(w, chosenNote)
	if !ok {
		return
	}
	if !canAccessCourse(loggedUser, course) {
		oapi.Forbidden(w)
		return
	}

	if err := deleteNoteWithChildren(chosenNote); err != nil {
		oapi.ServerError(w, err)
		return
	}

	chosenNote.PrepareResponse()
	oapi.SendResp(w, chosenNote)
}

func uploadNoteFile(w http.ResponseWriter, r *http.Request) {
	chosenNote := r.Context().Value(app.ContextKeyChosenNote).(*noteman.Note)
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	course, ok := noteCourseForRequest(w, chosenNote)
	if !ok {
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

	if !parseLimitedMultipartForm(w, r) {
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

	noteToProcess := *savedNote
	go processNote(&noteToProcess, loggedUser.ID)

	savedNote.PrepareResponse()
	oapi.SendRespStatus(w, http.StatusAccepted, savedNote)
}

func getNoteQuizzes(w http.ResponseWriter, r *http.Request) {
	chosenNote := r.Context().Value(app.ContextKeyChosenNote).(*noteman.Note)
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	course, ok := noteCourseForRequest(w, chosenNote)
	if !ok {
		return
	}
	if !canAccessCourse(loggedUser, course) {
		oapi.Forbidden(w)
		return
	}

	quizzes, total, err := app.Quizzes.GetAll(&quizman.Filter{
		NoteID:  chosenNote.ID,
		OrderBy: "id",
	}, 1, 100)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, map[string]interface{}{
		"items": quizzes,
		"total": total,
	})
}

func submitNoteQuiz(w http.ResponseWriter, r *http.Request) {
	chosenNote := r.Context().Value(app.ContextKeyChosenNote).(*noteman.Note)
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)

	course, ok := noteCourseForRequest(w, chosenNote)
	if !ok {
		return
	}
	if !canAccessCourse(loggedUser, course) {
		oapi.Forbidden(w)
		return
	}

	var data quizSubmissionPayload
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		oapi.CustomError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(data.Answers) == 0 {
		oapi.CustomError(w, http.StatusBadRequest, "Answers are required")
		return
	}

	quizzes, _, err := app.Quizzes.GetAll(&quizman.Filter{
		NoteID:  chosenNote.ID,
		OrderBy: "id",
	}, 1, 0)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	if len(quizzes) == 0 {
		oapi.CustomError(w, http.StatusBadRequest, "No quizzes found for note")
		return
	}

	quizByID := make(map[int]*quizman.Quiz, len(quizzes))
	for _, quiz := range quizzes {
		quizByID[quiz.ID] = quiz
	}

	submittedAnswers := make(map[int]string, len(data.Answers))
	for _, answer := range data.Answers {
		if answer.QuizID <= 0 {
			oapi.CustomError(w, http.StatusBadRequest, "Invalid quiz answer")
			return
		}
		if _, ok := quizByID[answer.QuizID]; !ok {
			oapi.CustomError(w, http.StatusBadRequest, "Quiz answer does not belong to this note")
			return
		}
		submittedAnswers[answer.QuizID] = strings.TrimSpace(answer.Answer)
	}

	score := 0
	for _, quiz := range quizzes {
		if submittedAnswers[quiz.ID] == strings.TrimSpace(quiz.CorrectAnswer) {
			score++
		}
	}

	total := len(quizzes)
	percentage := (score*100 + total/2) / total
	passed := score*100 >= quizPassPercent*total

	result, err := app.Quizzes.SaveResult(&quizman.QuizResult{
		UserID:     loggedUser.ID,
		NoteID:     chosenNote.ID,
		Score:      score,
		Total:      total,
		Percentage: percentage,
		Passed:     passed,
	})
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	if passed {
		chosenNote.Status = noteman.STATUS_COMPLETED
	} else {
		chosenNote.Status = noteman.STATUS_IN_PROGRESS
	}

	savedNote, err := app.Notes.Save(chosenNote)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	savedNote.PrepareResponse()

	updatedCourse, err := syncCourseProgress(chosenNote.CourseID)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	resp := quizSubmissionResponse{
		Score:      score,
		Total:      total,
		Percentage: percentage,
		Passed:     passed,
		Result:     result,
		Note:       savedNote,
		Course:     updatedCourse,
	}

	if passed {
		resp.Message = "Тест амжилттай давлаа. Хичээлийн явц шинэчлэгдлээ."
		oapi.SendResp(w, resp)
		return
	}

	resp.Regenerating = true
	resp.Message = "Тестийн оноо 90%-аас их биш байна. Шинэ тест үүсгэж байна."

	noteToRegenerate := *savedNote
	go regenerateNoteQuizzes(&noteToRegenerate, loggedUser.ID)

	oapi.SendRespStatus(w, http.StatusAccepted, resp)
}

func syncCourseProgress(courseID int) (*courseman.Course, error) {
	course, err := app.Courses.GetByID(courseID, "Notes")
	if err != nil {
		return nil, err
	}

	total := len(course.Notes)
	completed := 0
	for _, note := range course.Notes {
		if note.Status == noteman.STATUS_COMPLETED {
			completed++
		}
	}

	progress := 0
	if total > 0 {
		progress = (completed*100 + total/2) / total
	}

	course.Progress = progress
	if total > 0 && completed == total {
		course.Status = courseman.STATUS_COMPLETED
	} else {
		course.Status = courseman.STATUS_IN_PROGRESS
	}

	notes := course.Notes
	course.Notes = nil
	if _, err := app.Courses.Save(course); err != nil {
		return nil, err
	}
	course.Notes = notes

	return course, nil
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

func noteCourseForRequest(w http.ResponseWriter, note *noteman.Note) (*courseman.Course, bool) {
	course, err := app.Courses.GetByID(note.CourseID)
	if err != nil {
		if errors.Is(err, courseman.ErrNotFound) {
			oapi.NotFound(w)
			return nil, false
		}
		oapi.ServerError(w, err)
		return nil, false
	}

	return course, true
}

func deleteNoteWithChildren(note *noteman.Note) error {
	quizzes, _, err := app.Quizzes.GetAll(&quizman.Filter{NoteID: note.ID}, 1, 0)
	if err != nil {
		return err
	}
	for _, quiz := range quizzes {
		if err := app.Quizzes.Delete(quiz.ID); err != nil {
			return err
		}
	}

	return app.Notes.Delete(note.ID)
}
