package courseman

import (
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/entities"
)

const (
	STATUS_IN_PROGRESS = "in_progress"
	STATUS_COMPLETED   = "completed"
)

type Course struct {
	entities.Model
	UserID      int    `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Progress    int    `json:"progress"`
	FilePath    string `json:"file_url"`
	Icon        string `json:"icon"`
}
