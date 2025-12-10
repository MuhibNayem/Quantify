package domain

import (
	"time"

	"gorm.io/gorm"
)

// CashDrawerSession represents a period of time a cash drawer is open/active.
type CashDrawerSession struct {
	gorm.Model
	UserID       uint `gorm:"not null;index"` // The cashier responsible
	User         User
	LocationID   uint `gorm:"not null;index"`
	Location     Location
	StartTime    time.Time `gorm:"not null"`
	EndTime      *time.Time
	StartingCash float64 `gorm:"not null"`
	EndingCash   float64 // Actual cash counted by cashier at closing
	SystemCash   float64 // Expected cash calculated by system (Starting + Sales - Returns - Drops)
	Variance     float64 // EndingCash - SystemCash
	Notes        string
	Status       string `gorm:"default:'OPEN'"` // OPEN, CLOSED
	TotalSales   float64
	TotalRefunds float64
	TotalDrops   float64 // Cash removed during the shift (e.g., safe drops)
}

// CashDrop represents a removal of cash from the drawer during a session.
type CashDrop struct {
	gorm.Model
	SessionID uint `gorm:"not null;index"`
	Session   CashDrawerSession
	Amount    float64 `gorm:"not null"`
	Reason    string
	DroppedBy uint `gorm:"not null"`
	DroppedAt time.Time
}
