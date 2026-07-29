package domain

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Product represents a product in the inventory.
type Product struct {
	gorm.Model
	SKU           string `gorm:"uniqueIndex;not null"`
	Name          string `gorm:"not null;index"` // Index for name searches
	Description   string
	CategoryID    uint `gorm:"index"` // Index for category filters
	Category      Category
	SubCategoryID uint `gorm:"index"` // Index for subcategory filters
	SubCategory   SubCategory
	SupplierID    uint `gorm:"index"` // Index for supplier filters
	Supplier      Supplier
	Brand         string `gorm:"index"` // Index for brand filters
	PurchasePrice float64
	SellingPrice  float64
	BarcodeUPC    string `gorm:"index"` // Index for barcode lookups
	ImageURLs     string // Storing as comma-separated string or JSON string for simplicity
	Status        string `gorm:"default:'Active';index"` // Index for status filters (Active, Archived, Discontinued)
	LocationID    uint   `gorm:"index"`                  // Index for location filters
	Location      Location
	// BaseUnit is the unit stock is tracked in. nil → countable "each".
	// SellingPrice/PurchasePrice are expressed PER BASE UNIT by default.
	BaseUnitID *uint `gorm:"index" json:"BaseUnitID"`
	BaseUnit   Unit  `json:"BaseUnit"`
	// AllowFractional marks measured goods (wire, fabric, fuel) that can be
	// sold in non-whole quantities (1.2 ft, 14 in). Countable goods stay whole.
	AllowFractional bool `gorm:"default:false" json:"AllowFractional"`
	// ProductUnits are the alternate units this product can be sold/purchased in.
	ProductUnits []ProductUnit `gorm:"foreignKey:ProductID" json:"ProductUnits"`
}

// GetID implements the Searchable interface for Product.
func (p *Product) GetID() uint {
	return p.ID
}

// GetSearchableContent implements the Searchable interface for Product.
func (p *Product) GetSearchableContent() string {
	return strings.Join([]string{p.Name, p.SKU, p.Description, p.Brand, p.BarcodeUPC}, " ")
}

// GetEntityType implements the Searchable interface for Product.
func (p *Product) GetEntityType() string {
	return "product"
}

// Category represents a product category.
type Category struct {
	gorm.Model
	Name          string `gorm:"uniqueIndex;not null"`
	SubCategories []SubCategory
}

// GetID implements the Searchable interface for Category.
func (c *Category) GetID() uint {
	return c.ID
}

// GetSearchableContent implements the Searchable interface for Category.
func (c *Category) GetSearchableContent() string {
	return c.Name
}

// GetEntityType implements the Searchable interface for Category.
func (c *Category) GetEntityType() string {
	return "category"
}

// SubCategory represents a product sub-category.
type SubCategory struct {
	gorm.Model
	Name       string `gorm:"not null;index"`
	CategoryID uint   `gorm:"not null;index"` // Index for category lookups
	Category   Category
}

// Supplier represents a product supplier.
type Supplier struct {
	gorm.Model
	Name          string `gorm:"not null;index"` // Index for name searches
	ContactPerson string
	Email         string `gorm:"index"` // Index for email lookups
	Phone         string `gorm:"index"` // Index for phone lookups
	Address       string
}

// GetID implements the Searchable interface for Supplier.
func (s *Supplier) GetID() uint {
	return s.ID
}

// GetSearchableContent implements the Searchable interface for Supplier.
func (s *Supplier) GetSearchableContent() string {
	return strings.Join([]string{s.Name, s.ContactPerson, s.Email, s.Phone}, " ")
}

// GetEntityType implements the Searchable interface for Supplier.
func (s *Supplier) GetEntityType() string {
	return "supplier"
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
	ProductID   uint `gorm:"not null;index:idx_product_location"`
	Product     Product
	LocationID  uint `gorm:"index:idx_product_location"`
	Location    Location
	BatchNumber string     `gorm:"not null"`
	Quantity    float64    `gorm:"not null"`
	ExpiryDate  *time.Time // Pointer to allow null for non-perishable
}

