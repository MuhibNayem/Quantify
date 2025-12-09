package domain

import (
	"time"

	"gorm.io/gorm"
)

// AuditLog tracks sensitive actions within the system for compliance and security.
type AuditLog struct {
	gorm.Model
	Action    string `gorm:"not null;index"` // e.g., "CREATE", "UPDATE", "DELETE", "VOID", "DISCOUNT"
	Entity    string `gorm:"not null;index"` // e.g., "Product", "Order", "User", "Settings"
	EntityID  string `gorm:"not null;index"` // ID of the affected entity
	UserID    uint   `gorm:"not null;index"` // Who performed the action
	User      User
	Changes   string // JSON string describing the changes (old vs new values)
	IPAddress string // Optional: IP address of the user
	UserAgent string // Optional: User agent string
	Timestamp time.Time `gorm:"index"`
}
