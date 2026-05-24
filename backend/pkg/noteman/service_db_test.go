package noteman_test

import (
	"errors"
	"testing"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/internal/testdb"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/courseman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
)

func TestNoteServiceDatabaseCRUDAndFilters(t *testing.T) {
	db := testdb.Open(t)
	service := noteman.NewService(db, testdb.DiscardLogger(), testdb.DiscardLogger())

	course := &courseman.Course{UserID: 10, Title: "Algorithms", Description: "Course"}
	if err := db.Save(course).Error; err != nil {
		t.Fatalf("create course: %v", err)
	}

	first, err := service.Save(&noteman.Note{
		CourseID:      course.ID,
		Title:         "Sorting note",
		Summary:       "Quick sort and merge sort",
		Status:        noteman.STATUS_IN_PROGRESS,
		ProcessStatus: noteman.PROCESS_STATUS_COMPLETED,
		FilePath:      "/tmp/sorting.pdf",
	})
	if err != nil {
		t.Fatalf("Save(first) error = %v", err)
	}
	if _, err := service.Save(&noteman.Note{
		CourseID:      course.ID,
		Title:         "Graph note",
		Summary:       "Trees",
		Status:        noteman.STATUS_COMPLETED,
		ProcessStatus: noteman.PROCESS_STATUS_PROCESSING,
	}); err != nil {
		t.Fatalf("Save(second) error = %v", err)
	}

	got, err := service.GetByID(first.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Title != first.Title || !got.HasFile {
		t.Fatalf("GetByID() = %+v, want title and HasFile", got)
	}

	assertNoteFilterCount(t, service, &noteman.Filter{CourseID: course.ID}, 2)
	assertNoteFilterCount(t, service, &noteman.Filter{Keyword: "Quick"}, 1)
	assertNoteFilterCount(t, service, &noteman.Filter{ProcessStatus: noteman.PROCESS_STATUS_PROCESSING}, 1)
	if _, err := service.Get(&noteman.Note{Title: "Missing"}); !errors.Is(err, noteman.ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrNotFound", err)
	}
	if _, err := service.GetByID(999999); !errors.Is(err, noteman.ErrNotFound) {
		t.Fatalf("GetByID(missing) error = %v, want ErrNotFound", err)
	}

	if err := service.Delete(first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.GetByID(first.ID); !errors.Is(err, noteman.ErrNotFound) {
		t.Fatalf("GetByID(deleted) error = %v, want ErrNotFound", err)
	}
}

func assertNoteFilterCount(t *testing.T, service *noteman.Service, filter *noteman.Filter, want int) {
	t.Helper()

	notes, total, err := service.GetAll(filter, 1, 25)
	if err != nil {
		t.Fatalf("GetAll(%+v) error = %v", filter, err)
	}
	if total != want || len(notes) != want {
		t.Fatalf("GetAll(%+v) returned len=%d total=%d, want %d", filter, len(notes), total, want)
	}
}
