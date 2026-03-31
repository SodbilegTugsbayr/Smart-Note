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
	UserID      int        `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"`
	HasBook     bool       `json:"has_book"`
	FilePath    string     `json:"-"`
	Icon        string     `json:"icon"`
	Sections    []*Section `gorm:"serializer:json;type:jsonb" json:"sections,omitempty"`
}

type Section struct {
	SectionName string `json:"section_name"`
	StartPage   int    `json:"start_page"`
	EndPage     int    `json:"end_page"`
}
