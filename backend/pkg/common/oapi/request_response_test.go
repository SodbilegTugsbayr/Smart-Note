package oapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testRequestPayload struct {
	Name string `json:"name"`
}

type testResponsePayload struct {
	OK     bool   `json:"ok"`
	Header string `json:"header"`
}

func TestAPIRequestDoSendsJSONHeadersAndDecodesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		if got := r.Header.Get("X-Test"); got != "token-1" {
			t.Fatalf("X-Test = %q, want token-1", got)
		}

		var payload testRequestPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Name != "lesson" {
			t.Fatalf("payload.Name = %q, want lesson", payload.Name)
		}

		SendResp(w, testResponsePayload{OK: true, Header: r.Header.Get("X-Test")})
	}))
	defer server.Close()

	result := testResponsePayload{}
	apiResp, err := (&APIRequest{
		Method:  http.MethodPost,
		URL:     server.URL,
		Headers: map[string]string{"X-Test": "token-1"},
		Data:    testRequestPayload{Name: "lesson"},
		Result:  &result,
	}).Do()
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer apiResp.CloseBody()

	if !result.OK || result.Header != "token-1" {
		t.Fatalf("decoded result = %+v, want OK with echoed header", result)
	}
}

func TestAPIRequestDoReturnsNonOKErrorDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Error-Code", "42")
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	apiResp, err := NewRequest(http.MethodGet, server.URL).Do()
	if err == nil {
		t.Fatal("Do() error = nil, want non-OK error")
	}
	defer apiResp.CloseBody()

	if apiResp.Code != 42 {
		t.Fatalf("apiResp.Code = %d, want 42", apiResp.Code)
	}
	if apiResp.ErrMessage != "bad request\n" {
		t.Fatalf("apiResp.ErrMessage = %q, want %q", apiResp.ErrMessage, "bad request\n")
	}
	if apiResp.Response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", apiResp.Response.StatusCode)
	}
}

func TestResponseHelpersWriteExpectedStatusCodes(t *testing.T) {
	tests := []struct {
		name string
		call func(http.ResponseWriter)
		want int
		body string
	}{
		{name: "client error", call: func(w http.ResponseWriter) { ClientError(w, http.StatusUnauthorized) }, want: http.StatusUnauthorized, body: "Unauthorized\n"},
		{name: "not found", call: NotFound, want: http.StatusNotFound, body: "Not Found\n"},
		{name: "forbidden", call: Forbidden, want: http.StatusForbidden, body: "Forbidden\n"},
		{name: "custom", call: func(w http.ResponseWriter) { CustomError(w, http.StatusConflict, "user_exists") }, want: http.StatusConflict, body: "user_exists\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			tt.call(rr)

			if rr.Code != tt.want {
				t.Fatalf("status = %d, want %d", rr.Code, tt.want)
			}
			if rr.Body.String() != tt.body {
				t.Fatalf("body = %q, want %q", rr.Body.String(), tt.body)
			}
		})
	}
}

func TestServerErrorWritesInternalServerError(t *testing.T) {
	original := ErrorLog
	ErrorLog = log.New(io.Discard, "", 0)
	defer func() { ErrorLog = original }()

	rr := httptest.NewRecorder()
	ServerError(rr, errors.New("database down"))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
	if rr.Body.String() != "Internal Server Error\n" {
		t.Fatalf("body = %q, want internal server error", rr.Body.String())
	}
}
