package userman_test

import (
	"errors"
	"testing"

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

	if err := service.Delete(alice.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.GetByID(alice.ID); !errors.Is(err, userman.ErrNotFound) {
		t.Fatalf("GetByID(deleted) error = %v, want ErrNotFound", err)
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
