package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/eguneapi"
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

func TestCanViewCourseAllowsPublicCourse(t *testing.T) {
	course := &courseman.Course{UserID: 7, IsPublic: true}
	otherUser := &userman.User{Model: entities.Model{ID: 42}, Role: userman.ROLE_USER}

	if !canViewCourse(otherUser, course) {
		t.Fatal("canViewCourse() = false, want true for public course")
	}
	if canAccessCourse(otherUser, course) {
		t.Fatal("canAccessCourse() = true, want false for public course mutation")
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

	user.Role = userman.ROLE_ADMIN
	if _, err := app.Users.Save(user); err != nil {
		t.Fatalf("promote user: %v", err)
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

func TestMeUserCourseNoteAndAdminRoutesWithTestDB(t *testing.T) {
	handler := setupDBBackedRouter(t)
	cookies, signedUp := signupForHandlerTest(t, handler, "route-crud@example.com")

	meResp := doRequest(t, handler, http.MethodGet, "/api/me", nil, "", cookies)
	if meResp.Code != http.StatusOK {
		t.Fatalf("me status = %d, body = %q, want 200", meResp.Code, meResp.Body.String())
	}

	updateMe := doRequest(t, handler, http.MethodPost, "/api/me", map[string]string{
		"firstname":    "Updated",
		"lastname":     "Tester",
		"email":        signedUp.Email,
		"phone_number": "99112233",
	}, "application/json", cookies)
	if updateMe.Code != http.StatusOK {
		t.Fatalf("update me status = %d, body = %q, want 200", updateMe.Code, updateMe.Body.String())
	}
	var updatedUser userman.User
	if err := json.NewDecoder(updateMe.Body).Decode(&updatedUser); err != nil {
		t.Fatalf("decode updated user: %v", err)
	}
	if updatedUser.FirstName != "Updated" || updatedUser.PhoneNumber != "99112233" {
		t.Fatalf("updated user = %+v, want updated name and phone", updatedUser)
	}

	course := createCourseForRouteTest(t, handler, cookies, "Databases", "SQL indexes")
	listCourses := doRequest(t, handler, http.MethodGet, "/api/course/?keyword=Data&page=1&size=10", nil, "", cookies)
	if listCourses.Code != http.StatusOK {
		t.Fatalf("list courses status = %d, body = %q, want 200", listCourses.Code, listCourses.Body.String())
	}
	getCourseResp := doRequest(t, handler, http.MethodGet, fmt.Sprintf("/api/course/%d/", course.ID), nil, "", cookies)
	if getCourseResp.Code != http.StatusOK {
		t.Fatalf("get course status = %d, body = %q, want 200", getCourseResp.Code, getCourseResp.Body.String())
	}

	note := createNoteForRouteTest(t, handler, cookies, course.ID, "Indexes", "Initial summary")
	updateNoteResp := doRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/notes/%d/", note.ID), map[string]string{
		"title":   "Updated indexes",
		"summary": "Updated summary",
	}, "application/json", cookies)
	if updateNoteResp.Code != http.StatusOK {
		t.Fatalf("update note status = %d, body = %q, want 200", updateNoteResp.Code, updateNoteResp.Body.String())
	}

	user, err := app.Users.GetByID(signedUp.ID)
	if err != nil {
		t.Fatalf("GetByID(signedUp) error = %v", err)
	}
	user.Role = userman.ROLE_ADMIN
	if _, err := app.Users.Save(user); err != nil {
		t.Fatalf("promote user: %v", err)
	}

	adminCourses := doRequest(t, handler, http.MethodGet, "/api/admin/courses?keyword=Data&order_by=title&page=1&size=10", nil, "", cookies)
	if adminCourses.Code != http.StatusOK {
		t.Fatalf("admin courses status = %d, body = %q, want 200", adminCourses.Code, adminCourses.Body.String())
	}
	adminNotes := doRequest(t, handler, http.MethodGet, "/api/admin/notes?process_status=&order_by=created_at%20desc&page=1&size=10", nil, "", cookies)
	if adminNotes.Code != http.StatusOK {
		t.Fatalf("admin notes status = %d, body = %q, want 200", adminNotes.Code, adminNotes.Body.String())
	}
	usersList := doRequest(t, handler, http.MethodGet, "/api/users/?role=admin", nil, "", cookies)
	if usersList.Code != http.StatusOK {
		t.Fatalf("users list status = %d, body = %q, want 200", usersList.Code, usersList.Body.String())
	}
	getUserResp := doRequest(t, handler, http.MethodGet, fmt.Sprintf("/api/users/%d/", signedUp.ID), nil, "", cookies)
	if getUserResp.Code != http.StatusOK {
		t.Fatalf("get user status = %d, body = %q, want 200", getUserResp.Code, getUserResp.Body.String())
	}
	editUserResp := doRequest(t, handler, http.MethodPut, fmt.Sprintf("/api/users/%d/", signedUp.ID), map[string]string{
		"firstname":    "Edited",
		"lastname":     "Admin",
		"email":        signedUp.Email,
		"phone_number": "99887766",
	}, "application/json", cookies)
	if editUserResp.Code != http.StatusOK {
		t.Fatalf("edit user status = %d, body = %q, want 200", editUserResp.Code, editUserResp.Body.String())
	}
	deleteNoteResp := doRequest(t, handler, http.MethodDelete, fmt.Sprintf("/api/notes/%d/", note.ID), nil, "", cookies)
	if deleteNoteResp.Code != http.StatusOK {
		t.Fatalf("delete note status = %d, body = %q, want 200", deleteNoteResp.Code, deleteNoteResp.Body.String())
	}
	deleteUserResp := doRequest(t, handler, http.MethodDelete, fmt.Sprintf("/api/users/%d/", signedUp.ID), nil, "", cookies)
	if deleteUserResp.Code != http.StatusOK {
		t.Fatalf("delete user status = %d, body = %q, want 200", deleteUserResp.Code, deleteUserResp.Body.String())
	}
}

func TestMiddlewareRejectsUnauthenticatedAndUnverifiedUsersWithTestDB(t *testing.T) {
	handler := setupDBBackedRouter(t)

	unauthenticated := doRequest(t, handler, http.MethodGet, "/api/me", nil, "", nil)
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want 401", unauthenticated.Code)
	}

	cookies, signedUp := signupForHandlerTest(t, handler, "unverified@example.com")
	user, err := app.Users.GetByID(signedUp.ID)
	if err != nil {
		t.Fatalf("get signed-up user: %v", err)
	}
	user.IsVerified = false
	if _, err := app.Users.Save(user); err != nil {
		t.Fatalf("mark user unverified: %v", err)
	}

	resp := doRequest(t, handler, http.MethodGet, "/api/me", nil, "", cookies)
	if resp.Code != http.StatusPreconditionFailed {
		t.Fatalf("unverified status = %d, want 412", resp.Code)
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

func TestAIFlashcardAndChatValidationRoutesWithTestDB(t *testing.T) {
	handler := setupDBBackedRouter(t)
	cookies, _ := signupForHandlerTest(t, handler, "route-ai@example.com")
	course := createCourseForRouteTest(t, handler, cookies, "AI Course", "Generated content")
	note := createNoteForRouteTest(t, handler, cookies, course.ID, "AI note", "Existing summary")

	addFlashcardResp := doRequest(t, handler, http.MethodPost, "/api/flashcards", flashcardPayload{
		CourseID:   course.ID,
		NoteID:     note.ID,
		Term:       "Vector",
		Definition: "A numeric representation",
	}, "application/json", cookies)
	if addFlashcardResp.Code != http.StatusOK {
		t.Fatalf("add flashcard status = %d, body = %q, want 200", addFlashcardResp.Code, addFlashcardResp.Body.String())
	}

	emptyChat := doRequest(t, handler, http.MethodPost, "/api/ai/chat", chatPayload{
		CourseID: course.ID,
		Question: "",
	}, "application/json", cookies)
	if emptyChat.Code != http.StatusBadRequest {
		t.Fatalf("empty chat status = %d, want 400", emptyChat.Code)
	}

	emptyCourse := createCourseForRouteTest(t, handler, cookies, "Empty", "No notes")
	noContextChat := doRequest(t, handler, http.MethodPost, "/api/ai/chat", chatPayload{
		CourseID: emptyCourse.ID,
		Question: "What is this?",
	}, "application/json", cookies)
	if noContextChat.Code != http.StatusBadRequest {
		t.Fatalf("no-context chat status = %d, body = %q, want 400", noContextChat.Code, noContextChat.Body.String())
	}
}

func TestProcessNoteAndRegenerateQuizzesWithFakesAndTestDB(t *testing.T) {
	setupDBBackedRouter(t)
	app.OCRService = fakeOCRClient{text: " \x00raw OCR text\x00 "}
	app.EguneService = fakeEguneClient{
		output: &eguneapi.GeneratedOutput{
			Note: noteman.Note{
				Title:   "\x00Generated title",
				Summary: "Generated summary\x00",
				KeyConcepts: []*noteman.KeyConcept{
					{Concept: "\x00Concept", Definition: "Definition\x00"},
				},
				FlashCards: []*noteman.FlashCard{
					{Question: "\x00Q", Answer: "A\x00"},
				},
			},
			Quizzes: []quizman.Quiz{
				{Question: "\x00Generated quiz?", Options: []string{"A\x00", "B", "C", "D"}, CorrectAnswer: "A\x00"},
			},
		},
		answer: "chat answer",
	}

	user, course, note := createProcessingFixture(t, "/tmp/source.pdf")
	processNote(note, user.ID)

	savedNote, err := app.Notes.GetByID(note.ID)
	if err != nil {
		t.Fatalf("GetByID(processed note) error = %v", err)
	}
	if savedNote.ProcessStatus != noteman.PROCESS_STATUS_COMPLETED || savedNote.Summary != "Generated summary" || savedNote.RawContent != "raw OCR text" {
		t.Fatalf("processed note = %+v, want completed generated note", savedNote)
	}
	if savedNote.Title != "Generated title" || savedNote.KeyConcepts[0].Concept != "Concept" || savedNote.KeyConcepts[0].Definition != "Definition" {
		t.Fatalf("sanitized generated note = %+v, want no null bytes", savedNote)
	}
	if savedNote.FlashCards[0].Question != "Q" || savedNote.FlashCards[0].Answer != "A" {
		t.Fatalf("sanitized flash cards = %+v, want no null bytes", savedNote.FlashCards)
	}
	quizzes, total, err := app.Quizzes.GetAll(&quizman.Filter{NoteID: note.ID}, 1, 25)
	if err != nil {
		t.Fatalf("GetAll(quizzes) error = %v", err)
	}
	if total != 1 || len(quizzes) != 1 || quizzes[0].NoteID != note.ID {
		t.Fatalf("generated quizzes len=%d total=%d data=%+v, want one quiz for note", len(quizzes), total, quizzes)
	}
	if quizzes[0].Question != "Generated quiz?" || quizzes[0].Options[0] != "A" || quizzes[0].CorrectAnswer != "A" {
		t.Fatalf("sanitized quiz = %+v, want no null bytes", quizzes[0])
	}

	if answer, err := app.EguneService.AnswerQuestion(buildCourseContext(&courseman.Course{Notes: []*noteman.Note{savedNote}}), "Question?"); err != nil || answer != "chat answer" {
		t.Fatalf("fake AnswerQuestion() = (%q, %v), want chat answer", answer, err)
	}

	app.EguneService = fakeEguneClient{
		output: &eguneapi.GeneratedOutput{
			Note: noteman.Note{Title: "Regenerated"},
			Quizzes: []quizman.Quiz{
				{Question: "\x00Replacement?", Options: []string{"A", "\x00B", "C", "D"}, CorrectAnswer: "\x00B"},
			},
		},
	}
	regenerateNoteQuizzes(savedNote, user.ID)
	replaced, total, err := app.Quizzes.GetAll(&quizman.Filter{NoteID: savedNote.ID}, 1, 25)
	if err != nil {
		t.Fatalf("GetAll(regenerated quizzes) error = %v", err)
	}
	if total != 1 || replaced[0].Question != "Replacement?" || replaced[0].Options[1] != "B" || replaced[0].CorrectAnswer != "B" {
		t.Fatalf("regenerated quizzes len=%d data=%+v, want replacement quiz", total, replaced)
	}

	failedNote := &noteman.Note{CourseID: course.ID, Title: "Failure", FilePath: "/tmp/fail.pdf"}
	if _, err := app.Notes.Save(failedNote); err != nil {
		t.Fatalf("save failed note fixture: %v", err)
	}
	app.OCRService = fakeOCRClient{err: errors.New("ocr unavailable")}
	processNote(failedNote, user.ID)
	savedFailed, err := app.Notes.GetByID(failedNote.ID)
	if err != nil {
		t.Fatalf("GetByID(failed note) error = %v", err)
	}
	if savedFailed.ProcessStatus != noteman.PROCESS_STATUS_FAILED {
		t.Fatalf("failed note status = %q, want failed", savedFailed.ProcessStatus)
	}
}

func TestNoteProcessJobQueuesAndRetriesTransientFailures(t *testing.T) {
	setupDBBackedRouter(t)
	app.OCRService = fakeOCRClient{text: "raw OCR text"}
	app.EguneService = fakeEguneClient{err: errors.New("chat completion error: 500 Internal Server Error")}

	user, _, note := createProcessingFixture(t, "/tmp/source.pdf")
	if err := enqueueNoteProcessing(note.ID, user.ID); err != nil {
		t.Fatalf("enqueueNoteProcessing() error = %v", err)
	}

	claimed, err := claimNextNoteProcessJob()
	if err != nil {
		t.Fatalf("claimNextNoteProcessJob() error = %v", err)
	}
	if claimed == nil || claimed.Status != noteman.PROCESS_JOB_STATUS_PROCESSING || claimed.Attempts != 1 {
		t.Fatalf("claimed job = %+v, want processing attempt 1", claimed)
	}

	processNoteJob(claimed)

	var requeued noteman.NoteProcessJob
	if err := app.DB.First(&requeued, claimed.ID).Error; err != nil {
		t.Fatalf("get requeued job: %v", err)
	}
	if requeued.Status != noteman.PROCESS_JOB_STATUS_QUEUED || requeued.LastError == "" || requeued.NextRunAt == nil {
		t.Fatalf("requeued job = %+v, want queued with retry metadata", requeued)
	}

	savedNote, err := app.Notes.GetByID(note.ID)
	if err != nil {
		t.Fatalf("get queued note: %v", err)
	}
	if savedNote.ProcessStatus != noteman.PROCESS_STATUS_QUEUED {
		t.Fatalf("note status = %q, want queued", savedNote.ProcessStatus)
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

func createCourseForRouteTest(t *testing.T, handler http.Handler, cookies []*http.Cookie, title, description string) *courseman.Course {
	t.Helper()

	resp := doMultipartRequest(t, handler, http.MethodPost, "/api/course/", map[string]string{
		"title":       title,
		"description": description,
	}, cookies)
	if resp.Code != http.StatusOK {
		t.Fatalf("create course status = %d, body = %q, want 200", resp.Code, resp.Body.String())
	}

	var course courseman.Course
	if err := json.NewDecoder(resp.Body).Decode(&course); err != nil {
		t.Fatalf("decode course: %v", err)
	}
	return &course
}

func createNoteForRouteTest(t *testing.T, handler http.Handler, cookies []*http.Cookie, courseID int, title, summary string) *noteman.Note {
	t.Helper()

	resp := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/course/%d/notes", courseID), map[string]string{
		"title":   title,
		"summary": summary,
	}, "application/json", cookies)
	if resp.Code != http.StatusOK {
		t.Fatalf("create note status = %d, body = %q, want 200", resp.Code, resp.Body.String())
	}

	var note noteman.Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		t.Fatalf("decode note: %v", err)
	}
	return &note
}

func createProcessingFixture(t *testing.T, filePath string) (*userman.User, *courseman.Course, *noteman.Note) {
	t.Helper()

	user, err := app.Users.Save(&userman.User{
		FirstName:  "Processor",
		LastName:   "Tester",
		Email:      fmt.Sprintf("processor-%d@example.com", time.Now().UnixNano()),
		AuthType:   userman.AUTH_TYPE_BASIC,
		Role:       userman.ROLE_USER,
		IsVerified: true,
	})
	if err != nil {
		t.Fatalf("save processing user: %v", err)
	}
	course, err := app.Courses.Save(&courseman.Course{
		UserID:      user.ID,
		Title:       "Processing course",
		Description: "Processing course",
		Status:      courseman.STATUS_IN_PROGRESS,
	})
	if err != nil {
		t.Fatalf("save processing course: %v", err)
	}
	note, err := app.Notes.Save(&noteman.Note{
		CourseID:      course.ID,
		Title:         "Processing note",
		FilePath:      filePath,
		Status:        noteman.STATUS_IN_PROGRESS,
		ProcessStatus: noteman.PROCESS_STATUS_PROCESSING,
	})
	if err != nil {
		t.Fatalf("save processing note: %v", err)
	}
	return user, course, note
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

type fakeOCRClient struct {
	text string
	err  error
}

func (f fakeOCRClient) GetTextFromFile(filename string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.text, nil
}

type fakeEguneClient struct {
	output *eguneapi.GeneratedOutput
	answer string
	err    error
}

func (f fakeEguneClient) GenerateNote(rawContent string) (*eguneapi.GeneratedOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.output, nil
}

func (f fakeEguneClient) AnswerQuestion(courseContext, question string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.answer, nil
}

func doMultipartFileRequest(t *testing.T, handler http.Handler, method, path, fieldName, filename string, fileData []byte, fields map[string]string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s) error = %v", key, err)
		}
	}
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(fileData); err != nil {
		t.Fatalf("part.Write() error = %v", err)
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

func TestOnSocketConnectUnauthenticatedSkipsRegistration(t *testing.T) {
	app.CustomerWSCs = make(map[int][]*websocket.Connection)
	app.CustomerWSCsMutex = &sync.RWMutex{}

	req := httptest.NewRequest(http.MethodGet, "/api/ws", nil)
	ctx := context.WithValue(req.Context(), app.ContextKeyIsAuthenticated, false)
	conn := &websocket.Connection{Key: "skipped"}

	if err := onSocketConnect(req.WithContext(ctx), conn); err != nil {
		t.Fatalf("onSocketConnect(unauth) error = %v", err)
	}
	if len(app.CustomerWSCs) != 0 {
		t.Fatalf("CustomerWSCs = %+v, want empty when unauthenticated", app.CustomerWSCs)
	}
}

func TestOnSocketConnectRegistersAuthenticatedUserAndCleansUp(t *testing.T) {
	app.CustomerWSCs = make(map[int][]*websocket.Connection)
	app.CustomerWSCsMutex = &sync.RWMutex{}

	user := &userman.User{}
	user.ID = 42

	req := httptest.NewRequest(http.MethodGet, "/api/ws", nil)
	ctx := context.WithValue(req.Context(), app.ContextKeyIsAuthenticated, true)
	ctx = context.WithValue(ctx, app.ContextKeyAuthUser, user)

	conn := &websocket.Connection{Key: "conn-1"}
	if err := onSocketConnect(req.WithContext(ctx), conn); err != nil {
		t.Fatalf("onSocketConnect(auth) error = %v", err)
	}
	if len(app.CustomerWSCs[42]) != 1 || app.CustomerWSCs[42][0].Key != "conn-1" {
		t.Fatalf("CustomerWSCs[42] = %+v, want one registered connection", app.CustomerWSCs[42])
	}

	conn.OnClose()
	if len(app.CustomerWSCs[42]) != 0 {
		t.Fatalf("CustomerWSCs[42] after close = %+v, want empty", app.CustomerWSCs[42])
	}
}

func TestPingAndLogoutRoutes(t *testing.T) {
	handler := setupDBBackedRouter(t)

	pingResp := doRequest(t, handler, http.MethodGet, "/ping", nil, "", nil)
	if pingResp.Code != http.StatusOK {
		t.Fatalf("ping status = %d, want 200", pingResp.Code)
	}
	if !strings.Contains(pingResp.Body.String(), "OK") {
		t.Fatalf("ping body = %q, want OK", pingResp.Body.String())
	}

	cookies, _ := signupForHandlerTest(t, handler, "logout-user@example.com")
	logoutResp := doRequest(t, handler, http.MethodPost, "/api/logout", nil, "", cookies)
	if logoutResp.Code != http.StatusOK {
		t.Fatalf("logout status = %d, body = %q, want 200", logoutResp.Code, logoutResp.Body.String())
	}
}

func TestUploadFileRoute(t *testing.T) {
	handler := setupDBBackedRouter(t)
	cookies, _ := signupForHandlerTest(t, handler, "upload-user@example.com")

	missing := doMultipartRequest(t, handler, http.MethodPost, "/api/upload", map[string]string{"_": "x"}, cookies)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing file status = %d, body = %q, want 400", missing.Code, missing.Body.String())
	}

	resp := doMultipartFileRequest(t, handler, http.MethodPost, "/api/upload", "file", "pic.png", minimalPNG(), nil, cookies)
	if resp.Code != http.StatusOK {
		t.Fatalf("upload status = %d, body = %q, want 200", resp.Code, resp.Body.String())
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode upload body: %v", err)
	}
	if !strings.HasPrefix(body["file_url"], "/pub/images/attachments/") || !strings.HasSuffix(body["file_url"], ".png") {
		t.Fatalf("file_url = %q, want /pub/images/attachments/<uuid>/...png", body["file_url"])
	}

	reject := doMultipartFileRequest(t, handler, http.MethodPost, "/api/upload", "file", "note.txt", []byte("just text content"), nil, cookies)
	if reject.Code != http.StatusBadRequest {
		t.Fatalf("unsupported file status = %d, want 400", reject.Code)
	}
}

