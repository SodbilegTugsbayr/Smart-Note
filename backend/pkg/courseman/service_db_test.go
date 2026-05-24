package courseman_test

import (
	"errors"
	"testing"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/internal/testdb"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/courseman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
)

func TestCourseServiceDatabaseCRUDFiltersAndPreloads(t *testing.T) {
	db := testdb.Open(t)
	service := courseman.NewService(db, testdb.DiscardLogger(), testdb.DiscardLogger())

	first, err := service.Save(&courseman.Course{
		UserID:      10,
		Title:       "Algorithms",
		Description: "Sorting and graphs",
		Status:      courseman.STATUS_IN_PROGRESS,
		IsPublic:    true,
		Sections: []*courseman.Section{
			{SectionName: "Sorting", StartPage: 1, EndPage: 3},
		},
	})
	if err != nil {
		t.Fatalf("Save(first) error = %v", err)
	}
	second, err := service.Save(&courseman.Course{
		UserID:      20,
		Title:       "Databases",
		Description: "SQL and indexes",
		Status:      courseman.STATUS_COMPLETED,
	})
	if err != nil {
		t.Fatalf("Save(second) error = %v", err)
	}
	if err := db.Save(&noteman.Note{CourseID: first.ID, Title: "Sorting note"}).Error; err != nil {
		t.Fatalf("create note: %v", err)
	}

	assertCourseFilterCount(t, service, &courseman.Filter{UserID: 10}, 1)
	assertCourseFilterCount(t, service, &courseman.Filter{Keyword: "SQL"}, 1)

	courses, total, err := service.GetAll(&courseman.Filter{OrderBy: "title desc"}, 1, 1)
	if err != nil {
		t.Fatalf("GetAll(paged) error = %v", err)
	}
	if total != 2 || len(courses) != 1 || courses[0].ID != second.ID {
		t.Fatalf("GetAll(paged) got len=%d total=%d firstID=%d, want second course", len(courses), total, courses[0].ID)
	}

	withNotes, err := service.GetByID(first.ID, "Notes")
	if err != nil {
		t.Fatalf("GetByID(preload) error = %v", err)
	}
	if len(withNotes.Notes) != 1 {
		t.Fatalf("GetByID(preload).Notes len = %d, want 1", len(withNotes.Notes))
	}
	if len(withNotes.Sections) != 1 || withNotes.Sections[0].SectionName != "Sorting" {
		t.Fatalf("GetByID().Sections = %+v, want Sorting section", withNotes.Sections)
	}
	if _, err := service.Get(&courseman.Course{Title: "Missing"}); !errors.Is(err, courseman.ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrNotFound", err)
	}
	if _, err := service.GetByID(999999); !errors.Is(err, courseman.ErrNotFound) {
		t.Fatalf("GetByID(missing) error = %v, want ErrNotFound", err)
	}

	if err := service.Delete(first.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.GetByID(first.ID); !errors.Is(err, courseman.ErrNotFound) {
		t.Fatalf("GetByID(deleted) error = %v, want ErrNotFound", err)
	}
}

func TestCourseServiceCount(t *testing.T) {
	db := testdb.Open(t)
	service := courseman.NewService(db, testdb.DiscardLogger(), testdb.DiscardLogger())

	if _, err := service.Save(&courseman.Course{UserID: 1, Title: "Calc I"}); err != nil {
		t.Fatalf("Save(calc1): %v", err)
	}
	if _, err := service.Save(&courseman.Course{UserID: 2, Title: "Calc II"}); err != nil {
		t.Fatalf("Save(calc2): %v", err)
	}

	total, err := service.Count(nil)
	if err != nil {
		t.Fatalf("Count(nil) error = %v", err)
	}
	if total != 2 {
		t.Fatalf("Count(nil) = %d, want 2", total)
	}

	total, err = service.Count(&courseman.Filter{UserID: 1})
	if err != nil {
		t.Fatalf("Count(UserID) error = %v", err)
	}
	if total != 1 {
		t.Fatalf("Count(UserID=1) = %d, want 1", total)
	}

	total, err = service.Count(&courseman.Filter{Keyword: "Calc"})
	if err != nil {
		t.Fatalf("Count(Keyword) error = %v", err)
	}
	if total != 2 {
		t.Fatalf("Count(Keyword) = %d, want 2", total)
	}
}

func assertCourseFilterCount(t *testing.T, service *courseman.Service, filter *courseman.Filter, want int) {
	t.Helper()

	courses, total, err := service.GetAll(filter, 1, 25)
	if err != nil {
		t.Fatalf("GetAll(%+v) error = %v", filter, err)
	}
	if total != want || len(courses) != want {
		t.Fatalf("GetAll(%+v) returned len=%d total=%d, want %d", filter, len(courses), total, want)
	}
}
