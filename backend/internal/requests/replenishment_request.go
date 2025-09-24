package requests

import "time"

// ForecastGenerationRequest represents the request body for triggering demand forecast generation.
type ForecastGenerationRequest struct {
	PeriodInDays int  `json:"periodInDays" binding:"required,gt=0"`
	ProductID    *uint `json:"productId"` // Optional, for specific product forecast
}

// CreatePOFromSuggestionRequest represents the request body for creating a PO from a suggestion.
type CreatePOFromSuggestionRequest struct {
	SuggestionID uint `json:"suggestionId" binding:"required"`
}

// ApprovePORequest represents the request body for approving a purchase order.
type ApprovePORequest struct {
	POID uint `json:"poId" binding:"required"`
}

// UpdatePORequest represents the request body for updating a purchase order.
type UpdatePORequest struct {
	SupplierID         uint       `json:"supplierId,omitempty"`	
	Status             string     `json:"status,omitempty" binding:"omitempty,oneof=DRAFT APPROVED SENT RECEIVED CANCELLED"`
	OrderDate          *time.Time `json:"orderDate,omitempty"`
	ExpectedDeliveryDate *time.Time `json:"expectedDeliveryDate,omitempty"`
	PurchaseOrderItems []POItemRequest `json:"items,omitempty"`
}

// POItemRequest represents an item within a purchase order update request.
type POItemRequest struct {
	ProductID       uint    `json:"productId" binding:"required"`	
	OrderedQuantity int     `json:"orderedQuantity" binding:"required,gt=0"`
	UnitPrice       float64 `json:"unitPrice" binding:"required,gt=0"`
}

// ReceivePORequest represents the request body for receiving goods for a purchase order.
type ReceivePORequest struct {
	Items []ReceivePOItemRequest `json:"items" binding:"required,min=1"`
}

// ReceivePOItemRequest represents an item being received for a purchase order.
type ReceivePOItemRequest struct {
	PurchaseOrderItemID uint `json:"purchaseOrderItemId" binding:"required"`
	ReceivedQuantity    int  `json:"receivedQuantity" binding:"required,gt=0"`
	BatchNumber         string `json:"batchNumber" binding:"required"`
	ExpiryDate          *time.Time `json:"expiryDate"`
}