func TestDeleteCourseRoute(t *testing.T) {
	handler := setupDBBackedRouter(t)
	ownerCookies, _ := signupForHandlerTest(t, handler, "course-owner@example.com")
	intruderCookies, _ := signupForHandlerTest(t, handler, "course-intruder@example.com")

	course := createCourseForRouteTest(t, handler, ownerCookies, "ToDelete", "Removable")

	forbidden := doRequest(t, handler, http.MethodDelete, fmt.Sprintf("/api/course/%d/", course.ID), nil, "", intruderCookies)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("intruder delete status = %d, want 403", forbidden.Code)
	}

	deleted := doRequest(t, handler, http.MethodDelete, fmt.Sprintf("/api/course/%d/", course.ID), nil, "", ownerCookies)
	if deleted.Code != http.StatusOK {
		t.Fatalf("owner delete status = %d, body = %q, want 200", deleted.Code, deleted.Body.String())
	}
	if _, err := app.Courses.GetByID(course.ID); err == nil {
		t.Fatal("course still retrievable after delete")
	}
}

func TestUploadNoteFileRoute(t *testing.T) {
	handler := setupDBBackedRouter(t)
	app.OCRService = fakeOCRClient{text: "raw OCR text"}
	app.EguneService = fakeEguneClient{output: &eguneapi.GeneratedOutput{
		Note: noteman.Note{Title: "Generated", Summary: "Summary"},
	}}

	cookies, _ := signupForHandlerTest(t, handler, "note-upload@example.com")
	course := createCourseForRouteTest(t, handler, cookies, "Bio", "Biology")
	note := createNoteForRouteTest(t, handler, cookies, course.ID, "Cells", "Initial")

	resp := doMultipartFileRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/notes/%d/file", note.ID), "file", "cells.png", minimalPNG(), nil, cookies)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("upload note file status = %d, body = %q, want 202", resp.Code, resp.Body.String())
	}

	again := doMultipartFileRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/notes/%d/file", note.ID), "file", "more.png", minimalPNG(), nil, cookies)
	if again.Code != http.StatusBadRequest {
		t.Fatalf("re-upload status = %d, want 400 (already has file)", again.Code)
	}
}

