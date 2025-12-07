package domain

import (
	"time"

	"gorm.io/gorm"
)

// Order represents a sales order.
type Order struct {
	gorm.Model
	OrderNumber   string `gorm:"uniqueIndex;not null"`
	UserID        uint   `gorm:"index"` // Customer ID, nullable if guest checkout supported (but here we enforce user)
	User          User
	TotalAmount   float64   `gorm:"not null"`
	Status        string    `gorm:"default:'COMPLETED';index"` // PENDING, COMPLETED, CANCELLED, RETURNED
	PaymentMethod string    `gorm:"not null"`
	OrderDate     time.Time `gorm:"index"`
	OrderItems    []OrderItem
}

// OrderItem represents an item within an order.
type OrderItem struct {
	gorm.Model
	OrderID     uint `gorm:"not null;index"`
	Order       Order
	ProductID   uint `gorm:"not null;index"`
	Product     Product
	Quantity    int     `gorm:"not null"`
	UnitPrice   float64 `gorm:"not null"`
	TotalPrice  float64 `gorm:"not null"`
	IsReturned  bool    `gorm:"default:false"`
	ReturnedQty int     `gorm:"default:0"`
}

// Return represents a return request or processed return.
type Return struct {
	gorm.Model
	OrderID      uint `gorm:"not null;index"`
	Order        Order
	UserID       uint `gorm:"not null;index"` // User who requested the return (Customer)
	User         User
	Status       string  `gorm:"default:'PENDING';index"` // PENDING, APPROVED, REJECTED, COMPLETED
	Reason       string  `gorm:"not null"`
	RefundAmount float64 `gorm:"not null"`
	ApprovedBy   *uint   // UserID of the approver (Staff/Manager)
	ApprovedAt   *time.Time
	ReturnItems  []ReturnItem
}

// ReturnItem represents an item within a return.
type ReturnItem struct {
	gorm.Model
	ReturnID    uint `gorm:"not null;index"`
	Return      Return
	OrderItemID uint `gorm:"not null;index"`
	OrderItem   OrderItem
	ProductID   uint `gorm:"not null"`
	Product     Product
	Quantity    int    `gorm:"not null"`
	Condition   string `gorm:"default:'GOOD'"` // GOOD, DAMAGED, OPENED
	Reason      string `gorm:"not null"`
}
