package domain

import (
	"gorm.io/gorm"
)

// SystemSetting represents a configurable system setting.
type SystemSetting struct {
	gorm.Model
	Key         string `gorm:"uniqueIndex;not null"`
	Value       string `gorm:"type:text"` // Store value as string/JSON
	Group       string `gorm:"index"`     // e.g., "General", "Business", "Policy"
	Type        string `gorm:"not null"`  // e.g., "string", "boolean", "number", "long_text"
	Description string
}
