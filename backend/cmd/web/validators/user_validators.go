package validators

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/peakdot/go-nuxt-example/backend/pkg/userman"
)

func ValidateUser(user *userman.User) error {
	user.Name = strings.TrimSpace(user.Name)
	user.PhoneNumber = strings.TrimSpace(user.PhoneNumber)
	user.Email = strings.TrimSpace(user.Email)

	if user.Email == "" {
		return errors.New("email is required")
	}

	if _, err := mail.ParseAddress(user.Email); err != nil {
		return err
	}

	if utf8.RuneCountInString(user.Name) > 50 ||
		utf8.RuneCountInString(user.PhoneNumber) > 50 ||
		utf8.RuneCountInString(user.Email) > 255 {
		return errors.New("invalid user info")
	}

	return nil
}
