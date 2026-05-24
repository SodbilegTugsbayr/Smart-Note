package validators

import (
	"strings"
	"testing"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

func TestValidateUserTrimsAndAcceptsValidUser(t *testing.T) {
	user := &userman.User{
		FirstName:   " Alice ",
		PhoneNumber: " 99112233 ",
		Email:       " alice@example.com ",
	}

	if err := ValidateUser(user); err != nil {
		t.Fatalf("ValidateUser() error = %v", err)
	}
	if user.FirstName != "Alice" || user.PhoneNumber != "99112233" || user.Email != "alice@example.com" {
		t.Fatalf("ValidateUser() did not trim fields: %+v", user)
	}
}

func TestValidateUserRejectsMissingEmail(t *testing.T) {
	err := ValidateUser(&userman.User{})
	if err == nil || err.Error() != "email_is_required" {
		t.Fatalf("ValidateUser() error = %v, want email_is_required", err)
	}
}

func TestValidateUserRejectsInvalidEmail(t *testing.T) {
	if err := ValidateUser(&userman.User{Email: "not an email"}); err == nil {
		t.Fatal("ValidateUser() error = nil, want invalid email error")
	}
}

func TestValidateUserRejectsTooLongFields(t *testing.T) {
	err := ValidateUser(&userman.User{
		FirstName: strings.Repeat("a", 51),
		Email:     "alice@example.com",
	})
	if err == nil || err.Error() != "invalid_user_info" {
		t.Fatalf("ValidateUser() error = %v, want invalid_user_info", err)
	}
}
