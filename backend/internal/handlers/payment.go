package handlers

import (
	"net/http"

	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService services.PaymentService
}

func NewPaymentHandler(paymentService services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// CreatePayment godoc
// @Summary Create a payment session
// @Description Initiates a payment session for an order
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body requests.PaymentRequest true "Payment request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /payment/create [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req requests.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid payment request payload", http.StatusBadRequest, err))
		return
	}

	// In a real application, you would get the user details from the context
	// For now, we will use dummy data
	cusName := "Test Customer"
	cusEmail := "test@example.com"
	cusPhone := "123456789"

	redirectURL, err := h.paymentService.CreatePayment(req.PaymentMethod, req.Amount, req.OrderID, "BDT", cusName, cusEmail, cusPhone)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to create payment session", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"redirect_url": redirectURL})
}

// HandleSSLCommerzIPN godoc
// @Summary Handle SSLCommerz IPN
// @Description Processes SSLCommerz Instant Payment Notification
// @Tags payments
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Router /payment/ipn [post]
func (h *PaymentHandler) HandleSSLCommerzIPN(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.Error(appErrors.NewAppError("Invalid SSLCommerz IPN payload", http.StatusBadRequest, err))
		return
	}

	if err := h.paymentService.HandleSSLCommerzIPN(c.Request.PostForm); err != nil {
		c.Error(appErrors.NewAppError("SSLCommerz IPN validation failed", http.StatusBadRequest, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "SSLCommerz IPN handled successfully"})
}

// HandleBkashCallback godoc
// @Summary Handle bKash callback
// @Description Processes bKash payment callback
// @Tags payments
// @Produce json
// @Param paymentID query string true "Payment ID"
// @Param status query string true "Payment status"
// @Success 302 {string} string "Redirect"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /payment/bkash/callback [get]
func (h *PaymentHandler) HandleBkashCallback(c *gin.Context) {
	paymentID := c.Query("paymentID")
	status := c.Query("status")

	if paymentID == "" || status == "" {
		c.Error(appErrors.NewAppError("Missing paymentID or status in bKash callback", http.StatusBadRequest, nil))
		return
	}

	redirectURL, err := h.paymentService.HandleBkashCallback(paymentID, status)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to handle bKash callback", http.StatusInternalServerError, err))
		return
	}

	c.Redirect(http.StatusFound, redirectURL)
}

// PaymentSuccess godoc
// @Summary Payment success callback
// @Description Handles successful payment confirmation
// @Tags payments
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /payment/success [post]
func (h *PaymentHandler) PaymentSuccess(c *gin.Context) {
	// Here you would typically redirect the user to a success page on your frontend
	c.JSON(http.StatusOK, gin.H{"status": "Payment successful"})
}

// PaymentFail godoc
// @Summary Payment failure callback
// @Description Handles failed payment
// @Tags payments
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /payment/fail [post]
func (h *PaymentHandler) PaymentFail(c *gin.Context) {
	// Here you would typically redirect the user to a failure page on your frontend
	c.JSON(http.StatusOK, gin.H{"status": "Payment failed"})
}

// PaymentCancel godoc
// @Summary Payment cancellation callback
// @Description Handles cancelled payment
// @Tags payments
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /payment/cancel [post]
func (h *PaymentHandler) PaymentCancel(c *gin.Context) {
	// Here you would typically redirect the user to a cancellation page on your frontend
	c.JSON(http.StatusOK, gin.H{"status": "Payment cancelled"})
}
