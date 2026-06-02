package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/courseman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/quizman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

func getAdminStats(w http.ResponseWriter, r *http.Request) {
	userTotal, err := app.Users.Count(nil)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	adminTotal, err := app.Users.Count(&userman.Filter{Role: userman.ROLE_ADMIN})
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	regularUserTotal, err := app.Users.Count(&userman.Filter{Role: userman.ROLE_USER})
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	courseTotal, err := app.Courses.Count(nil)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	noteTotal, err := app.Notes.Count(nil)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	quizTotal, err := app.Quizzes.Count(nil)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	completedCourses, err := countWhere(&courseman.Course{}, "status = ?", courseman.STATUS_COMPLETED)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	inProgressCourses, err := countWhere(&courseman.Course{}, "status = ?", courseman.STATUS_IN_PROGRESS)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	publicCourses, err := countWhere(&courseman.Course{}, "is_public = ?", true)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	completedNotes, err := countWhere(&noteman.Note{}, "process_status = ?", noteman.PROCESS_STATUS_COMPLETED)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	processingNotes, err := countWhere(&noteman.Note{}, "process_status IN ?", []string{
		noteman.PROCESS_STATUS_PROCESSING,
		noteman.PROCESS_STATUS_OCR_PROCESSING,
		noteman.PROCESS_STATUS_AI_GENERATING,
	})
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	queuedNotes, err := countWhere(&noteman.Note{}, "process_status = ?", noteman.PROCESS_STATUS_QUEUED)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	failedNotes, err := countWhere(&noteman.Note{}, "process_status = ?", noteman.PROCESS_STATUS_FAILED)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	draftNotes, err := countWhere(&noteman.Note{}, "process_status = '' OR process_status IS NULL")
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	notesWithFiles, err := countWhere(&noteman.Note{}, "COALESCE(file_path, '') <> ''")
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	recentCourses, _, err := app.Courses.GetAll(&courseman.Filter{OrderBy: "created_at desc"}, 1, 5, "Notes")
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	recentNotes, _, err := app.Notes.GetAll(&noteman.Filter{OrderBy: "created_at desc"}, 1, 5)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, map[string]interface{}{
		"totals": map[string]int{
			"users":   userTotal,
			"courses": courseTotal,
			"notes":   noteTotal,
			"quizzes": quizTotal,
		},
		"users": map[string]int{
			"admins":  adminTotal,
			"regular": regularUserTotal,
		},
		"courses": map[string]int{
			"completed":   completedCourses,
			"in_progress": inProgressCourses,
			"public":      publicCourses,
		},
		"notes": map[string]int{
			"completed":  completedNotes,
			"queued":     queuedNotes,
			"processing": processingNotes,
			"failed":     failedNotes,
			"draft":      draftNotes,
			"with_files": notesWithFiles,
		},
		"recent_courses": recentCourses,
		"recent_notes":   recentNotes,
	})
}

func getAdminCourses(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, size := pageAndSize(q.Get("page"), q.Get("size"))

	courses, total, err := app.Courses.GetAll(&courseman.Filter{
		Keyword: q.Get("keyword"),
		OrderBy: safeAdminOrder(q.Get("order_by"), "created_at desc"),
	}, page, size, "Notes")
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	users, err := usersForCourses(courses)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, map[string]interface{}{
		"items": courses,
		"total": total,
		"users": users,
	})
}

func getAdminNotes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, size := pageAndSize(q.Get("page"), q.Get("size"))
	courseID, _ := strconv.Atoi(q.Get("course_id"))

	notes, total, err := app.Notes.GetAll(&noteman.Filter{
		Keyword:       q.Get("keyword"),
		CourseID:      courseID,
		ProcessStatus: safeNoteProcessStatus(q.Get("process_status")),
		OrderBy:       safeAdminOrder(q.Get("order_by"), "created_at desc"),
	}, page, size)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	courses, err := coursesForNotes(notes)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}
	users, err := usersForCoursesMap(courses)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, map[string]interface{}{
		"items":   notes,
		"total":   total,
		"courses": courses,
		"users":   users,
	})
}

