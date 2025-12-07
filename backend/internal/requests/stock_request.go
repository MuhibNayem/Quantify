package requests

import "time"

// StockInBatchRequest represents the request body for adding new stock with batch information.
type StockInBatchRequest struct {
	Quantity    int        `json:"quantity" binding:"required,gt=0"`
	BatchNumber string     `json:"batchNumber" binding:"required"`
	ExpiryDate  *time.Time `json:"expiryDate"` // Optional for non-perishable
}

// StockAdjustmentRequest represents the request body for performing a manual stock adjustment.
type StockAdjustmentRequest struct {
	Type       string `json:"type" binding:"required,oneof=STOCK_IN STOCK_OUT"` // "STOCK_IN" or "STOCK_OUT"
	Quantity   int    `json:"quantity" binding:"required,gt=0"`
	ReasonCode string `json:"reasonCode" binding:"required"`
	Notes      string `json:"notes"`
}
type CheckoutItem struct {
	ProductID uint `json:"productId" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type CheckoutRequest struct {
	Items         []CheckoutItem `json:"items" binding:"required,dive"`
	CustomerID    *uint          `json:"customerId"`
	PaymentMethod string         `json:"paymentMethod" binding:"required"`
}
