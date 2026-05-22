package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

const noteProcessProgressMessageType = "NOTE_PROCESS_PROGRESS"

type noteProcessProgressPayload struct {
	CourseID      int           `json:"course_id"`
	NoteID        int           `json:"note_id"`
	Stage         string        `json:"stage"`
	Progress      int           `json:"progress"`
	Message       string        `json:"message"`
	ProcessStatus string        `json:"process_status"`
	Error         string        `json:"error,omitempty"`
	Note          *noteman.Note `json:"note,omitempty"`
}

func processNote(note *noteman.Note, recipientUserIDs ...int) {
	var filePath string

	note.ProcessStatus = noteman.PROCESS_STATUS_PROCESSING
	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to update processing status: ", err)
	}
	publishNoteProcessProgress(note, recipientUserIDs, "started", 10, "AI боловсруулалт эхэллээ", "", false)

	if note.IsFromBook {
		tempFile, err := extractPages(note.FilePath, note.StartPage, note.EndPage)
		if err != nil {
			app.ErrorLog.Println("failed to extract pages: ", err)
			markNoteProcessingFailed(note, recipientUserIDs, "Файлын хуудсыг бэлтгэхэд алдаа гарлаа", err)
			return
		}
		defer os.Remove(tempFile)
		filePath = tempFile
	} else {
		filePath = note.FilePath
	}
	if strings.TrimSpace(filePath) == "" {
		err := fmt.Errorf("note has no source file")
		app.ErrorLog.Println(err)
		markNoteProcessingFailed(note, recipientUserIDs, "Боловсруулах файл олдсонгүй", err)
		return
	}

	publishNoteProcessProgress(note, recipientUserIDs, "ocr_started", 25, "Файлаас текст таньж байна", "", false)
	rawText, err := app.OCRService.GetTextFromFile(filePath)
	if err != nil {
		app.ErrorLog.Println("OCR failed: ", err)
		markNoteProcessingFailed(note, recipientUserIDs, "Файлаас текст танихад алдаа гарлаа", err)
		return
	}

	note.RawContent = rawText
	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to save note: ", err)
	}
	publishNoteProcessProgress(note, recipientUserIDs, "ocr_completed", 60, "Текст таньж дууслаа", "", false)

	publishNoteProcessProgress(note, recipientUserIDs, "egune_started", 75, "Тэмдэглэлийн агуулга үүсгэж байна", "", false)
	output, err := app.EguneService.GenerateNote(rawText)
	if err != nil {
		app.ErrorLog.Println("failed to generate note content: ", err)
		markNoteProcessingFailed(note, recipientUserIDs, "Тэмдэглэлийн агуулга үүсгэхэд алдаа гарлаа", err)
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
		markNoteProcessingFailed(note, recipientUserIDs, "Тэмдэглэл хадгалахад алдаа гарлаа", err)
		return
	}

	publishNoteProcessProgress(note, recipientUserIDs, "completed", 100, "Тэмдэглэл бэлэн боллоо", "", true)
}

func markNoteProcessingFailed(note *noteman.Note, recipientUserIDs []int, message string, cause error) {
	note.ProcessStatus = noteman.PROCESS_STATUS_FAILED
	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to mark note processing failed: ", err)
	}

	errMessage := ""
	if cause != nil {
		errMessage = cause.Error()
	}
	publishNoteProcessProgress(note, recipientUserIDs, "failed", 100, message, errMessage, true)
}

func publishNoteProcessProgress(note *noteman.Note, recipientUserIDs []int, stage string, progress int, message, errMessage string, includeNote bool) {
	if note == nil {
		return
	}

	payload := noteProcessProgressPayload{
		CourseID:      note.CourseID,
		NoteID:        note.ID,
		Stage:         stage,
		Progress:      progress,
		Message:       message,
		ProcessStatus: note.ProcessStatus,
		Error:         errMessage,
	}
	if includeNote {
		note.PrepareResponse()
		payload.Note = note
	}

	data, err := json.Marshal(payload)
	if err != nil {
		app.ErrorLog.Println("failed to marshal note process progress: ", err)
		return
	}

	for _, userID := range noteProcessRecipientIDs(note, recipientUserIDs) {
		sendUserSocketMessage(userID, noteProcessProgressMessageType, string(data))
	}
}

func noteProcessRecipientIDs(note *noteman.Note, userIDs []int) []int {
	seen := make(map[int]bool)
	recipients := make([]int, 0, len(userIDs)+1)

	add := func(userID int) {
		if userID <= 0 || seen[userID] {
			return
		}
		seen[userID] = true
		recipients = append(recipients, userID)
	}

	for _, userID := range userIDs {
		add(userID)
	}

	if note != nil && note.CourseID > 0 {
		course, err := app.Courses.GetByID(note.CourseID)
		if err != nil {
			app.ErrorLog.Println("failed to resolve note process recipient: ", err)
		} else {
			add(course.UserID)
		}
	}

	return recipients
}

func sendUserSocketMessage(userID int, msgType, msg string) {
	app.CustomerWSCsMutex.RLock()
	conns := make([]interface{ Send(string, string) }, 0, len(app.CustomerWSCs[userID]))
	for _, conn := range app.CustomerWSCs[userID] {
		conns = append(conns, conn)
	}
	app.CustomerWSCsMutex.RUnlock()

	for _, conn := range conns {
		conn.Send(msgType, msg)
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