func reprocessAdminNote(w http.ResponseWriter, r *http.Request) {
	chosenNote := r.Context().Value(app.ContextKeyChosenNote).(*noteman.Note)
	loggedUser := r.Context().Value(app.ContextKeyAuthUser).(*userman.User)
	if !noteHasSourceFile(chosenNote) {
		oapi.CustomError(w, http.StatusBadRequest, "Дахин боловсруулах эх файл алга")
		return
	}
	if noteProcessStatusIsActive(chosenNote.ProcessStatus) {
		oapi.CustomError(w, http.StatusBadRequest, "Тэмдэглэл одоо боловсруулагдаж байна")
		return
	}

	if err := resetGeneratedNoteContent(chosenNote); err != nil {
		oapi.ServerError(w, err)
		return
	}

	if err := enqueueNoteProcessing(chosenNote.ID, loggedUser.ID); err != nil {
		oapi.ServerError(w, err)
		return
	}

	chosenNote.PrepareResponse()
	oapi.SendResp(w, chosenNote)
}

func resetGeneratedNoteContent(note *noteman.Note) error {
	quizzes, _, err := app.Quizzes.GetAll(&quizman.Filter{NoteID: note.ID}, 1, 0)
	if err != nil {
		return err
	}
	for _, quiz := range quizzes {
		if err := app.Quizzes.Delete(quiz.ID); err != nil {
			return err
		}
	}

	note.Summary = ""
	note.RawContent = ""
	note.KeyConcepts = nil
	note.FlashCards = nil
	note.Status = noteman.STATUS_IN_PROGRESS
	note.ProcessStatus = noteman.PROCESS_STATUS_QUEUED

	_, err = app.Notes.Save(note)
	return err
}

func countWhere(model interface{}, query string, args ...interface{}) (int, error) {
	var count int64
	if err := app.DB.Model(model).Where(query, args...).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func pageAndSize(rawPage, rawSize string) (int, int) {
	page, _ := strconv.Atoi(rawPage)
	size, _ := strconv.Atoi(rawSize)
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 25
	}
	return page, size
}

func safeAdminOrder(value, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "id", "id desc", "created_at", "created_at desc", "title", "title desc", "course_id", "course_id desc", "process_status", "process_status desc":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return fallback
	}
}

func safeNoteProcessStatus(value string) string {
	switch strings.TrimSpace(value) {
	case noteman.PROCESS_STATUS_QUEUED,
		noteman.PROCESS_STATUS_COMPLETED,
		noteman.PROCESS_STATUS_PROCESSING,
		noteman.PROCESS_STATUS_OCR_PROCESSING,
		noteman.PROCESS_STATUS_AI_GENERATING,
		noteman.PROCESS_STATUS_FAILED:
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func noteProcessStatusIsActive(status string) bool {
	switch strings.TrimSpace(status) {
	case noteman.PROCESS_STATUS_QUEUED,
		noteman.PROCESS_STATUS_PROCESSING,
		noteman.PROCESS_STATUS_OCR_PROCESSING,
		noteman.PROCESS_STATUS_AI_GENERATING:
		return true
	default:
		return false
	}
}

func usersForCourses(courses []*courseman.Course) (map[int]*userman.User, error) {
	ids := make([]int, 0, len(courses))
	seen := make(map[int]bool)
	for _, course := range courses {
		if course.UserID > 0 && !seen[course.UserID] {
			ids = append(ids, course.UserID)
			seen[course.UserID] = true
		}
	}
	return usersByIDs(ids)
}

func usersForCoursesMap(courses map[int]*courseman.Course) (map[int]*userman.User, error) {
	ids := make([]int, 0, len(courses))
	seen := make(map[int]bool)
	for _, course := range courses {
		if course.UserID > 0 && !seen[course.UserID] {
			ids = append(ids, course.UserID)
			seen[course.UserID] = true
		}
	}
	return usersByIDs(ids)
}

func usersByIDs(ids []int) (map[int]*userman.User, error) {
	users := make(map[int]*userman.User)
	if len(ids) == 0 {
		return users, nil
	}

	list, _, err := app.Users.GetAll(&userman.Filter{IDs: ids}, 1, 0)
	if err != nil {
		return nil, err
	}
	for _, user := range list {
		users[user.ID] = user
	}
	return users, nil
}

func coursesForNotes(notes []*noteman.Note) (map[int]*courseman.Course, error) {
	courseIDs := make([]int, 0, len(notes))
	seen := make(map[int]bool)
	for _, note := range notes {
		if note.CourseID > 0 && !seen[note.CourseID] {
			courseIDs = append(courseIDs, note.CourseID)
			seen[note.CourseID] = true
		}
	}

	courses := make(map[int]*courseman.Course)
	if len(courseIDs) == 0 {
		return courses, nil
	}

	var list []*courseman.Course
	if err := app.DB.Where("id IN ?", courseIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, course := range list {
		courses[course.ID] = course
	}
	return courses, nil
}
