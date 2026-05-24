package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/internal/testdb"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/websocket"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/courseman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/entities"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/quizman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
	"github.com/golangcollege/sessions"
)

func TestCanAccessCourseAllowsOwnerAndAdminOnly(t *testing.T) {
	course := &courseman.Course{UserID: 7}

	tests := []struct {
		name string
		user *userman.User
		want bool
	}{
		{name: "owner", user: &userman.User{Role: userman.ROLE_USER}, want: true},
		{name: "admin", user: &userman.User{Role: userman.ROLE_ADMIN}, want: true},
		{name: "other user", user: &userman.User{Role: userman.ROLE_USER}, want: false},
	}
	tests[0].user.ID = 7
	tests[1].user.ID = 42
	tests[2].user.ID = 42

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canAccessCourse(tt.user, course); got != tt.want {
				t.Fatalf("canAccessCourse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublicQuizResponsesHideCorrectAnswersAndSkipNil(t *testing.T) {
	responses := publicQuizResponses([]*quizman.Quiz{
		{
			Model:         quizModel(10),
			NoteID:        3,
			Question:      "Question?",
			Options:       []string{"A", "B", "C", "D"},
			CorrectAnswer: "A",
		},
		nil,
	})

	if len(responses) != 1 {
		t.Fatalf("len(publicQuizResponses()) = %d, want 1", len(responses))
	}
	if responses[0].ID != 10 || responses[0].NoteID != 3 || responses[0].Question != "Question?" {
		t.Fatalf("public quiz response = %+v", responses[0])
	}

	data, err := json.Marshal(responses[0])
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if strings.Contains(string(data), "correct_answer") {
		t.Fatalf("public quiz response leaked answer: %s", string(data))
	}
}

func TestPublicGeneratedQuizResponsesHideCorrectAnswers(t *testing.T) {
	responses := publicGeneratedQuizResponses([]quizman.Quiz{
		{
			Model:         quizModel(11),
			NoteID:        4,
			Question:      "Generated?",
			Options:       []string{"A", "B", "C", "D"},
			CorrectAnswer: "B",
		},
	})

	if len(responses) != 1 {
		t.Fatalf("len(publicGeneratedQuizResponses()) = %d, want 1", len(responses))
	}

	data, err := json.Marshal(responses[0])
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if strings.Contains(string(data), "correct_answer") {
		t.Fatalf("public generated quiz response leaked answer: %s", string(data))
	}
}

func TestBuildCourseContextIncludesProcessedNoteContent(t *testing.T) {
	course := &courseman.Course{
		Notes: []*noteman.Note{
			{
				Title:      "Lesson 1",
				Summary:    "Summary text",
				RawContent: "Raw content",
				KeyConcepts: []*noteman.KeyConcept{
					{Concept: "Term", Definition: "Definition"},
				},
			},
		},
	}

	context := buildCourseContext(course)
	for _, want := range []string{"## Lesson 1", "Summary text", "- Term: Definition", "Raw content"} {
		if !strings.Contains(context, want) {
			t.Fatalf("buildCourseContext() missing %q in %q", want, context)
		}
	}
}

func TestNoteHasSourceFileAndTruncateRunes(t *testing.T) {
	if !noteHasSourceFile(&noteman.Note{FilePath: " material.pdf "}) {
		t.Fatal("noteHasSourceFile() should return true for non-blank file path")
	}
	if noteHasSourceFile(&noteman.Note{FilePath: " \n "}) {
		t.Fatal("noteHasSourceFile() should return false for blank file path")
	}

	if got := truncateRunes("abcdef", 3); got != "abc" {
		t.Fatalf("truncateRunes() = %q, want abc", got)
	}
	if got := truncateRunes("abc", 5); got != "abc" {
		t.Fatalf("truncateRunes() = %q, want abc", got)
	}
}

func TestAdminPaginationOrderAndStatusHelpers(t *testing.T) {
	page, size := pageAndSize("-1", "500")
	if page != 1 || size != 25 {
		t.Fatalf("pageAndSize() = (%d, %d), want (1, 25)", page, size)
	}

	page, size = pageAndSize("3", "50")
	if page != 3 || size != 50 {
		t.Fatalf("pageAndSize(valid) = (%d, %d), want (3, 50)", page, size)
	}

	if got := safeAdminOrder("TITLE DESC", "created_at desc"); got != "title desc" {
		t.Fatalf("safeAdminOrder() = %q, want title desc", got)
	}
	if got := safeAdminOrder("drop table users", "created_at desc"); got != "created_at desc" {
		t.Fatalf("safeAdminOrder(invalid) = %q, want fallback", got)
	}

	if got := safeNoteProcessStatus(noteman.PROCESS_STATUS_FAILED); got != noteman.PROCESS_STATUS_FAILED {
		t.Fatalf("safeNoteProcessStatus() = %q, want failed", got)
	}
	if got := safeNoteProcessStatus("unknown"); got != "" {
		t.Fatalf("safeNoteProcessStatus(invalid) = %q, want blank", got)
	}
}

func TestValidateAndSaveFileToDiskAcceptsSupportedPNG(t *testing.T) {
	fh := multipartFileHeader(t, "scan.png", minimalPNG())
	dir := t.TempDir()

	path, err := validateAndSaveFileToDisk(fh, dir)
	if err != nil {
		t.Fatalf("validateAndSaveFileToDisk() error = %v", err)
	}

	if filepath.Base(path) != "0.png" {
		t.Fatalf("saved path = %q, want file named 0.png", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Equal(data, minimalPNG()) {
		t.Fatal("saved file content does not match uploaded content")
	}
}

func TestSignupLoginAndAdminStatsRoutesWithTestDB(t *testing.T) {
	handler := setupDBBackedRouter(t)
	cookies, signedUp := signupForHandlerTest(t, handler, "route-admin@example.com")

	denied := doRequest(t, handler, http.MethodGet, "/api/admin/stats", nil, "", cookies)
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("normal user admin status = %d, want 401", denied.Code)
	}

	user, err := app.Users.GetByID(signedUp.ID)
	if err != nil {
		t.Fatalf("GetByID(signedUp) error = %v", err)
	}
	if user.PasswordHash == "" || user.PasswordHash == "StrongPass123!" {
		t.Fatal("signup should store a non-plaintext password hash")
	}
	user.Role = userman.ROLE_ADMIN
	if _, err := app.Users.Save(user); err != nil {
		t.Fatalf("promote user: %v", err)
	}

	wrongLogin := doRequest(t, handler, http.MethodPost, "/pub/auth/login", map[string]string{
		"email":    signedUp.Email,
		"password": "wrong-password",
	}, "application/json", nil)
	if wrongLogin.Code != http.StatusUnauthorized {
		t.Fatalf("wrong login status = %d, want 401", wrongLogin.Code)
	}

	correctLogin := doRequest(t, handler, http.MethodPost, "/pub/auth/login", map[string]string{
		"email":    signedUp.Email,
		"password": "StrongPass123!",
	}, "application/json", nil)
	if correctLogin.Code != http.StatusOK {
		t.Fatalf("correct login status = %d, body = %q, want 200", correctLogin.Code, correctLogin.Body.String())
	}

	allowed := doRequest(t, handler, http.MethodGet, "/api/admin/stats", nil, "", cookies)
	if allowed.Code != http.StatusOK {
		t.Fatalf("admin stats status = %d, body = %q, want 200", allowed.Code, allowed.Body.String())
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(allowed.Body).Decode(&stats); err != nil {
		t.Fatalf("decode admin stats: %v", err)
	}
	if _, ok := stats["totals"]; !ok {
		t.Fatalf("admin stats missing totals: %+v", stats)
	}
}

func TestCourseNoteQuizRoutesWithTestDB(t *testing.T) {
	handler := setupDBBackedRouter(t)
	cookies, _ := signupForHandlerTest(t, handler, "route-student@example.com")

	createCourse := doMultipartRequest(t, handler, http.MethodPost, "/api/course/", map[string]string{
		"title":       "Algorithms",
		"description": "Sorting and graphs",
	}, cookies)
	if createCourse.Code != http.StatusOK {
		t.Fatalf("create course status = %d, body = %q, want 200", createCourse.Code, createCourse.Body.String())
	}

	var course courseman.Course
	if err := json.NewDecoder(createCourse.Body).Decode(&course); err != nil {
		t.Fatalf("decode course: %v", err)
	}
	if course.ID == 0 || course.Status != courseman.STATUS_IN_PROGRESS {
		t.Fatalf("created course = %+v, want saved in-progress course", course)
	}

	updateCourse := doRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/course/%d/", course.ID), map[string]interface{}{
		"progress":  150,
		"status":    courseman.STATUS_COMPLETED,
		"is_public": true,
	}, "application/json", cookies)
	if updateCourse.Code != http.StatusOK {
		t.Fatalf("update course status = %d, body = %q, want 200", updateCourse.Code, updateCourse.Body.String())
	}

	var updatedCourse courseman.Course
	if err := json.NewDecoder(updateCourse.Body).Decode(&updatedCourse); err != nil {
		t.Fatalf("decode updated course: %v", err)
	}
	if updatedCourse.Progress != 100 || !updatedCourse.IsPublic {
		t.Fatalf("updated course = %+v, want clamped progress and public", updatedCourse)
	}

	createNote := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/course/%d/notes", course.ID), map[string]string{
		"title":   "Sorting note",
		"summary": "Sorting summary",
	}, "application/json", cookies)
	if createNote.Code != http.StatusOK {
		t.Fatalf("create note status = %d, body = %q, want 200", createNote.Code, createNote.Body.String())
	}

	var note noteman.Note
	if err := json.NewDecoder(createNote.Body).Decode(&note); err != nil {
		t.Fatalf("decode note: %v", err)
	}
	quizA, err := app.Quizzes.Save(&quizman.Quiz{
		NoteID:        note.ID,
		Question:      "First question?",
		Options:       []string{"A", "B", "C", "D"},
		CorrectAnswer: "A",
	})
	if err != nil {
		t.Fatalf("save quiz A: %v", err)
	}
	quizB, err := app.Quizzes.Save(&quizman.Quiz{
		NoteID:        note.ID,
		Question:      "Second question?",
		Options:       []string{"A", "B", "C", "D"},
		CorrectAnswer: "B",
	})
	if err != nil {
		t.Fatalf("save quiz B: %v", err)
	}

	getQuizzes := doRequest(t, handler, http.MethodGet, fmt.Sprintf("/api/notes/%d/quizzes", note.ID), nil, "", cookies)
	if getQuizzes.Code != http.StatusOK {
		t.Fatalf("get quizzes status = %d, body = %q, want 200", getQuizzes.Code, getQuizzes.Body.String())
	}
	if strings.Contains(getQuizzes.Body.String(), "correct_answer") {
		t.Fatalf("get quizzes leaked correct answer: %s", getQuizzes.Body.String())
	}

	submit := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/notes/%d/quizzes/submit", note.ID), quizSubmissionPayload{
		Answers: []quizSubmissionAnswer{
			{QuizID: quizA.ID, Answer: "A"},
			{QuizID: quizB.ID, Answer: "B"},
		},
	}, "application/json", cookies)
	if submit.Code != http.StatusOK {
		t.Fatalf("submit quiz status = %d, body = %q, want 200", submit.Code, submit.Body.String())
	}

	var result quizSubmissionResponse
	if err := json.NewDecoder(submit.Body).Decode(&result); err != nil {
		t.Fatalf("decode quiz result: %v", err)
	}
	if !result.Passed || result.Percentage != 100 || result.Score != 2 {
		t.Fatalf("quiz result = %+v, want passing 100%% score", result)
	}
	if result.Course == nil || result.Course.Progress != 100 || result.Course.Status != courseman.STATUS_COMPLETED {
		t.Fatalf("updated course in result = %+v, want completed course", result.Course)
	}
}

