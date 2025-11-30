package userman

import (
	"database/sql"
	"time"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/entities"
)

const (
	ROLE_BASIC = "basic"
	ROLE_ADMIN = "admin"

	AUTH_TYPE_BASIC    = "username_password"
	AUTH_TYPE_FACEBOOK = "facebook"
	AUTH_TYPE_GOOGLE   = "google"
)

type User struct {
	entities.Model
	UUID           string       `json:"uuid"`
	AuthType       string       `json:"auth_type"`
	Role           string       `json:"role"` // @TO-DO
	PasswordHash   string       `json:"-"`
	Name           string       `json:"name"`
	PhoneNumber    string       `json:"phone_number"`
	Location       string       `json:"location"`
	FacebookURL    string       `json:"facebook_url"`
	MessengerURL   string       `json:"messenger_url"`
	Email          string       `json:"email"`
	GoogleID       string       `json:"google_id"`
	FacebookID     string       `json:"facebook_id"`
	ProfilePicture string       `json:"profile_picture"`
	PlanID         int          `json:"plan_id"` // @TO-DO
	IsVerified     bool         `json:"is_verified"`
	LastLogin      time.Time    `json:"last_login"`
	SelfDeletedAt  sql.NullTime `json:"self_deleted_at,omitempty"`
}