func TestReprocessAdminNoteRoute(t *testing.T) {
	handler := setupDBBackedRouter(t)
	app.OCRService = fakeOCRClient{text: "raw"}
	app.EguneService = fakeEguneClient{output: &eguneapi.GeneratedOutput{
		Note: noteman.Note{Title: "Re", Summary: "Re-summary"},
	}}

	cookies, signedUp := signupForHandlerTest(t, handler, "reprocess-admin@example.com")
	user, err := app.Users.GetByID(signedUp.ID)
	if err != nil {
		t.Fatalf("get signed user: %v", err)
	}
	user.Role = userman.ROLE_ADMIN
	if _, err := app.Users.Save(user); err != nil {
		t.Fatalf("promote admin: %v", err)
	}

	tmp, err := os.CreateTemp(t.TempDir(), "src-*.pdf")
	if err != nil {
		t.Fatalf("create tmp file: %v", err)
	}
	_ = tmp.Close()

	course, err := app.Courses.Save(&courseman.Course{UserID: user.ID, Title: "Admin course"})
	if err != nil {
		t.Fatalf("save course: %v", err)
	}
	noteWithFile, err := app.Notes.Save(&noteman.Note{
		CourseID:      course.ID,
		Title:         "With file",
		FilePath:      tmp.Name(),
		ProcessStatus: noteman.PROCESS_STATUS_COMPLETED,
		Summary:       "old summary",
	})
	if err != nil {
		t.Fatalf("save note: %v", err)
	}

	reprocessResp := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/admin/notes/%d/reprocess", noteWithFile.ID), nil, "", cookies)
	if reprocessResp.Code != http.StatusOK {
		t.Fatalf("reprocess status = %d, body = %q, want 200", reprocessResp.Code, reprocessResp.Body.String())
	}

	noFileNote, err := app.Notes.Save(&noteman.Note{CourseID: course.ID, Title: "No file"})
	if err != nil {
		t.Fatalf("save no-file note: %v", err)
	}
	noFileResp := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/admin/notes/%d/reprocess", noFileNote.ID), nil, "", cookies)
	if noFileResp.Code != http.StatusBadRequest {
		t.Fatalf("reprocess(no file) status = %d, want 400", noFileResp.Code)
	}

	processingNote, err := app.Notes.Save(&noteman.Note{
		CourseID:      course.ID,
		Title:         "Processing",
		FilePath:      tmp.Name(),
		ProcessStatus: noteman.PROCESS_STATUS_PROCESSING,
	})
	if err != nil {
		t.Fatalf("save processing note: %v", err)
	}
	busyResp := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/admin/notes/%d/reprocess", processingNote.ID), nil, "", cookies)
	if busyResp.Code != http.StatusBadRequest {
		t.Fatalf("reprocess(processing) status = %d, want 400", busyResp.Code)
	}
}

