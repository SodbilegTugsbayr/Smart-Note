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

func TestNewResponseStoresData(t *testing.T) {
	payload := testResponsePayload{OK: true, Header: "x"}
	resp := NewResponse(payload)
	if resp == nil {
		t.Fatal("NewResponse returned nil")
	}
	got, ok := resp.Data.(testResponsePayload)
	if !ok {
		t.Fatalf("Data type = %T, want testResponsePayload", resp.Data)
	}
	if got != payload {
		t.Fatalf("Data = %+v, want %+v", got, payload)
	}
}

func TestSendRespWritesJSONWithDefaultStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	SendResp(rr, testResponsePayload{OK: true, Header: "h"})

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var decoded testResponsePayload
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !decoded.OK || decoded.Header != "h" {
		t.Fatalf("decoded = %+v, want OK with header h", decoded)
	}
}

func TestSendRespStatusWritesStatusAndJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	SendRespStatus(rr, http.StatusCreated, testResponsePayload{OK: true, Header: "created"})

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var decoded testResponsePayload
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if decoded.Header != "created" {
		t.Fatalf("decoded.Header = %q, want created", decoded.Header)
	}
}

func TestSendFormErrorWritesBadRequest(t *testing.T) {
	rr := httptest.NewRecorder()
	err := SendFormError(rr, map[string]string{"field": "required"})
	if err != nil {
		t.Fatalf("SendFormError error = %v", err)
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var decoded map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if decoded["field"] != "required" {
		t.Fatalf("decoded = %v, want field=required", decoded)
	}
}

func TestRedirectWritesSeeOtherAndURL(t *testing.T) {
	rr := httptest.NewRecorder()
	Redirect(rr, "/next")

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}

	var decoded string
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if decoded != "/next" {
		t.Fatalf("decoded = %q, want /next", decoded)
	}
}

func TestAPIResponseSendWritesHeadersAndJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := &APIResponse{
		Headers: map[string]string{"X-Custom": "value"},
		Data:    testResponsePayload{OK: true, Header: "send"},
	}

	if err := resp.Send(rr); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if got := rr.Header().Get("X-Custom"); got != "value" {
		t.Fatalf("X-Custom = %q, want value", got)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var decoded testResponsePayload
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !decoded.OK || decoded.Header != "send" {
		t.Fatalf("decoded = %+v, want OK with header send", decoded)
	}
}

func TestForwardResponseCopiesStatusHeadersAndBody(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Error-Code", "7")
		http.Error(w, "upstream boom", http.StatusBadGateway)
	}))
	defer upstream.Close()

	apiResp, err := NewRequest(http.MethodGet, upstream.URL).Do()
	if err == nil {
		t.Fatal("Do() error = nil, want non-OK error from upstream")
	}
	defer apiResp.CloseBody()

	rr := httptest.NewRecorder()
	ForwardResponse(rr, apiResp)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/plain; charset=utf-8", got)
	}
	if got := rr.Header().Get("Error-Code"); got != "7" {
		t.Fatalf("Error-Code = %q, want 7", got)
	}
	if rr.Body.String() != "upstream boom\n" {
		t.Fatalf("body = %q, want upstream boom", rr.Body.String())
	}
}

func TestAPIResponseCloseBodyNilSafe(t *testing.T) {
	if err := (&APIResponse{}).CloseBody(); err != nil {
		t.Fatalf("CloseBody on empty response = %v, want nil", err)
	}
}
