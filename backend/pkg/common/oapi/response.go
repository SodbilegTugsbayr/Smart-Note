package oapi

import (
	"encoding/json"
	"log"
	"net/http"
)

type APIResponse struct {
	Headers    map[string]string
	Data       interface{}
	Code       int
	ErrMessage string
	Response   *http.Response
}

func NewResponse(Data interface{}) *APIResponse {
	return &APIResponse{Data: Data}
}

func SendResp(w http.ResponseWriter, Data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(Data)
	if err != nil {
		log.Fatal(err)
	}
}

func SendFormError(w http.ResponseWriter, Data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	return json.NewEncoder(w).Encode(Data)
}

func ForwardResponse(w http.ResponseWriter, apiResp *APIResponse) {
	// Copy headers from the response to our relay
	w.Header().Set("Content-Type", apiResp.Response.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", apiResp.Response.Header.Get("Content-Length"))
	w.Header().Set("Error-Code", apiResp.Response.Header.Get("Error-Code"))

	// Copy status
	w.WriteHeader(apiResp.Response.StatusCode)

	// Copy the body
	_, err := w.Write([]byte(apiResp.ErrMessage))
	if err != nil {
		log.Fatal(err)
	}
}

func Redirect(w http.ResponseWriter, url string) {
	w.WriteHeader(http.StatusSeeOther)
	SendResp(w, url)
}

func (resp *APIResponse) Send(w http.ResponseWriter) error {
	for key, value := range resp.Headers {
		w.Header().Set(key, value)
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(resp.Data)
}

func (resp *APIResponse) CloseBody() error {
	if resp.Response != nil && resp.Response.Body != nil {
		return resp.Response.Body.Close()
	}
	return nil
}
