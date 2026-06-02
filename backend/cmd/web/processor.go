package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/eguneapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/quizman"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const noteProcessProgressMessageType = "NOTE_PROCESS_PROGRESS"
const noteQuizRegenerationMessageType = "NOTE_QUIZ_REGENERATION"

const (
	noteProcessWorkerCount = 4
	ocrWorkerLimit         = 4
	eguneWorkerLimit       = 2
	noteProcessMaxAttempts = 3
	noteProcessPollDelay   = 2 * time.Second
	noteProcessStaleAfter  = 30 * time.Minute
)

var (
	ocrLimiter   = make(chan struct{}, ocrWorkerLimit)
	eguneLimiter = make(chan struct{}, eguneWorkerLimit)
)

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

type noteQuizRegenerationPayload struct {
	CourseID int            `json:"course_id"`
	NoteID   int            `json:"note_id"`
	Stage    string         `json:"stage"`
	Message  string         `json:"message"`
	Error    string         `json:"error,omitempty"`
	Quizzes  []quizResponse `json:"quizzes,omitempty"`
}

func startNoteProcessingWorkers() func() {
	ctx, cancel := context.WithCancel(context.Background())
	if err := requeueStaleNoteProcessJobs(); err != nil {
		app.ErrorLog.Println("failed to requeue stale note process jobs: ", err)
	}
	for i := 0; i < noteProcessWorkerCount; i++ {
		go noteProcessWorker(ctx)
	}
	return cancel
}

func requeueStaleNoteProcessJobs() error {
	now := time.Now()
	staleBefore := now.Add(-noteProcessStaleAfter)
	return app.DB.Model(&noteman.NoteProcessJob{}).
		Where("status = ? AND started_at < ? AND attempts < max_attempts", noteman.PROCESS_JOB_STATUS_PROCESSING, staleBefore).
		Updates(map[string]interface{}{
			"status":      noteman.PROCESS_JOB_STATUS_QUEUED,
			"next_run_at": now,
			"started_at":  nil,
		}).Error
}

func noteProcessWorker(ctx context.Context) {
	ticker := time.NewTicker(noteProcessPollDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		job, err := claimNextNoteProcessJob()
		if err != nil {
			app.ErrorLog.Println("failed to claim note process job: ", err)
		}
		if job == nil {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			continue
		}

		processNoteJob(job)
	}
}

