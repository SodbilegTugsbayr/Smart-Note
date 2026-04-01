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

	if note.IsFromBook {
		// Assuming you want to create a temporary PDF with only the pages from StartPage to EndPage
		tempFile, err := extractPages(note.FilePath, note.StartPage, note.EndPage)
		if err != nil {
			app.ErrorLog.Println("failed to extract pages: ", err)
		}
		defer os.Remove(tempFile)
		filePath = tempFile
	} else {
		filePath = note.FilePath
	}

	rawText, err := ocrapi.GetTextFromFile(filePath)
	if err != nil {
		app.ErrorLog.Println("OCR failed: ", err)
	}

	note.RawContent = rawText
	app.InfoLog.Println("Raw content: ", rawText)
	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to save note: ", err)
	}

	output, err := eguneapi.GenerateNote(rawText)
	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to generate note contetn: ", err)
	}

	for _, quiz := range output.Quizzes {
		app.InfoLog.Println("Question: ", quiz.Question)
	}

	note.Summary = output.Note.Summary
	note.KeyConcepts = output.Note.KeyConcepts
	note.FlashCards = output.Note.FlashCards

	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to save note: ", err)
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
