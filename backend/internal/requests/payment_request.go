package requests

// PaymentRequest represents the request body for creating a payment.
type PaymentRequest struct {
	OrderID       string  `json:"order_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=bkash card cash"`
}
