package domain

import (
	"time"

	"gorm.io/gorm"
)

// Product represents a product in the inventory.
type Product struct {
	gorm.Model
	SKU           string `gorm:"uniqueIndex;not null"`
	Name          string `gorm:"not null"`
	Description   string
	CategoryID    uint
	Category      Category
	SubCategoryID uint
	SubCategory   SubCategory
	SupplierID    uint
	Supplier      Supplier
	Brand         string
	PurchasePrice float64
	SellingPrice  float64
	BarcodeUPC    string `gorm:"uniqueIndex"`
	ImageURLs     string // Storing as comma-separated string or JSON string for simplicity
	Status        string `gorm:"default:'Active'"` // Active, Archived, Discontinued
	LocationID    uint
	Location      Location
}

// Category represents a product category.
type Category struct {
	gorm.Model
	Name          string `gorm:"uniqueIndex;not null"`
	SubCategories []SubCategory
}

// SubCategory represents a product sub-category.
type SubCategory struct {
	gorm.Model
	Name       string `gorm:"not null"`
	CategoryID uint   `gorm:"not null"`
	Category   Category
}

// Supplier represents a product supplier.
type Supplier struct {
	gorm.Model
	Name          string `gorm:"not null"`
	ContactPerson string
	Email         string
	Phone         string
	Address       string
}

// Location represents a physical inventory location (e.g., warehouse, store).
type Location struct {
	gorm.Model
	Name    string `gorm:"uniqueIndex;not null"`
	Address string
}

// Batch represents a batch of stock for a product.
type Batch struct {
	gorm.Model
	ProductID   uint `gorm:"not null"`
	Product     Product
	LocationID  uint
	Location    Location
	BatchNumber string     `gorm:"not null"`
	Quantity    int        `gorm:"not null"`
	ExpiryDate  *time.Time // Pointer to allow null for non-perishable
}

// StockAdjustment represents a manual adjustment to stock levels.
type StockAdjustment struct {
	gorm.Model
	ProductID        uint `gorm:"not null"`
	Product          Product
	LocationID       uint
	Location         Location
	Type             string `gorm:"not null"` // e.g., "STOCK_IN", "STOCK_OUT"
	Quantity         int    `gorm:"not null"`
	ReasonCode       string `gorm:"not null"` // e.g., "DAMAGED_GOODS", "STOCK_TAKE_CORRECTION"
	Notes            string
	AdjustedBy       uint // UserID of the person who made the adjustment
	AdjustedAt       time.Time
	PreviousQuantity int // Snapshot of quantity before adjustment
	NewQuantity      int // Snapshot of quantity after adjustment
}

// Alert represents a stock-related alert.
type Alert struct {
	gorm.Model
	ProductID   uint `gorm:"not null"`
	Product     Product
	Type        string `gorm:"not null"` // e.g., "LOW_STOCK", "OUT_OF_STOCK", "OVERSTOCK", "EXPIRY_ALERT"
	Message     string `gorm:"not null"`
	TriggeredAt time.Time
	Status      string `gorm:"default:'ACTIVE'"` // ACTIVE, RESOLVED
	BatchID     *uint  // Optional, for expiry alerts
	Batch       *Batch
}

// User represents a system user (for AdjustedBy in StockAdjustment, etc.)
type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"not null"` // e.g., "Admin", "Manager", "Staff"
	IsActive bool   `gorm:"default:false"`
}

// ProductAlertSettings stores alert thresholds per product.
type ProductAlertSettings struct {
	gorm.Model
	ProductID       uint `gorm:"uniqueIndex;not null"`
	Product         Product
	LowStockLevel   int
	OverStockLevel  int
	ExpiryAlertDays int
}

// UserNotificationSettings stores user preferences for notifications.
type UserNotificationSettings struct {
	gorm.Model
	UserID                    uint `gorm:"uniqueIndex;not null"`
	User                      User
	EmailNotificationsEnabled bool
	SMSNotificationsEnabled   bool
	EmailAddress              string
	PhoneNumber               string
}

// DemandForecast represents a product's demand forecast.
type DemandForecast struct {
	gorm.Model
	ProductID       uint `gorm:"not null"`
	Product         Product
	ForecastPeriod  string `gorm:"not null"` // e.g., "30_DAYS", "90_DAYS"
	PredictedDemand int    `gorm:"not null"`
	GeneratedAt     time.Time
}

// ReorderSuggestion represents a suggestion to reorder a product.
type ReorderSuggestion struct {
	gorm.Model
	ProductID              uint `gorm:"not null"`
	Product                Product
	SupplierID             uint `gorm:"not null"`
	Supplier               Supplier
	CurrentStock           int
	PredictedDemand        int
	SuggestedOrderQuantity int `gorm:"not null"`
	LeadTimeDays           int
	Status                 string `gorm:"default:'PENDING'"` // PENDING, APPROVED, REJECTED, PO_CREATED
	SuggestedAt            time.Time
}

// PurchaseOrder represents a purchase order to a supplier.
type PurchaseOrder struct {
	gorm.Model
	SupplierID           uint `gorm:"not null"`
	Supplier             Supplier
	Status               string `gorm:"default:'DRAFT'"` // DRAFT, APPROVED, SENT, PARTIALLY_RECEIVED, RECEIVED, CANCELLED
	OrderDate            time.Time
	ExpectedDeliveryDate *time.Time
	ActualDeliveryDate   *time.Time
	CreatedBy            uint  // UserID of the creator
	ApprovedBy           *uint // UserID of the approver
	ApprovedAt           *time.Time
	PurchaseOrderItems   []PurchaseOrderItem `gorm:"foreignKey:PurchaseOrderID"`
}

type StockTransfer struct {
	gorm.Model
	ProductID        uint `gorm:"not null"`
	Product          Product
	SourceLocationID uint `gorm:"not null"`
	SourceLocation   Location
	DestLocationID   uint `gorm:"not null"`
	DestLocation     Location
	Quantity         int    `gorm:"not null"`
	Status           string `gorm:"default:'PENDING'"` // PENDING, COMPLETED, CANCELLED
	TransferredBy    uint   // UserID of the person who initiated the transfer
	TransferredAt    time.Time
}

type RefreshToken struct {
	gorm.Model
	UserID    uint `gorm:"not null"`
	User      User
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
}

// PurchaseOrderItem represents an item within a purchase order.
type PurchaseOrderItem struct {
	gorm.Model
	PurchaseOrderID  uint `gorm:"not null"`
	PurchaseOrder    PurchaseOrder
	ProductID        uint `gorm:"not null"`
	Product          Product
	OrderedQuantity  int     `gorm:"not null"`
	ReceivedQuantity int     `gorm:"default:0"` // Quantity actually received
	UnitPrice        float64 `gorm:"not null"`
}

// Transaction represents a payment transaction.
type Transaction struct {
	gorm.Model
	OrderID              string `gorm:"not null"`
	Amount               int64  `gorm:"not null"` // Amount in smallest currency unit (e.g., cents)
	Currency             string `gorm:"not null"`
	PaymentMethod        string `gorm:"not null"` // e.g., "card", "bkash"
	Status               string `gorm:"not null"` // e.g., "pending", "succeeded", "failed"
	GatewayTransactionID string `gorm:"uniqueIndex;not null"`
}
