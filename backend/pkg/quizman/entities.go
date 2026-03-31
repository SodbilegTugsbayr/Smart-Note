package quizman

import (
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/entities"
)

const (
	STATUS_IN_PROGRESS = "in_progress"
	STATUS_COMPLETED   = "completed"
)

type Quiz struct {
	entities.Model
	NoteID        int      `json:"course_id"`
	Question      string   `json:"question"`
	Options       []string `json:"options" gorm:"serializer:json"`
	CorrectAnswer string   `json:"-"`
}
