package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
	"github.com/google/uuid"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if !parseLimitedMultipartForm(w, r) {
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		oapi.CustomError(w, http.StatusBadRequest, "File is required")
		return
	}
	_ = file.Close()

	uploadDir := filepath.Join(app.Config.FilePath, "attachments", uuid.NewString())
	uploadDir, _ = filepath.Abs(uploadDir)
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

	root, _ := filepath.Abs(app.Config.FilePath)
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, map[string]string{
		"file_url": "/pub/images/" + filepath.ToSlash(relPath),
	})
}
