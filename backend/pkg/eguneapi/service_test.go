package eguneapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeChatResponse struct {
	Choices []fakeChatChoice `json:"choices"`
}

type fakeChatChoice struct {
	Message fakeChatMessage `json:"message"`
}

type fakeChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func writeChatResponse(t *testing.T, w http.ResponseWriter, content string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fakeChatResponse{
		Choices: []fakeChatChoice{{Message: fakeChatMessage{Role: "assistant", Content: content}}},
	}); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

func TestNewServiceStoresConfig(t *testing.T) {
	svc := NewService("https://example.com", "key-1", "gpt-test")
	if svc.BaseURL != "https://example.com" || svc.APIKey != "key-1" || svc.Model != "gpt-test" {
		t.Fatalf("NewService() = %+v, want stored fields", svc)
	}
}

func TestGenerateNoteParsesStructuredResponse(t *testing.T) {
	generated := `{"note":{"title":"T","summary":"S","key_concepts":[{"concept":"C","definition":"D"}],"flash_cards":[{"question":"Q","answer":"A"}]},"quizzes":[{"question":"Qz","options":["a","b","c","d"],"correct_answer":"a"}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Fatalf("path = %q, want chat completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want Bearer test-key", got)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "raw lesson") {
			t.Fatalf("request body missing raw content: %s", string(body))
		}
		writeChatResponse(t, w, generated)
	}))
	defer server.Close()

	svc := NewService(server.URL, "test-key", "model-x")
	output, err := svc.GenerateNote("raw lesson content")
	if err != nil {
		t.Fatalf("GenerateNote() error = %v", err)
	}
	if output.Note.Title != "T" || output.Note.Summary != "S" || output.Note.RawContent != "raw lesson content" {
		t.Fatalf("output.Note = %+v, want parsed fields with raw content set", output.Note)
	}
	if len(output.Quizzes) != 1 || output.Quizzes[0].Question != "Qz" || output.Quizzes[0].CorrectAnswer != "a" {
		t.Fatalf("output.Quizzes = %+v, want one parsed quiz", output.Quizzes)
	}
}

func TestGenerateNoteErrorsOnInvalidJSONContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeChatResponse(t, w, "not-json")
	}))
	defer server.Close()

	svc := NewService(server.URL, "k", "m")
	if _, err := svc.GenerateNote("anything"); err == nil {
		t.Fatal("GenerateNote() error = nil, want unmarshal error")
	}
}

func TestGenerateNoteErrorsWhenNoteMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeChatResponse(t, w, `{"quizzes":[]}`)
	}))
	defer server.Close()

	svc := NewService(server.URL, "k", "m")
	_, err := svc.GenerateNote("anything")
	if err == nil || !strings.Contains(err.Error(), "note is missing") {
		t.Fatalf("GenerateNote() error = %v, want note missing", err)
	}
}

func TestGenerateNoteErrorsOnNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[]}`)
	}))
	defer server.Close()

	svc := NewService(server.URL, "k", "m")
	_, err := svc.GenerateNote("anything")
	if err == nil || !strings.Contains(err.Error(), "no choices") {
		t.Fatalf("GenerateNote() error = %v, want no choices", err)
	}
}

func TestGenerateNoteErrorsOnUpstreamFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := NewService(server.URL, "k", "m")
	if _, err := svc.GenerateNote("anything"); err == nil {
		t.Fatal("GenerateNote() error = nil, want chat completion error")
	}
}

func TestAnswerQuestionReturnsContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "course ctx") || !strings.Contains(string(body), "ask me") {
			t.Fatalf("request body missing context/question: %s", string(body))
		}
		writeChatResponse(t, w, "the answer")
	}))
	defer server.Close()

	svc := NewService(server.URL, "k", "m")
	answer, err := svc.AnswerQuestion("course ctx", "ask me")
	if err != nil {
		t.Fatalf("AnswerQuestion() error = %v", err)
	}
	if answer != "the answer" {
		t.Fatalf("AnswerQuestion() = %q, want the answer", answer)
	}
}

func TestAnswerQuestionErrorsOnNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[]}`)
	}))
	defer server.Close()

	svc := NewService(server.URL, "k", "m")
	_, err := svc.AnswerQuestion("ctx", "q")
	if err == nil || !strings.Contains(err.Error(), "no choices") {
		t.Fatalf("AnswerQuestion() error = %v, want no choices", err)
	}
}

func TestAnswerQuestionErrorsOnUpstreamFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := NewService(server.URL, "k", "m")
	if _, err := svc.AnswerQuestion("ctx", "q"); err == nil {
		t.Fatal("AnswerQuestion() error = nil, want chat completion error")
	}
}
