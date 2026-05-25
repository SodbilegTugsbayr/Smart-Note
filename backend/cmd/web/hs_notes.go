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
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
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

type quizResponse struct {
	ID       int      `json:"id"`
	NoteID   int      `json:"note_id"`
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

type quizAnswerResult struct {
	QuizID         int      `json:"quiz_id"`
	Question       string   `json:"question"`
	Options        []string `json:"options"`
	SelectedAnswer string   `json:"selected_answer"`
	CorrectAnswer  string   `json:"correct_answer"`
	Correct        bool     `json:"correct"`
}

type quizSubmissionResponse struct {
	Score        int                 `json:"score"`
	Total        int                 `json:"total"`
	Percentage   int                 `json:"percentage"`
	Passed       bool                `json:"passed"`
	Regenerating bool                `json:"regenerating"`
	Message      string              `json:"message"`
	Answers      []quizAnswerResult  `json:"answers"`
	Result       *quizman.QuizResult `json:"result,omitempty"`
	Note         *noteman.Note       `json:"note,omitempty"`
	Course       *courseman.Course   `json:"course,omitempty"`
}

const quizPassPercent = 90
const maxNoteUploadPDFPages = 50
const largeNoteUploadMessage = "Файл хэтэрхий том байна. Хичээлийн ном оруулах хэсгээр файлыг оруулна уу"

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
	if _, err := syncCourseProgress(chosenCourse.ID); err != nil {
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
	if _, err := syncCourseProgress(chosenNote.CourseID); err != nil {
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
	isPDF, err := isPDFFile(path)
	if err != nil {
		_ = os.RemoveAll(uploadDir)
		oapi.CustomError(w, http.StatusBadRequest, "Файл уншихад алдаа гарлаа")
		return
	}
	if isPDF {
		pageCount, err := pdfapi.PageCountFile(path)
		if err != nil {
			_ = os.RemoveAll(uploadDir)
			oapi.CustomError(w, http.StatusBadRequest, "PDF файл уншихад алдаа гарлаа")
			return
		}
		if pageCount > maxNoteUploadPDFPages {
			_ = os.RemoveAll(uploadDir)
			oapi.CustomError(w, http.StatusBadRequest, largeNoteUploadMessage)
			return
		}
	}

	chosenNote.FilePath = path
	chosenNote.IsFromBook = false
	chosenNote.Summary = ""
	chosenNote.RawContent = ""
	chosenNote.KeyConcepts = nil
	chosenNote.FlashCards = nil
	chosenNote.Status = noteman.STATUS_IN_PROGRESS
	chosenNote.ProcessStatus = noteman.PROCESS_STATUS_QUEUED

	savedNote, err := app.Notes.Save(chosenNote)
	if err != nil {
		_ = os.RemoveAll(uploadDir)
		oapi.ServerError(w, err)
		return
	}
	if _, err := syncCourseProgress(chosenNote.CourseID); err != nil {
		oapi.ServerError(w, err)
		return
	}

	if err := enqueueNoteProcessing(savedNote.ID, loggedUser.ID); err != nil {
		oapi.ServerError(w, err)
		return
	}

	savedNote.PrepareResponse()
	oapi.SendRespStatus(w, http.StatusAccepted, savedNote)
}

func isPDFFile(path string) (bool, error) {
	if strings.EqualFold(filepath.Ext(path), ".pdf") {
		return true, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buff := make([]byte, 512)
	n, err := file.Read(buff)
	if err != nil && err != io.EOF {
		return false, err
	}

	return http.DetectContentType(buff[:n]) == "application/pdf", nil
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
	results, err := app.Quizzes.GetResults(loggedUser.ID, chosenNote.ID)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	var latestResult *quizman.QuizResult
	if len(results) > 0 {
		latestResult = results[0]
	}

	oapi.SendResp(w, map[string]interface{}{
		"items":         publicQuizResponses(quizzes),
		"total":         total,
		"latest_result": latestResult,
		"results":       results,
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
	answerResults := make([]quizAnswerResult, 0, len(quizzes))
	for _, quiz := range quizzes {
		selectedAnswer := submittedAnswers[quiz.ID]
		correctAnswer := strings.TrimSpace(quiz.CorrectAnswer)
		correct := selectedAnswer == correctAnswer
		if correct {
			score++
		}
		answerResults = append(answerResults, quizAnswerResult{
			QuizID:         quiz.ID,
			Question:       quiz.Question,
			Options:        quiz.Options,
			SelectedAnswer: selectedAnswer,
			CorrectAnswer:  correctAnswer,
			Correct:        correct,
		})
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
		Answers:    quizResultAnswers(answerResults),
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
		Answers:    answerResults,
		Note:       savedNote,
		Course:     updatedCourse,
	}

	if passed {
		resp.Message = "Тест амжилттай давлаа. Хичээлийн явц шинэчлэгдлээ."
		oapi.SendResp(w, resp)
		return
	}

	resp.Regenerating = true
	resp.Message = "Тестийн оноо 90%-д хүрсэнгүй."

	noteToRegenerate := *savedNote
	go regenerateNoteQuizzes(&noteToRegenerate, loggedUser.ID)

	oapi.SendRespStatus(w, http.StatusAccepted, resp)
}

func quizResultAnswers(results []quizAnswerResult) []quizman.QuizResultAnswer {
	answers := make([]quizman.QuizResultAnswer, 0, len(results))
	for _, result := range results {
		answers = append(answers, quizman.QuizResultAnswer{
			QuizID:         result.QuizID,
			Question:       result.Question,
			Options:        result.Options,
			SelectedAnswer: result.SelectedAnswer,
			CorrectAnswer:  result.CorrectAnswer,
			Correct:        result.Correct,
		})
	}
	return answers
}

func publicQuizResponses(quizzes []*quizman.Quiz) []quizResponse {
	result := make([]quizResponse, 0, len(quizzes))
	for _, quiz := range quizzes {
		if quiz == nil {
			continue
		}
		result = append(result, publicQuizResponse(quiz))
	}
	return result
}

func publicQuizResponse(quiz *quizman.Quiz) quizResponse {
	return quizResponse{
		ID:       quiz.ID,
		NoteID:   quiz.NoteID,
		Question: quiz.Question,
		Options:  quiz.Options,
	}
}

func publicGeneratedQuizResponses(quizzes []quizman.Quiz) []quizResponse {
	result := make([]quizResponse, 0, len(quizzes))
	for i := range quizzes {
		result = append(result, publicQuizResponse(&quizzes[i]))
	}
	return result
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
