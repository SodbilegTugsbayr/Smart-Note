package validators

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

func ValidateUser(user *userman.User) error {
	user.FirstName = strings.TrimSpace(user.FirstName)
	user.PhoneNumber = strings.TrimSpace(user.PhoneNumber)
	user.Email = strings.TrimSpace(user.Email)

	if user.Email == "" {
		return errors.New("email_is_required")
	}

	if _, err := mail.ParseAddress(user.Email); err != nil {
		return err
	}

	if utf8.RuneCountInString(user.FirstName) > 50 || utf8.RuneCountInString(user.LastName) > 50 ||
		utf8.RuneCountInString(user.PhoneNumber) > 50 ||
		utf8.RuneCountInString(user.Email) > 255 {
		return errors.New("invalid_user_info")
	}

	return nil
}
