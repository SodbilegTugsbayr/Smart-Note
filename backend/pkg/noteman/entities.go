package noteman

import (
	"strings"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/entities"
	"gorm.io/gorm"
)

const (
	STATUS_IN_PROGRESS = "in_progress"
	STATUS_COMPLETED   = "completed"

	PROCESS_STATUS_PROCESSING = "processing"
	PROCESS_STATUS_COMPLETED  = "completed"
	PROCESS_STATUS_FAILED     = "failed"
)

type Note struct {
	entities.Model
	CourseID      int           `json:"course_id"`
	Title         string        `json:"title"`
	IsFromBook    bool          `json:"is_from_book"`
	StartPage     int           `json:"start_page"`
	EndPage       int           `json:"end_page"`
	FilePath      string        `json:"-"`
	HasFile       bool          `gorm:"-" json:"has_file"`
	Summary       string        `json:"summary"`
	Status        string        `json:"status"`
	ProcessStatus string        `json:"process_status"`
	RawContent    string        `json:"raw_content"`
	KeyConcepts   []*KeyConcept `gorm:"serializer:json;type:jsonb" json:"key_concepts,omitempty"`
	FlashCards    []*FlashCard  `gorm:"serializer:json;type:jsonb" json:"flash_cards,omitempty"`
}

func (n *Note) AfterFind(tx *gorm.DB) error {
	n.HasFile = strings.TrimSpace(n.FilePath) != ""
	return nil
}

func (n *Note) PrepareResponse() {
	n.HasFile = strings.TrimSpace(n.FilePath) != ""
}

type KeyConcept struct {
	Concept    string `json:"concept"`
	Definition string `json:"definition"`
}

type FlashCard struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}
