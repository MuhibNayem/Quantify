package requests

import "time"

// StockInBatchRequest represents the request body for adding new stock with batch information.
type StockInBatchRequest struct {
	Quantity    float64    `json:"quantity" binding:"required,gt=0"`
	BatchNumber string     `json:"batchNumber" binding:"required"`
	ExpiryDate  *time.Time `json:"expiryDate"` // Optional for non-perishable
}

// StockAdjustmentRequest represents the request body for performing a manual stock adjustment.
type StockAdjustmentRequest struct {
	Type       string  `json:"type" binding:"required,oneof=STOCK_IN STOCK_OUT"` // "STOCK_IN" or "STOCK_OUT"
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
	ReasonCode string  `json:"reasonCode" binding:"required"`
	Notes      string  `json:"notes"`
}
type CheckoutItem struct {
	ProductID uint    `json:"productId" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	UnitID    *uint   `json:"unitId"` // selling unit; nil/omitted = product base unit
}

type CheckoutRequest struct {
	Items              []CheckoutItem `json:"items" binding:"required,dive"`
	CustomerID         *uint          `json:"customerId"`
	PaymentMethod      string         `json:"paymentMethod" binding:"required"`
	PointsToRedeem     int            `json:"pointsToRedeem"`    // Optional points to redeem
	RegisterLocationID *uint          `json:"registerLocationId" binding:"omitempty"` // Register-specific location override for stock deduction
}