func TestProcessCourseNotesAndGenerateFlashcardsRoutes(t *testing.T) {
	handler := setupDBBackedRouter(t)
	app.OCRService = fakeOCRClient{text: "raw OCR"}
	app.EguneService = fakeEguneClient{output: &eguneapi.GeneratedOutput{
		Note: noteman.Note{Title: "Gen", Summary: "Generated summary"},
		Quizzes: []quizman.Quiz{
			{Question: "Q?", Options: []string{"A", "B"}, CorrectAnswer: "A"},
		},
	}}

	cookies, _ := signupForHandlerTest(t, handler, "ai-batch@example.com")
	course := createCourseForRouteTest(t, handler, cookies, "Batch", "Process all")

	tmp, err := os.CreateTemp(t.TempDir(), "batch-*.pdf")
	if err != nil {
		t.Fatalf("create tmp: %v", err)
	}
	_ = tmp.Close()
	if _, err := app.Notes.Save(&noteman.Note{CourseID: course.ID, Title: "Needs processing", FilePath: tmp.Name()}); err != nil {
		t.Fatalf("save note: %v", err)
	}

	processResp := doRequest(t, handler, http.MethodPost, "/api/ai/process-notes", courseIDPayload{CourseID: course.ID}, "application/json", cookies)
	if processResp.Code != http.StatusOK {
		t.Fatalf("process-notes status = %d, body = %q, want 200", processResp.Code, processResp.Body.String())
	}

	flashResp := doRequest(t, handler, http.MethodPost, "/api/ai/generate-flashcards", courseIDPayload{CourseID: course.ID}, "application/json", cookies)
	if flashResp.Code != http.StatusOK {
		t.Fatalf("generate-flashcards status = %d, body = %q, want 200", flashResp.Code, flashResp.Body.String())
	}

	badJSON := doRequest(t, handler, http.MethodPost, "/api/ai/process-notes", "not-json", "application/json", cookies)
	if badJSON.Code != http.StatusBadRequest {
		t.Fatalf("bad-json process-notes status = %d, want 400", badJSON.Code)
	}
	notFound := doRequest(t, handler, http.MethodPost, "/api/ai/process-notes", courseIDPayload{CourseID: 999999}, "application/json", cookies)
	if notFound.Code != http.StatusNotFound {
		t.Fatalf("missing-course process-notes status = %d, want 404", notFound.Code)
	}

	badFlash := doRequest(t, handler, http.MethodPost, "/api/ai/generate-flashcards", "not-json", "application/json", cookies)
	if badFlash.Code != http.StatusBadRequest {
		t.Fatalf("bad-json generate-flashcards status = %d, want 400", badFlash.Code)
	}
}

