package quizman_test

import (
	"errors"
	"testing"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/internal/testdb"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/courseman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/quizman"
)

func TestQuizServiceDatabaseCRUDFiltersAndResults(t *testing.T) {
	db := testdb.Open(t)
	service := quizman.NewService(db, testdb.DiscardLogger(), testdb.DiscardLogger())

	course := &courseman.Course{UserID: 10, Title: "Algorithms", Description: "Course"}
	if err := db.Save(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}

	note := &noteman.Note{CourseID: course.ID, Title: "Sorting note"}
	if err := db.Save(note).Error; err != nil {
		t.Fatalf("create note: %v", err)
	}

	first, err := service.Save(&quizman.Quiz{
		NoteID:        note.ID,
		Question:      "What is stable sorting?",
		Options:       []string{"A", "B", "C", "D"},
		CorrectAnswer: "A",
	})
	if err != nil {
		t.Fatalf("Save(first) error = %v", err)
	}
	if _, err := service.Save(&quizman.Quiz{
		NoteID:        note.ID,
		Question:      "What is a graph?",
		Options:       []string{"A", "B", "C", "D"},
		CorrectAnswer: "B",
	}); err != nil {
		t.Fatalf("Save(second) error = %v", err)
	}

	got, err := service.GetByID(first.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Question != first.Question || got.Options[0] != "A" {
		t.Fatalf("GetByID() = %+v, want first quiz with options", got)
	}

	assertQuizFilterCount(t, service, &quizman.Filter{NoteID: note.ID}, 2)
	assertQuizFilterCount(t, service, &quizman.Filter{Keyword: "graph"}, 1)
	if _, err := service.Get(&quizman.Quiz{Question: "Missing"}); !errors.Is(err, quizman.ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrNotFound", err)
	}
	if _, err := service.GetByID(999999); !errors.Is(err, quizman.ErrNotFound) {
		t.Fatalf("GetByID(missing) error = %v, want ErrNotFound", err)
	}

	result, err := service.SaveResult(&quizman.QuizResult{
		UserID:     7,
		NoteID:     note.ID,
		Score:      2,
		Total:      2,
		Percentage: 100,
		Passed:     true,
	})
	if err != nil {
		t.Fatalf("SaveResult() error = %v", err)
	}
	if result.ID == 0 || !result.Passed {
		t.Fatalf("SaveResult() = %+v, want saved passing result", result)
	}

	if err := service.Delete(first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.GetByID(first.ID); !errors.Is(err, quizman.ErrNotFound) {
		t.Fatalf("GetByID(deleted) error = %v, want ErrNotFound", err)
	}
}

func TestQuizServiceCountAndUserFilter(t *testing.T) {
	db := testdb.Open(t)
	service := quizman.NewService(db, testdb.DiscardLogger(), testdb.DiscardLogger())

	course := &courseman.Course{UserID: 1, Title: "Physics"}
	if err := db.Save(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}
	note := &noteman.Note{CourseID: course.ID, Title: "Kinematics"}
	if err := db.Save(note).Error; err != nil {
		t.Fatalf("create note: %v", err)
	}

	if _, err := service.Save(&quizman.Quiz{NoteID: note.ID, Question: "v=?", Options: []string{"A", "B"}, CorrectAnswer: "A"}); err != nil {
		t.Fatalf("Save(q1): %v", err)
	}
	if _, err := service.Save(&quizman.Quiz{NoteID: note.ID, Question: "a=?", Options: []string{"A", "B"}, CorrectAnswer: "B"}); err != nil {
		t.Fatalf("Save(q2): %v", err)
	}

	total, err := service.Count(nil)
	if err != nil {
		t.Fatalf("Count(nil) error = %v", err)
	}
	if total != 2 {
		t.Fatalf("Count(nil) = %d, want 2", total)
	}

	total, err = service.Count(&quizman.Filter{NoteID: note.ID})
	if err != nil {
		t.Fatalf("Count(NoteID) error = %v", err)
	}
	if total != 2 {
		t.Fatalf("Count(NoteID) = %d, want 2", total)
	}

	if _, _, err := service.GetAll(&quizman.Filter{OrderBy: "id desc"}, 0, 0); err != nil {
		t.Fatalf("GetAll(OrderBy) error = %v", err)
	}
}

func assertQuizFilterCount(t *testing.T, service *quizman.Service, filter *quizman.Filter, want int) {
	t.Helper()

	quizzes, total, err := service.GetAll(filter, 1, 25)
	if err != nil {
		t.Fatalf("GetAll(%+v) error = %v", filter, err)
	}
	if total != want || len(quizzes) != want {
		t.Fatalf("GetAll(%+v) returned len=%d total=%d, want %d", filter, len(quizzes), total, want)
	}
}
