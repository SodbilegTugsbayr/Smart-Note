package courseman

import (
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/entities"
)

const (
	ROLE_BASIC = "basic"
	ROLE_ADMIN = "admin"

	AUTH_TYPE_BASIC    = "username_password"
	AUTH_TYPE_FACEBOOK = "facebook"
	AUTH_TYPE_GOOGLE   = "google"
)

type Course struct {
	entities.Model
	UserID   int    `json:"user_id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	FilePath string `json:"file_url"`
}