func TestSaveCourseValidationAndWithBookFile(t *testing.T) {
	handler := setupDBBackedRouter(t)
	app.OCRService = fakeOCRClient{text: "raw"}
	app.EguneService = fakeEguneClient{output: &eguneapi.GeneratedOutput{
		Note: noteman.Note{Title: "Generated", Summary: "Gen"},
	}}

	cookies, _ := signupForHandlerTest(t, handler, "course-saver@example.com")

	noTitle := doMultipartRequest(t, handler, http.MethodPost, "/api/course/", map[string]string{"description": "X"}, cookies)
	if noTitle.Code != http.StatusBadRequest {
		t.Fatalf("missing title status = %d, want 400", noTitle.Code)
	}

	noDesc := doMultipartRequest(t, handler, http.MethodPost, "/api/course/", map[string]string{"title": "X"}, cookies)
	if noDesc.Code != http.StatusBadRequest {
		t.Fatalf("missing description status = %d, want 400", noDesc.Code)
	}

	badFile := doMultipartFileRequest(t, handler, http.MethodPost, "/api/course/", "file", "bad.txt", []byte("plain text"), map[string]string{
		"title":       "With bad file",
		"description": "Some description",
	}, cookies)
	if badFile.Code != http.StatusBadRequest {
		t.Fatalf("unsupported file status = %d, want 400", badFile.Code)
	}

	withBook := doMultipartFileRequest(t, handler, http.MethodPost, "/api/course/", "file", "book.png", minimalPNG(), map[string]string{
		"title":       "WithBook",
		"description": "Has a book",
		"icon":        "Book",
		"sections":    `[{"section_name":"Intro","start_page":1,"end_page":5}]`,
	}, cookies)
	if withBook.Code != http.StatusOK {
		t.Fatalf("with-book status = %d, body = %q, want 200", withBook.Code, withBook.Body.String())
	}

	badSections := doMultipartFileRequest(t, handler, http.MethodPost, "/api/course/", "file", "book.png", minimalPNG(), map[string]string{
		"title":       "BadSections",
		"description": "Sections malformed",
		"sections":    "not-json",
	}, cookies)
	if badSections.Code != http.StatusBadRequest {
		t.Fatalf("bad sections status = %d, want 400", badSections.Code)
	}
}

func TestUploadCourseBookForExistingCourse(t *testing.T) {
	handler := setupDBBackedRouter(t)
	cookies, _ := signupForHandlerTest(t, handler, "course-book-uploader@example.com")
	course := createCourseForRouteTest(t, handler, cookies, "Blank", "Blank course")

	upload := doMultipartFileRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/course/%d/book", course.ID), "file", "book.png", minimalPNG(), map[string]string{
		"sections": `[{"section_name":"Intro","start_page":1,"end_page":5}]`,
	}, cookies)
	if upload.Code != http.StatusOK {
		t.Fatalf("upload book status = %d, body = %q, want 200", upload.Code, upload.Body.String())
	}

	var updated courseman.Course
	if err := json.NewDecoder(upload.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated course: %v", err)
	}
	if !updated.HasBook || len(updated.Notes) != 1 || updated.Notes[0].ProcessStatus != noteman.PROCESS_STATUS_QUEUED {
		t.Fatalf("updated course = %+v, want one queued book note", updated)
	}

	again := doMultipartFileRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/course/%d/book", course.ID), "file", "book.png", minimalPNG(), nil, cookies)
	if again.Code != http.StatusBadRequest {
		t.Fatalf("upload existing book status = %d, want 400", again.Code)
	}
}

