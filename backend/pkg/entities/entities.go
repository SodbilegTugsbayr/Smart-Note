package entities

// #region Import
import (
	"time"

	"gorm.io/gorm"
)

// #endregion Import

type Model struct {
	ID        int            `gorm:"primary_key" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