// StockAdjustment represents a manual adjustment to stock levels.
type StockAdjustment struct {
	gorm.Model
	ProductID        uint `gorm:"not null;index"`
	Product          Product
	LocationID       uint
	Location         Location
	Type             string `gorm:"not null"` // e.g., "STOCK_IN", "STOCK_OUT"
	Quantity         float64 `gorm:"not null"`
	ReasonCode       string  `gorm:"not null"` // e.g., "DAMAGED_GOODS", "STOCK_TAKE_CORRECTION"
	Notes            string
	AdjustedBy       uint      // UserID of the person who made the adjustment
	AdjustedAt       time.Time `gorm:"index"`
	PreviousQuantity float64   // Snapshot of quantity before adjustment
	NewQuantity      float64   // Snapshot of quantity after adjustment
}

// Alert represents a stock-related alert.
type Alert struct {
	gorm.Model
	ProductID   uint `gorm:"not null;index:idx_alert_product_status"` // Composite index for product + status queries
	Product     Product
	Type        string    `gorm:"not null;index"` // Index for filtering by alert type
	Message     string    `gorm:"not null"`
	TriggeredAt time.Time `gorm:"index"`                                           // Index for date range queries
	Status      string    `gorm:"default:'ACTIVE';index:idx_alert_product_status"` // ACTIVE, RESOLVED
	BatchID     *uint     // Optional, for expiry alerts
	Batch       *Batch
}

// User represents a system user (for AdjustedBy in StockAdjustment, etc.)
type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	// Role           string `gorm:"not null"` // Deprecated in favor of RoleID
	LegacyRole     string `gorm:"column:role"`              // Map to old column for migration
	RoleID         uint   `gorm:"index;not null;default:0"` // Foreign key to roles table
	Role           Role
	RoleName       string `json:"role" gorm:"-"` // Computed field for API compatibility
	IsActive       bool   `gorm:"default:false"`
	FirstName      string
	LastName       string
	Email          string `gorm:"uniqueIndex"`
	PhoneNumber    string `gorm:"uniqueIndex"`
	Address        string
	LoyaltyAccount *LoyaltyAccount `json:"loyalty,omitempty"`
}

// AfterFind populates the RoleName field for API compatibility.
func (u *User) AfterFind(tx *gorm.DB) (err error) {
	if u.Role.ID != 0 {
		u.RoleName = u.Role.Name
	} else {
		u.RoleName = ""
	}
	return
}

// GetID implements the Searchable interface for User.
func (u *User) GetID() uint {
	return u.ID
}

// GetSearchableContent implements the Searchable interface for User.
func (u *User) GetSearchableContent() string {
	return strings.Join([]string{u.Username, u.FirstName, u.LastName, u.Email, u.PhoneNumber}, " ")
}

// GetEntityType implements the Searchable interface for User.
func (u *User) GetEntityType() string {
	return "user"
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
	PredictedDemand float64 `gorm:"not null"`
	ConfidenceScore float64
	Reasoning       string
	GeneratedAt     time.Time
}