func TestUpdateCourseAllFieldBranches(t *testing.T) {
	handler := setupDBBackedRouter(t)
	cookies, _ := signupForHandlerTest(t, handler, "course-updater@example.com")
	course := createCourseForRouteTest(t, handler, cookies, "Original", "Original")

	badJSON := doRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/course/%d/", course.ID), "not-json", "application/json", cookies)
	if badJSON.Code != http.StatusBadRequest {
		t.Fatalf("bad json status = %d, want 400", badJSON.Code)
	}

	emptyTitle := ""
	emptyTitleResp := doRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/course/%d/", course.ID), updateCoursePayload{Title: &emptyTitle}, "application/json", cookies)
	if emptyTitleResp.Code != http.StatusBadRequest {
		t.Fatalf("empty title status = %d, want 400", emptyTitleResp.Code)
	}

	badStatus := "invalid"
	badStatusResp := doRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/course/%d/", course.ID), updateCoursePayload{Status: &badStatus}, "application/json", cookies)
	if badStatusResp.Code != http.StatusBadRequest {
		t.Fatalf("bad status status = %d, want 400", badStatusResp.Code)
	}

	newTitle := "Renamed"
	newDesc := "New desc"
	newStatus := courseman.STATUS_COMPLETED
	overProgress := 150
	isPublic := true
	icon := "Star"
	full := doRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/course/%d/", course.ID), updateCoursePayload{
		Title:       &newTitle,
		Description: &newDesc,
		Status:      &newStatus,
		Progress:    &overProgress,
		IsPublic:    &isPublic,
		Icon:        &icon,
	}, "application/json", cookies)
	if full.Code != http.StatusOK {
		t.Fatalf("full update status = %d, body = %q, want 200", full.Code, full.Body.String())
	}

	negProgress := -10
	negResp := doRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/course/%d/", course.ID), updateCoursePayload{Progress: &negProgress}, "application/json", cookies)
	if negResp.Code != http.StatusOK {
		t.Fatalf("negative progress status = %d, want 200 (clamped)", negResp.Code)
	}

	intruder, _ := signupForHandlerTest(t, handler, "course-intruder-update@example.com")
	publicGet := doRequest(t, handler, http.MethodGet, fmt.Sprintf("/api/course/%d/", course.ID), nil, "", intruder)
	if publicGet.Code != http.StatusOK {
		t.Fatalf("intruder public get status = %d, body = %q, want 200", publicGet.Code, publicGet.Body.String())
	}

	forbidden := doRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/course/%d/", course.ID), updateCoursePayload{Title: &newTitle}, "application/json", intruder)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("intruder update status = %d, want 403", forbidden.Code)
	}
}

func TestUpdateNoteValidation(t *testing.T) {
	handler := setupDBBackedRouter(t)
	cookies, _ := signupForHandlerTest(t, handler, "note-updater@example.com")
	course := createCourseForRouteTest(t, handler, cookies, "Course", "Desc")
	note := createNoteForRouteTest(t, handler, cookies, course.ID, "Original", "Initial summary")

	badJSON := doRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/notes/%d/", note.ID), "not-json", "application/json", cookies)
	if badJSON.Code != http.StatusBadRequest {
		t.Fatalf("bad json status = %d, want 400", badJSON.Code)
	}

	emptyTitle := ""
	emptyResp := doRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/notes/%d/", note.ID), updateNotePayload{Title: &emptyTitle}, "application/json", cookies)
	if emptyResp.Code != http.StatusBadRequest {
		t.Fatalf("empty title status = %d, want 400", emptyResp.Code)
	}

	intruder, _ := signupForHandlerTest(t, handler, "note-intruder-update@example.com")
	newTitle := "Hijacked"
	forbidden := doRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/notes/%d/", note.ID), updateNotePayload{Title: &newTitle}, "application/json", intruder)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("intruder update status = %d, want 403", forbidden.Code)
	}

	deleteIntruderResp := doRequest(t, handler, http.MethodDelete, fmt.Sprintf("/api/notes/%d/", note.ID), nil, "", intruder)
	if deleteIntruderResp.Code != http.StatusForbidden {
		t.Fatalf("intruder delete status = %d, want 403", deleteIntruderResp.Code)
	}
}

func TestSubmitNoteQuizValidationAndScoring(t *testing.T) {
	handler := setupDBBackedRouter(t)
	app.OCRService = fakeOCRClient{text: "raw"}
	app.EguneService = fakeEguneClient{output: &eguneapi.GeneratedOutput{
		Note: noteman.Note{Title: "Re", Summary: "Re"},
		Quizzes: []quizman.Quiz{
			{Question: "Replacement?", Options: []string{"A", "B"}, CorrectAnswer: "B"},
		},
	}}

	cookies, _ := signupForHandlerTest(t, handler, "quiz-taker@example.com")
	course := createCourseForRouteTest(t, handler, cookies, "Quizzy", "Quiz course")
	note := createNoteForRouteTest(t, handler, cookies, course.ID, "Topic", "")

	noQuizzes := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/notes/%d/quizzes/submit", note.ID), quizSubmissionPayload{
		Answers: []quizSubmissionAnswer{{QuizID: 1, Answer: "A"}},
	}, "application/json", cookies)
	if noQuizzes.Code != http.StatusBadRequest {
		t.Fatalf("no-quizzes submit status = %d, want 400", noQuizzes.Code)
	}

	q1, err := app.Quizzes.Save(&quizman.Quiz{NoteID: note.ID, Question: "Q1", Options: []string{"A", "B"}, CorrectAnswer: "A"})
	if err != nil {
		t.Fatalf("save quiz q1: %v", err)
	}
	q2, err := app.Quizzes.Save(&quizman.Quiz{NoteID: note.ID, Question: "Q2", Options: []string{"A", "B"}, CorrectAnswer: "B"})
	if err != nil {
		t.Fatalf("save quiz q2: %v", err)
	}

	badJSON := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/notes/%d/quizzes/submit", note.ID), "not-json", "application/json", cookies)
	if badJSON.Code != http.StatusBadRequest {
		t.Fatalf("bad json status = %d, want 400", badJSON.Code)
	}

	noAnswers := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/notes/%d/quizzes/submit", note.ID), quizSubmissionPayload{}, "application/json", cookies)
	if noAnswers.Code != http.StatusBadRequest {
		t.Fatalf("no answers status = %d, want 400", noAnswers.Code)
	}

	invalidID := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/notes/%d/quizzes/submit", note.ID), quizSubmissionPayload{
		Answers: []quizSubmissionAnswer{{QuizID: 0, Answer: "A"}},
	}, "application/json", cookies)
	if invalidID.Code != http.StatusBadRequest {
		t.Fatalf("invalid quiz id status = %d, want 400", invalidID.Code)
	}

	wrongNote := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/notes/%d/quizzes/submit", note.ID), quizSubmissionPayload{
		Answers: []quizSubmissionAnswer{{QuizID: 999999, Answer: "A"}},
	}, "application/json", cookies)
	if wrongNote.Code != http.StatusBadRequest {
		t.Fatalf("wrong-note quiz status = %d, want 400", wrongNote.Code)
	}

	pass := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/notes/%d/quizzes/submit", note.ID), quizSubmissionPayload{
		Answers: []quizSubmissionAnswer{{QuizID: q1.ID, Answer: "A"}, {QuizID: q2.ID, Answer: "B"}},
	}, "application/json", cookies)
	if pass.Code != http.StatusOK {
		t.Fatalf("pass submit status = %d, body = %q, want 200", pass.Code, pass.Body.String())
	}

	getQuizzes := doRequest(t, handler, http.MethodGet, fmt.Sprintf("/api/notes/%d/quizzes", note.ID), nil, "", cookies)
	if getQuizzes.Code != http.StatusOK {
		t.Fatalf("get quizzes status = %d, want 200", getQuizzes.Code)
	}
	if !strings.Contains(getQuizzes.Body.String(), `"results"`) || !strings.Contains(getQuizzes.Body.String(), `"correct_answer"`) {
		t.Fatalf("get quizzes body = %q, want saved quiz result details", getQuizzes.Body.String())
	}

	fail := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/notes/%d/quizzes/submit", note.ID), quizSubmissionPayload{
		Answers: []quizSubmissionAnswer{{QuizID: q1.ID, Answer: "B"}, {QuizID: q2.ID, Answer: "A"}},
	}, "application/json", cookies)
	if fail.Code != http.StatusAccepted {
		t.Fatalf("failing submit status = %d, body = %q, want 202", fail.Code, fail.Body.String())
	}
}

