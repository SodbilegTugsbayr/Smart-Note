package main

import (
	"errors"
	"net/http"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
)

const (
	maxUploadFileSize      int64 = 50 << 20
	maxMultipartUploadSize int64 = maxUploadFileSize + (5 << 20)
)

func parseLimitedMultipartForm(w http.ResponseWriter, r *http.Request) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartUploadSize)

	if err := r.ParseMultipartForm(maxUploadFileSize); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			oapi.CustomError(w, http.StatusRequestEntityTooLarge, "file too large: max 50MB allowed")
			return false
		}

		oapi.CustomError(w, http.StatusBadRequest, "Invalid form data")
		return false
	}

	return true
}