// ReorderSuggestion represents a suggestion to reorder a product.
type ReorderSuggestion struct {
	gorm.Model
	ProductID              uint `gorm:"not null"`
	Product                Product
	SupplierID             uint `gorm:"not null"`
	Supplier               Supplier
	CurrentStock           float64
	PredictedDemand        float64
	SuggestedOrderQuantity float64 `gorm:"not null"`
	LeadTimeDays           int
	SnapshotPrice          float64 `gorm:"type:decimal(12,2);default:0"` // Product.PurchasePrice at the time the suggestion was generated; used to detect price drift when the PO is created
	Status                 string  `gorm:"default:'PENDING'"`             // PENDING, APPROVED, REJECTED, PO_CREATED
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
	Quantity         float64 `gorm:"not null"`
	Status           string  `gorm:"default:'PENDING'"` // PENDING, COMPLETED, CANCELLED
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
	OrderedQuantity  float64 `gorm:"not null"`
	ReceivedQuantity float64 `gorm:"default:0"` // Quantity actually received
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

// LoyaltyAccount represents a customer's loyalty account.
type LoyaltyAccount struct {
	gorm.Model
	UserID uint `gorm:"uniqueIndex;not null"`
	User   User
	Points int    `gorm:"default:0"`
	Tier   string `gorm:"default:'Bronze'"` // e.g., Bronze, Silver, Gold
}

type Job struct {
	gorm.Model
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	Payload       string     `json:"payload"`
	Result        string     `json:"result"`
	LastError     string     `json:"lastError"`
	RetryCount    int        `json:"retryCount"`
	MaxRetries    int        `json:"maxRetries"`
	LastAttemptAt *time.Time `json:"lastAttemptAt"`
}

// Notification represents an in-app notification for a user.
type Notification struct {
	gorm.Model
	UserID      uint `gorm:"not null;index:idx_user_read"` // Composite index for user + read status
	User        User
	Type        string `gorm:"not null;index"` // Index for filtering by type
	Title       string `gorm:"not null"`
	Message     string `gorm:"not null"`
	Payload     string // JSON string for additional data (e.g., productID, orderID)
	IsRead      bool   `gorm:"default:false;index:idx_user_read"` // Part of composite index
	ReadAt      *time.Time
	TriggeredAt time.Time `gorm:"index"` // Index for date sorting
}

// AlertRoleSubscription links an alert type to a user role.
type AlertRoleSubscription struct {
	gorm.Model
	AlertType string `gorm:"not null;uniqueIndex:idx_alert_role"`
	Role      string `gorm:"not null;uniqueIndex:idx_alert_role"`
}

// UserNotificationSubscription tracks per-user opt-in/out for a specific alert
// type. When no row exists for (UserID, Type) the user is treated as opted-in;
// an explicit row with Enabled=false records an opt-out.
type UserNotificationSubscription struct {
	gorm.Model
	UserID  uint   `gorm:"not null;uniqueIndex:idx_user_type,priority:1" json:"userId"`
	Type    string `gorm:"not null;uniqueIndex:idx_user_type,priority:2" json:"type"`
	Enabled bool   `gorm:"not null" json:"enabled"`
}

// TimeClock represents an employee's time clock entry.
type TimeClock struct {
	gorm.Model
	UserID     uint `gorm:"not null"`
	User       User
	ClockIn    time.Time
	ClockOut   *time.Time
	BreakStart *time.Time
	BreakEnd   *time.Time
	Status     string `gorm:"default:'CLOCKED_IN'"` // CLOCKED_IN, ON_BREAK, CLOCKED_OUT
	Notes      string
}

// PurchaseReturn represents a return of goods to a supplier.
type PurchaseReturn struct {
	gorm.Model
	PurchaseOrderID     uint `gorm:"index"` // Optional link to original PO
	PurchaseOrder       PurchaseOrder
	SupplierID          uint `gorm:"not null;index"`
	Supplier            Supplier
	Status              string `gorm:"default:'PENDING'"` // PENDING, COMPLETED, CANCELLED
	Reason              string
	RefundAmount        float64 // Estimated refund value
	ReturnedBy          uint    // UserID
	ReturnedAt          time.Time
	ApprovedBy          *uint      // UserID of the admin who approved
	ApprovedAt          *time.Time // Timestamp of approval intent
	CompletedAt         *time.Time // Timestamp of completion (stock deducted)
	CancelledAt         *time.Time // Timestamp of cancellation
	PurchaseReturnItems []PurchaseReturnItem
}

// PurchaseReturnItem represents an item within a purchase return.
type PurchaseReturnItem struct {
	gorm.Model
	PurchaseReturnID uint `gorm:"not null;index"`
	PurchaseReturn   PurchaseReturn
	ProductID        uint `gorm:"not null"`
	Product          Product
	Quantity         float64 `gorm:"not null"`
	Reason           string  `gorm:"not null"`
	BatchID          *uint   // Specific batch being returned (critical for stock deduction)
	Batch            *Batch
}

// ChurnRisk represents the AI-generated churn risk analysis for a customer.
type ChurnRisk struct {
	ChurnRiskScore    float64  `json:"churn_risk_score"`
	RiskLevel         string   `json:"risk_level"`
	PrimaryFactors    []string `json:"primary_factors"`
	RetentionStrategy string   `json:"retention_strategy"`
	SuggestedDiscount int      `json:"suggested_discount"`
}

// Promotion represents a discount rule.
type Promotion struct {
	gorm.Model
	Name          string `gorm:"not null"`
	Description   string
	DiscountType  string    `gorm:"not null"` // "PERCENTAGE" or "FIXED_AMOUNT"
	DiscountValue float64   `gorm:"not null"`
	StartDate     time.Time `gorm:"not null"`
	EndDate       time.Time `gorm:"not null"`
	IsActive      bool      `gorm:"default:true"`
	Priority      int       `gorm:"default:0"` // Higher number = higher priority

	// Targets (Nullable, only one should be set ideally, or hierarchical)
	ProductID     *uint `gorm:"index"`
	Product       *Product
	CategoryID    *uint `gorm:"index"`
	Category      *Category
	SubCategoryID *uint `gorm:"index"`
	SubCategory   *SubCategory
}