func enqueueNoteProcessing(noteID, recipientUserID int) error {
	note, err := app.Notes.GetByID(noteID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(note.FilePath) == "" {
		return fmt.Errorf("note has no source file")
	}

	now := time.Now()
	return app.DB.Transaction(func(tx *gorm.DB) error {
		var existing noteman.NoteProcessJob
		err := tx.Where("note_id = ? AND status IN ?", noteID, []string{
			noteman.PROCESS_JOB_STATUS_QUEUED,
			noteman.PROCESS_JOB_STATUS_PROCESSING,
		}).Order("created_at desc").First(&existing).Error
		if err == nil {
			if recipientUserID > 0 && existing.RecipientUserID == 0 {
				existing.RecipientUserID = recipientUserID
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		} else {
			job := &noteman.NoteProcessJob{
				NoteID:          noteID,
				RecipientUserID: recipientUserID,
				Status:          noteman.PROCESS_JOB_STATUS_QUEUED,
				MaxAttempts:     noteProcessMaxAttempts,
				NextRunAt:       &now,
			}
			if err := tx.Create(job).Error; err != nil {
				return err
			}
		}

		note.ProcessStatus = noteman.PROCESS_STATUS_QUEUED
		if err := tx.Save(note).Error; err != nil {
			return err
		}
		publishNoteProcessProgress(note, []int{recipientUserID}, "queued", 5, "AI боловсруулалтын дараалалд орлоо", "", true)
		return nil
	})
}

func claimNextNoteProcessJob() (*noteman.NoteProcessJob, error) {
	var claimed *noteman.NoteProcessJob
	now := time.Now()

	if err := app.DB.Transaction(func(tx *gorm.DB) error {
		var job noteman.NoteProcessJob
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND (next_run_at IS NULL OR next_run_at <= ?)", noteman.PROCESS_JOB_STATUS_QUEUED, now).
			Order("created_at asc, id asc").
			First(&job).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}

		job.Status = noteman.PROCESS_JOB_STATUS_PROCESSING
		job.Attempts++
		job.StartedAt = &now
		job.FinishedAt = nil
		if job.MaxAttempts <= 0 {
			job.MaxAttempts = noteProcessMaxAttempts
		}
		if err := tx.Save(&job).Error; err != nil {
			return err
		}
		claimed = &job
		return nil
	}); err != nil {
		return nil, err
	}

	return claimed, nil
}

func processNoteJob(job *noteman.NoteProcessJob) {
	note, err := app.Notes.GetByID(job.NoteID)
	if err != nil {
		markNoteProcessJobFailed(job, err)
		return
	}

	err = processNote(note, job.RecipientUserID)
	if err == nil {
		markNoteProcessJobCompleted(job)
		return
	}

	if job.Attempts < job.MaxAttempts && isRetryableProcessingError(err) {
		requeueNoteProcessJob(job, note, err)
		return
	}
	markNoteProcessJobFailed(job, err)
}

func markNoteProcessJobCompleted(job *noteman.NoteProcessJob) {
	now := time.Now()
	job.Status = noteman.PROCESS_JOB_STATUS_COMPLETED
	job.FinishedAt = &now
	job.LastError = ""
	if err := app.DB.Save(job).Error; err != nil {
		app.ErrorLog.Println("failed to mark note process job completed: ", err)
	}
}

func requeueNoteProcessJob(job *noteman.NoteProcessJob, note *noteman.Note, cause error) {
	delay := time.Duration(job.Attempts*job.Attempts) * 10 * time.Second
	nextRunAt := time.Now().Add(delay)
	job.Status = noteman.PROCESS_JOB_STATUS_QUEUED
	job.LastError = cause.Error()
	job.NextRunAt = &nextRunAt
	job.FinishedAt = nil

	if err := app.DB.Save(job).Error; err != nil {
		app.ErrorLog.Println("failed to requeue note process job: ", err)
	}
	if note != nil {
		note.ProcessStatus = noteman.PROCESS_STATUS_QUEUED
		if _, err := app.Notes.Save(note); err != nil {
			app.ErrorLog.Println("failed to mark note queued after retry: ", err)
		}
		publishNoteProcessProgress(note, []int{job.RecipientUserID}, "queued", 5, "Алдаа гарсан тул дахин оролдохоор дараалалд орууллаа", cause.Error(), true)
	}
}

func markNoteProcessJobFailed(job *noteman.NoteProcessJob, cause error) {
	now := time.Now()
	job.Status = noteman.PROCESS_JOB_STATUS_FAILED
	job.FinishedAt = &now
	if cause != nil {
		job.LastError = cause.Error()
	}
	if err := app.DB.Save(job).Error; err != nil {
		app.ErrorLog.Println("failed to mark note process job failed: ", err)
	}
}

func isRetryableProcessingError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	retryableParts := []string{
		"chat completion error",
		"500",
		"502",
		"503",
		"504",
		"429",
		"timeout",
		"deadline",
		"temporarily",
		"connection",
		"eof",
	}
	for _, part := range retryableParts {
		if strings.Contains(message, part) {
			return true
		}
	}
	return false
}

func runLimited[T any](limiter chan struct{}, fn func() (T, error)) (T, error) {
	limiter <- struct{}{}
	defer func() { <-limiter }()
	return fn()
}

func saveNoteProcessStatus(note *noteman.Note, status string) {
	note.ProcessStatus = status
	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to update note process status: ", err)
	}
}