func TestValidateAndSaveFileToDiskRejectsUnsupportedFile(t *testing.T) {
	fh := multipartFileHeader(t, "note.txt", []byte("plain text is not an uploadable study material"))

	_, err := validateAndSaveFileToDisk(fh, t.TempDir())
	if err == nil {
		t.Fatal("validateAndSaveFileToDisk() error = nil, want unsupported file error")
	}
	if !strings.Contains(err.Error(), "unsupported file type") {
		t.Fatalf("error = %q, want unsupported file type", err.Error())
	}
}

func TestParseLimitedMultipartFormRejectsInvalidMultipartBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("not multipart"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=missing")
	rr := httptest.NewRecorder()

	if parseLimitedMultipartForm(rr, req) {
		t.Fatal("parseLimitedMultipartForm() = true, want false")
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func multipartFileHeader(t *testing.T, filename string, data []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(maxUploadFileSize); err != nil {
		t.Fatalf("ParseMultipartForm() error = %v", err)
	}
	files := req.MultipartForm.File["file"]
	if len(files) != 1 {
		t.Fatalf("multipart files = %d, want 1", len(files))
	}
	return files[0]
}

func minimalPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
}

func quizModel(id int) entities.Model {
	return entities.Model{ID: id}
}

func setupDBBackedRouter(t *testing.T) http.Handler {
	t.Helper()

	db := testdb.Open(t)
	logger := testdb.DiscardLogger()

	app.DB = db
	app.InfoLog = logger
	app.ErrorLog = logger
	app.Config.FilePath = t.TempDir()
	app.Session = sessions.New([]byte("test-session-secret"))
	app.Session.Lifetime = time.Hour
	app.Session.Secure = false
	app.Session.HttpOnly = true
	app.Session.Path = "/"
	app.Users = userman.NewService(db, logger, logger)
	app.Courses = courseman.NewService(db, logger, logger)
	app.Notes = noteman.NewService(db, logger, logger)
	app.Quizzes = quizman.NewService(db, logger, logger)
	app.CustomerWSConnections = websocket.New()
	app.CustomerWSCs = make(map[int][]*websocket.Connection)
	app.CustomerWSCsMutex = &sync.RWMutex{}

	return routes()
}

func signupForHandlerTest(t *testing.T, handler http.Handler, email string) ([]*http.Cookie, *userman.User) {
	t.Helper()

	rr := doRequest(t, handler, http.MethodPost, "/pub/auth/signup", map[string]string{
		"firstname": "Route",
		"lastname":  "Tester",
		"email":     email,
		"password":  "StrongPass123!",
	}, "application/json", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("signup status = %d, body = %q, want 200", rr.Code, rr.Body.String())
	}

	var user userman.User
	if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
		t.Fatalf("decode signup user: %v", err)
	}

	return rr.Result().Cookies(), &user
}

func doRequest(t *testing.T, handler http.Handler, method, path string, payload interface{}, contentType string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode request payload: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func doMultipartRequest(t *testing.T, handler http.Handler, method, path string, fields map[string]string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s) error = %v", key, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	req := httptest.NewRequest(method, path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}