func TestLoginEdgeCases(t *testing.T) {
	handler := setupDBBackedRouter(t)

	badJSON := doRequest(t, handler, http.MethodPost, "/pub/auth/login", "not-json", "application/json", nil)
	if badJSON.Code != http.StatusBadRequest {
		t.Fatalf("login bad json status = %d, want 400", badJSON.Code)
	}

	missingFields := doRequest(t, handler, http.MethodPost, "/pub/auth/login", map[string]string{"email": "", "password": ""}, "application/json", nil)
	if missingFields.Code != http.StatusBadRequest {
		t.Fatalf("login missing fields status = %d, want 400", missingFields.Code)
	}

	unknown := doRequest(t, handler, http.MethodPost, "/pub/auth/login", map[string]string{
		"email":    "ghost@example.com",
		"password": "AnyPass123!",
	}, "application/json", nil)
	if unknown.Code != http.StatusUnauthorized {
		t.Fatalf("login unknown email status = %d, want 401", unknown.Code)
	}
}

func TestCreateCourseNoteForbiddenAndDefaultTitle(t *testing.T) {
	handler := setupDBBackedRouter(t)
	ownerCookies, _ := signupForHandlerTest(t, handler, "note-creator@example.com")
	intruderCookies, _ := signupForHandlerTest(t, handler, "note-intruder-create@example.com")
	course := createCourseForRouteTest(t, handler, ownerCookies, "Notes course", "Has notes")

	forbidden := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/course/%d/notes", course.ID), map[string]string{"title": "Hi"}, "application/json", intruderCookies)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("intruder note create status = %d, want 403", forbidden.Code)
	}

	emptyBody := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/course/%d/notes", course.ID), nil, "", ownerCookies)
	if emptyBody.Code != http.StatusOK {
		t.Fatalf("empty body note create status = %d, body = %q, want 200 (default title)", emptyBody.Code, emptyBody.Body.String())
	}
	var note noteman.Note
	if err := json.NewDecoder(emptyBody.Body).Decode(&note); err != nil {
		t.Fatalf("decode default note: %v", err)
	}
	if strings.TrimSpace(note.Title) == "" {
		t.Fatal("default note title is empty, want fallback")
	}
}

func TestSignupValidatorRejectsBadEmail(t *testing.T) {
	handler := setupDBBackedRouter(t)

	rr := doRequest(t, handler, http.MethodPost, "/pub/auth/signup", map[string]string{
		"firstname": "X",
		"lastname":  "Y",
		"email":     "not-an-email",
		"password":  "StrongPass123!",
	}, "application/json", nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("bad email signup status = %d, want 400", rr.Code)
	}

	missing := doRequest(t, handler, http.MethodPost, "/pub/auth/signup", map[string]string{
		"firstname": "X",
		"lastname":  "Y",
		"email":     "",
		"password":  "StrongPass123!",
	}, "application/json", nil)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing email signup status = %d, want 400", missing.Code)
	}
}