func processNote(note *noteman.Note, recipientUserIDs ...int) error {
	var filePath string

	saveNoteProcessStatus(note, noteman.PROCESS_STATUS_PROCESSING)
	publishNoteProcessProgress(note, recipientUserIDs, "started", 10, "AI боловсруулалт эхэллээ", "", true)

	if note.IsFromBook {
		tempFile, err := extractPages(note.FilePath, note.StartPage, note.EndPage)
		if err != nil {
			app.ErrorLog.Println("failed to extract pages: ", err)
			markNoteProcessingFailed(note, recipientUserIDs, "Файлын хуудсыг бэлтгэхэд алдаа гарлаа", err)
			return err
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
		return err
	}

	saveNoteProcessStatus(note, noteman.PROCESS_STATUS_OCR_PROCESSING)
	publishNoteProcessProgress(note, recipientUserIDs, noteman.PROCESS_STATUS_OCR_PROCESSING, 25, "Файлаас текст таньж байна", "", true)
	rawText, err := runLimited(ocrLimiter, func() (string, error) {
		return app.OCRService.GetTextFromFile(filePath)
	})
	if err != nil {
		app.ErrorLog.Println("OCR failed: ", err)
		markNoteProcessingFailed(note, recipientUserIDs, "Файлаас текст танихад алдаа гарлаа", err)
		return err
	}

	note.RawContent = rawText
	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to save note: ", err)
	}
	publishNoteProcessProgress(note, recipientUserIDs, "ocr_completed", 60, "Текст таньж дууслаа", "", false)

	saveNoteProcessStatus(note, noteman.PROCESS_STATUS_AI_GENERATING)
	publishNoteProcessProgress(note, recipientUserIDs, noteman.PROCESS_STATUS_AI_GENERATING, 75, "Тэмдэглэлийн агуулга үүсгэж байна", "", true)
	output, err := runLimited(eguneLimiter, func() (*eguneapi.GeneratedOutput, error) {
		return app.EguneService.GenerateNote(rawText)
	})
	if err != nil {
		app.ErrorLog.Println("failed to generate note content: ", err)
		markNoteProcessingFailed(note, recipientUserIDs, "Тэмдэглэлийн агуулга үүсгэхэд алдаа гарлаа", err)
		return err
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
	if strings.TrimSpace(note.Status) == "" {
		note.Status = noteman.STATUS_IN_PROGRESS
	}

	if _, err := app.Notes.Save(note); err != nil {
		app.ErrorLog.Println("failed to save note: ", err)
		markNoteProcessingFailed(note, recipientUserIDs, "Тэмдэглэл хадгалахад алдаа гарлаа", err)
		return err
	}

	publishNoteProcessProgress(note, recipientUserIDs, "completed", 100, "Тэмдэглэл бэлэн боллоо", "", true)
	return nil
}

func regenerateNoteQuizzes(note *noteman.Note, recipientUserIDs ...int) {
	if note == nil {
		return
	}

	publishNoteQuizRegeneration(note, recipientUserIDs, "started", "Тест дахин үүсгэж байна", "", nil)

	content := strings.TrimSpace(note.RawContent)
	if content == "" {
		content = strings.TrimSpace(note.Summary)
	}
	if content == "" {
		err := fmt.Errorf("note has no content for quiz regeneration")
		app.ErrorLog.Println(err)
		publishNoteQuizRegeneration(note, recipientUserIDs, "failed", "Тест дахин үүсгэх агуулга олдсонгүй", err.Error(), nil)
		return
	}

	output, err := runLimited(eguneLimiter, func() (*eguneapi.GeneratedOutput, error) {
		return app.EguneService.GenerateNote(content)
	})
	if err != nil {
		app.ErrorLog.Println("failed to regenerate quizzes: ", err)
		publishNoteQuizRegeneration(note, recipientUserIDs, "failed", "Тест дахин үүсгэхэд алдаа гарлаа", err.Error(), nil)
		return
	}

	generatedQuizzes := output.Quizzes
	if len(generatedQuizzes) == 0 {
		err := fmt.Errorf("quiz regeneration returned no quizzes")
		app.ErrorLog.Println(err)
		publishNoteQuizRegeneration(note, recipientUserIDs, "failed", "Шинэ тестийн асуулт үүссэнгүй", err.Error(), nil)
		return
	}

	savedQuizzes := make([]quizman.Quiz, 0, len(generatedQuizzes))
	if err := app.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("note_id = ?", note.ID).Delete(&quizman.Quiz{}).Error; err != nil {
			return err
		}

		for i := range generatedQuizzes {
			generatedQuizzes[i].NoteID = note.ID
			if err := tx.Save(&generatedQuizzes[i]).Error; err != nil {
				return err
			}
			savedQuizzes = append(savedQuizzes, generatedQuizzes[i])
		}

		return nil
	}); err != nil {
		app.ErrorLog.Println("failed to replace regenerated quizzes: ", err)
		publishNoteQuizRegeneration(note, recipientUserIDs, "failed", "Шинэ тест хадгалахад алдаа гарлаа", err.Error(), nil)
		return
	}

	publishNoteQuizRegeneration(note, recipientUserIDs, "completed", "Шинэ тест бэлэн боллоо", "", publicGeneratedQuizResponses(savedQuizzes))
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

func publishNoteQuizRegeneration(note *noteman.Note, recipientUserIDs []int, stage, message, errMessage string, quizzes []quizResponse) {
	if note == nil {
		return
	}

	payload := noteQuizRegenerationPayload{
		CourseID: note.CourseID,
		NoteID:   note.ID,
		Stage:    stage,
		Message:  message,
		Error:    errMessage,
		Quizzes:  quizzes,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		app.ErrorLog.Println("failed to marshal quiz regeneration progress: ", err)
		return
	}

	for _, userID := range noteProcessRecipientIDs(note, recipientUserIDs) {
		sendUserSocketMessage(userID, noteQuizRegenerationMessageType, string(data))
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
