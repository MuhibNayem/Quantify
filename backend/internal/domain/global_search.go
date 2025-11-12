package domain

import (
	"time"
)

type GlobalSearchEntry struct {
	ID           uint   `gorm:"primaryKey"`
	EntityType   string `gorm:"uniqueIndex:idx_entity"`
	EntityID     uint   `gorm:"uniqueIndex:idx_entity"`
	Content      string
	SearchVector string    `gorm:"type:tsvector;index"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
