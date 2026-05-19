package main

import (
	"fmt"
	"os"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/eguneapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/ocrapi"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func processNote(note *noteman.Note) {
	var filePath string

	note.ProcessStatus = noteman.PROCESS_STATUS_PROCESSING
	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to update processing status: ", err)
	}

	if note.IsFromBook {
		tempFile, err := extractPages(note.FilePath, note.StartPage, note.EndPage)
		if err != nil {
			app.ErrorLog.Println("failed to extract pages: ", err)
			markNoteProcessingFailed(note)
			return
		}
		defer os.Remove(tempFile)
		filePath = tempFile
	} else {
		filePath = note.FilePath
	}

	rawText, err := ocrapi.GetTextFromFile(filePath)
	if err != nil {
		app.ErrorLog.Println("OCR failed: ", err)
		markNoteProcessingFailed(note)
		return
	}

	note.RawContent = rawText
	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to save note: ", err)
	}

	output, err := eguneapi.GenerateNote(rawText)
	if err != nil {
		app.ErrorLog.Println("failed to generate note content: ", err)
		markNoteProcessingFailed(note)
		return
	}

	for _, quiz := range output.Quizzes {
		quiz.NoteID = note.ID
		if _, err := app.Quizzes.Save(&quiz); err != nil {
			app.ErrorLog.Println("failed to save quiz: ", err)
		}
	}

	note.Title = output.Note.Title
	note.Summary = output.Note.Summary
	note.KeyConcepts = output.Note.KeyConcepts
	note.FlashCards = output.Note.FlashCards
	note.ProcessStatus = noteman.PROCESS_STATUS_COMPLETED
	note.Status = noteman.STATUS_COMPLETED

	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to save note: ", err)
	}
}

func markNoteProcessingFailed(note *noteman.Note) {
	note.ProcessStatus = noteman.PROCESS_STATUS_FAILED
	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to mark note processing failed: ", err)
	}
}

func extractPages(src string, start, end int) (string, error) {
	pageRange := fmt.Sprintf("%d-%d", start, end)

	tmp, err := os.CreateTemp("", "note_pages_*.pdf")
	if err != nil {
		return "", err
	}
	tmp.Close()

	err = api.TrimFile(src, tmp.Name(), []string{pageRange}, nil)
	if err != nil {
		return "", err
	}

	return tmp.Name(), nil
}
