package handlers

import (
	"net/http"

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

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req requests.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real application, you would get the user details from the context
	// For now, we will use dummy data
	cusName := "Test Customer"
	cusEmail := "test@example.com"
	cusPhone := "123456789"

	redirectURL, err := h.paymentService.CreatePayment(req.PaymentMethod, req.Amount, req.OrderID, "BDT", cusName, cusEmail, cusPhone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment session: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"redirect_url": redirectURL})
}

func (h *PaymentHandler) HandleSSLCommerzIPN(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse form"})
		return
	}

	valid, err := h.paymentService.ValidateSSLCommerzIPN(c.Request.Form)
	if err != nil || !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SSLCommerz IPN validation failed"})
		return
	}

	// IPN is valid, now update the transaction status
	// In a real application, you would have a repository method to update the transaction
	// Example of getting a value: c.Request.Form.Get("tran_id")

	// You would typically update your database here based on the IPN data
	// e.g., find the transaction by tran_id and update its status

	c.JSON(http.StatusOK, gin.H{"status": "SSLCommerz IPN handled successfully"})
}

func (h *PaymentHandler) HandleBkashCallback(c *gin.Context) {
	paymentID := c.Query("paymentID")
	status := c.Query("status")

	if paymentID == "" || status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing paymentID or status in bKash callback"})
		return
	}

	redirectURL, err := h.paymentService.HandleBkashCallback(paymentID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle bKash callback: " + err.Error()})
		return
	}

	c.Redirect(http.StatusFound, redirectURL)
}

func (h *PaymentHandler) PaymentSuccess(c *gin.Context) {
	// Here you would typically redirect the user to a success page on your frontend
	c.JSON(http.StatusOK, gin.H{"status": "Payment successful"})
}

func (h *PaymentHandler) PaymentFail(c *gin.Context) {
	// Here you would typically redirect the user to a failure page on your frontend
	c.JSON(http.StatusOK, gin.H{"status": "Payment failed"})
}

func (h *PaymentHandler) PaymentCancel(c *gin.Context) {
	// Here you would typically redirect the user to a cancellation page on your frontend
	c.JSON(http.StatusOK, gin.H{"status": "Payment cancelled"})
}
