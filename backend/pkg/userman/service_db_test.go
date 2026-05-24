package userman_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/internal/testdb"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

func TestUserServiceDatabaseCRUDAndFilters(t *testing.T) {
	db := testdb.Open(t)
	service := userman.NewService(db, testdb.DiscardLogger(), testdb.DiscardLogger())

	alice, err := service.Save(&userman.User{
		FirstName:  "Alice",
		LastName:   "Student",
		Email:      "alice@example.com",
		AuthType:   userman.AUTH_TYPE_BASIC,
		Role:       userman.ROLE_USER,
		IsVerified: true,
	})
	if err != nil {
		t.Fatalf("Save(alice) error = %v", err)
	}
	bob, err := service.Save(&userman.User{
		FirstName: "Bob",
		LastName:  "Admin",
		Email:     "bob@example.com",
		AuthType:  "external",
		Role:      userman.ROLE_ADMIN,
	})
	if err != nil {
		t.Fatalf("Save(bob) error = %v", err)
	}

	got, err := service.GetByID(alice.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Email != alice.Email {
		t.Fatalf("GetByID().Email = %q, want %q", got.Email, alice.Email)
	}

	got, err = service.Get(&userman.User{Email: alice.Email})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != alice.ID {
		t.Fatalf("Get().ID = %d, want %d", got.ID, alice.ID)
	}

	if _, err := service.GetWithAuthTypes(&userman.User{Email: alice.Email}, []string{"external"}); !errors.Is(err, userman.ErrNotFound) {
		t.Fatalf("GetWithAuthTypes() error = %v, want ErrNotFound", err)
	}
	if _, err := service.Get(&userman.User{Email: "missing@example.com"}); !errors.Is(err, userman.ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrNotFound", err)
	}
	if _, err := service.GetByID(999999); !errors.Is(err, userman.ErrNotFound) {
		t.Fatalf("GetByID(missing) error = %v, want ErrNotFound", err)
	}

	total, err := service.Count(nil)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("Count() = %d, want 2", total)
	}

	assertUserFilterCount(t, service, &userman.Filter{Keyword: "lic"}, 1)
	assertUserFilterCount(t, service, &userman.Filter{Role: userman.ROLE_ADMIN}, 1)
	assertUserFilterCount(t, service, &userman.Filter{Email: "alice@"}, 1)
	assertUserFilterCount(t, service, &userman.Filter{Emails: []string{alice.Email, bob.Email}}, 2)
	assertUserFilterCount(t, service, &userman.Filter{IDs: []int{bob.ID}}, 1)

	users, total, err := service.GetAll(nil, 2, 1)
	if err != nil {
		t.Fatalf("GetAll(page 2) error = %v", err)
	}
	if total != 2 || len(users) != 1 {
		t.Fatalf("GetAll(page 2) returned len=%d total=%d, want one user on second page", len(users), total)
	}

	if err := service.Delete(alice.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.GetByID(alice.ID); !errors.Is(err, userman.ErrNotFound) {
		t.Fatalf("GetByID(deleted) error = %v, want ErrNotFound", err)
	}
}

func TestUserServiceGetWithAuthTypesHappyPath(t *testing.T) {
	db := testdb.Open(t)
	service := userman.NewService(db, testdb.DiscardLogger(), testdb.DiscardLogger())

	saved, err := service.Save(&userman.User{
		FirstName: "Carol",
		Email:     "carol@example.com",
		AuthType:  userman.AUTH_TYPE_BASIC,
		Role:      userman.ROLE_USER,
	})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := service.GetWithAuthTypes(&userman.User{Email: saved.Email}, []string{userman.AUTH_TYPE_BASIC})
	if err != nil {
		t.Fatalf("GetWithAuthTypes() error = %v", err)
	}
	if got.ID != saved.ID {
		t.Fatalf("GetWithAuthTypes().ID = %d, want %d", got.ID, saved.ID)
	}
}

func TestUserServiceGetRecentlyDeleted(t *testing.T) {
	db := testdb.Open(t)
	service := userman.NewService(db, testdb.DiscardLogger(), testdb.DiscardLogger())

	twoDaysAgo := time.Now().AddDate(0, 0, -2)
	deleted, err := service.Save(&userman.User{
		FirstName:     "Dan",
		Email:         "dan@example.com",
		AuthType:      userman.AUTH_TYPE_BASIC,
		Role:          userman.ROLE_USER,
		SelfDeletedAt: sql.NullTime{Time: twoDaysAgo, Valid: true},
	})
	if err != nil {
		t.Fatalf("Save(deleted) error = %v", err)
	}

	got, err := service.GetRecentlyDeleted(&userman.User{Email: deleted.Email}, []string{userman.AUTH_TYPE_BASIC})
	if err != nil {
		t.Fatalf("GetRecentlyDeleted() error = %v", err)
	}
	if got.ID != deleted.ID {
		t.Fatalf("GetRecentlyDeleted().ID = %d, want %d", got.ID, deleted.ID)
	}

	if _, err := service.GetRecentlyDeleted(&userman.User{Email: "absent@example.com"}, []string{userman.AUTH_TYPE_BASIC}); !errors.Is(err, userman.ErrNotFound) {
		t.Fatalf("GetRecentlyDeleted(missing) error = %v, want ErrNotFound", err)
	}
}

func assertUserFilterCount(t *testing.T, service *userman.Service, filter *userman.Filter, want int) {
	t.Helper()

	users, total, err := service.GetAll(filter, 1, 25)
	if err != nil {
		t.Fatalf("GetAll(%+v) error = %v", filter, err)
	}
	if total != want || len(users) != want {
		t.Fatalf("GetAll(%+v) returned len=%d total=%d, want %d", filter, len(users), total, want)
	}
}