func TestSetChosenMiddlewareNotFound(t *testing.T) {
	handler := setupDBBackedRouter(t)
	cookies, signedUp := signupForHandlerTest(t, handler, "middleware-404@example.com")
	user, err := app.Users.GetByID(signedUp.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	user.Role = userman.ROLE_ADMIN
	if _, err := app.Users.Save(user); err != nil {
		t.Fatalf("promote admin: %v", err)
	}

	missingCourse := doRequest(t, handler, http.MethodGet, "/api/course/999999/", nil, "", cookies)
	if missingCourse.Code != http.StatusNotFound {
		t.Fatalf("missing course status = %d, want 404", missingCourse.Code)
	}
	missingNote := doRequest(t, handler, http.MethodPatch, "/api/notes/999999/", map[string]string{"title": "X"}, "application/json", cookies)
	if missingNote.Code != http.StatusNotFound {
		t.Fatalf("missing note status = %d, want 404", missingNote.Code)
	}
	missingUser := doRequest(t, handler, http.MethodGet, "/api/users/999999/", nil, "", cookies)
	if missingUser.Code != http.StatusNotFound {
		t.Fatalf("missing user status = %d, want 404", missingUser.Code)
	}
}

func TestEditUserAndUpdateUserInfoValidation(t *testing.T) {
	handler := setupDBBackedRouter(t)
	cookies, signedUp := signupForHandlerTest(t, handler, "user-edit@example.com")
	user, err := app.Users.GetByID(signedUp.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	user.Role = userman.ROLE_ADMIN
	if _, err := app.Users.Save(user); err != nil {
		t.Fatalf("promote admin: %v", err)
	}

	badJSONEdit := doRequest(t, handler, http.MethodPut, fmt.Sprintf("/api/users/%d/", signedUp.ID), "not-json", "application/json", cookies)
	if badJSONEdit.Code != http.StatusBadRequest {
		t.Fatalf("edit bad json status = %d, want 400", badJSONEdit.Code)
	}

	noEmail := doRequest(t, handler, http.MethodPut, fmt.Sprintf("/api/users/%d/", signedUp.ID), map[string]string{
		"firstname": "X",
		"lastname":  "Y",
		"email":     "",
	}, "application/json", cookies)
	if noEmail.Code != http.StatusBadRequest {
		t.Fatalf("edit no email status = %d, want 400", noEmail.Code)
	}

	badJSONMe := doRequest(t, handler, http.MethodPost, "/api/me", "not-json", "application/json", cookies)
	if badJSONMe.Code != http.StatusBadRequest {
		t.Fatalf("me bad json status = %d, want 400", badJSONMe.Code)
	}

	noEmailMe := doRequest(t, handler, http.MethodPost, "/api/me", map[string]string{
		"firstname": "X",
		"lastname":  "Y",
		"email":     "",
	}, "application/json", cookies)
	if noEmailMe.Code != http.StatusBadRequest {
		t.Fatalf("me no email status = %d, want 400", noEmailMe.Code)
	}
}

func TestAskCourseChatValidationAndForbidden(t *testing.T) {
	handler := setupDBBackedRouter(t)
	app.EguneService = fakeEguneClient{answer: "ok"}

	ownerCookies, _ := signupForHandlerTest(t, handler, "chat-owner@example.com")
	intruderCookies, _ := signupForHandlerTest(t, handler, "chat-intruder@example.com")
	course := createCourseForRouteTest(t, handler, ownerCookies, "Chat course", "Chat course description")

	badJSON := doRequest(t, handler, http.MethodPost, "/api/ai/chat", "not-json", "application/json", ownerCookies)
	if badJSON.Code != http.StatusBadRequest {
		t.Fatalf("chat bad json status = %d, want 400", badJSON.Code)
	}

	missingCourse := doRequest(t, handler, http.MethodPost, "/api/ai/chat", chatPayload{CourseID: 999999, Question: "Q?"}, "application/json", ownerCookies)
	if missingCourse.Code != http.StatusNotFound {
		t.Fatalf("chat missing course status = %d, want 404", missingCourse.Code)
	}

	forbidden := doRequest(t, handler, http.MethodPost, "/api/ai/chat", chatPayload{CourseID: course.ID, Question: "Q?"}, "application/json", intruderCookies)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("chat forbidden status = %d, want 403", forbidden.Code)
	}
}

func TestAddFlashcardValidation(t *testing.T) {
	handler := setupDBBackedRouter(t)
	cookies, _ := signupForHandlerTest(t, handler, "flash-validator@example.com")
	emptyCourse := createCourseForRouteTest(t, handler, cookies, "Empty", "No notes")

	badJSON := doRequest(t, handler, http.MethodPost, "/api/flashcards", "not-json", "application/json", cookies)
	if badJSON.Code != http.StatusBadRequest {
		t.Fatalf("flashcard bad json status = %d, want 400", badJSON.Code)
	}

	emptyTerm := doRequest(t, handler, http.MethodPost, "/api/flashcards", flashcardPayload{CourseID: emptyCourse.ID, Term: "", Definition: "def"}, "application/json", cookies)
	if emptyTerm.Code != http.StatusBadRequest {
		t.Fatalf("flashcard empty term status = %d, want 400", emptyTerm.Code)
	}

	missingCourse := doRequest(t, handler, http.MethodPost, "/api/flashcards", flashcardPayload{CourseID: 999999, Term: "T", Definition: "D"}, "application/json", cookies)
	if missingCourse.Code != http.StatusNotFound {
		t.Fatalf("flashcard missing course status = %d, want 404", missingCourse.Code)
	}

	noNotes := doRequest(t, handler, http.MethodPost, "/api/flashcards", flashcardPayload{CourseID: emptyCourse.ID, Term: "T", Definition: "D"}, "application/json", cookies)
	if noNotes.Code != http.StatusBadRequest {
		t.Fatalf("flashcard no notes status = %d, want 400", noNotes.Code)
	}
}

func TestGetNoteQuizzesForbidden(t *testing.T) {
	handler := setupDBBackedRouter(t)
	ownerCookies, _ := signupForHandlerTest(t, handler, "quiz-owner@example.com")
	intruderCookies, _ := signupForHandlerTest(t, handler, "quiz-intruder@example.com")
	course := createCourseForRouteTest(t, handler, ownerCookies, "Quiz course", "Has quizzes")
	note := createNoteForRouteTest(t, handler, ownerCookies, course.ID, "Topic", "Summary")

	forbidden := doRequest(t, handler, http.MethodGet, fmt.Sprintf("/api/notes/%d/quizzes", note.ID), nil, "", intruderCookies)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("intruder get quizzes status = %d, want 403", forbidden.Code)
	}

	forbiddenSubmit := doRequest(t, handler, http.MethodPost, fmt.Sprintf("/api/notes/%d/quizzes/submit", note.ID), quizSubmissionPayload{
		Answers: []quizSubmissionAnswer{{QuizID: 1, Answer: "A"}},
	}, "application/json", intruderCookies)
	if forbiddenSubmit.Code != http.StatusForbidden {
		t.Fatalf("intruder submit quizzes status = %d, want 403", forbiddenSubmit.Code)
	}
}

func TestSignupRejectsInvalidPayloads(t *testing.T) {
	handler := setupDBBackedRouter(t)

	badJSON := doRequest(t, handler, http.MethodPost, "/pub/auth/signup", "not-json", "application/json", nil)
	if badJSON.Code != http.StatusBadRequest {
		t.Fatalf("signup bad json status = %d, want 400", badJSON.Code)
	}

	weakPass := doRequest(t, handler, http.MethodPost, "/pub/auth/signup", map[string]string{
		"firstname": "X",
		"lastname":  "Y",
		"email":     "x@example.com",
		"password":  "short",
	}, "application/json", nil)
	if weakPass.Code != http.StatusBadRequest {
		t.Fatalf("signup weak password status = %d, want 400", weakPass.Code)
	}

	_, _ = signupForHandlerTest(t, handler, "duplicate@example.com")
	dup := doRequest(t, handler, http.MethodPost, "/pub/auth/signup", map[string]string{
		"firstname": "Dup",
		"lastname":  "User",
		"email":     "duplicate@example.com",
		"password":  "StrongPass123!",
	}, "application/json", nil)
	if dup.Code == http.StatusOK {
		t.Fatalf("duplicate signup status = 200, want error")
	}
}
