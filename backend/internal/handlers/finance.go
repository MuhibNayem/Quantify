package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"inventory/backend/internal/services"
)

// FinanceHandler exposes the general ledger and accounts-receivable endpoints:
// trial balance, income statement, balance sheet, AR aging, outstanding AR, and
// payment receipt.
type FinanceHandler struct {
	Ledger *services.LedgerService
	AR     *services.AccountsReceivableService
	AP     *services.AccountsPayableService
}

func NewFinanceHandler(db *gorm.DB) *FinanceHandler {
	return &FinanceHandler{
		Ledger: services.NewLedgerService(db),
		AR:     services.NewAccountsReceivableService(db),
		AP:     services.NewAccountsPayableService(db),
	}
}

func parseRange(c *gin.Context) (*time.Time, *time.Time) {
	var from, to *time.Time
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = &t
		}
	}
	return from, to
}

// GetTrialBalance godoc
// @Summary      General ledger trial balance
// @Tags         finance
// @Produce      json
// @Param        from query string false "from date (RFC3339)"
// @Param        to query string false "to date (RFC3339)"
// @Success      200 {array} services.TrialBalanceRow
// @Security     BearerAuth
// @Router       /finance/trial-balance [get]
func (h *FinanceHandler) GetTrialBalance(c *gin.Context) {
	from, to := parseRange(c)
	rows, err := h.Ledger.TrialBalance(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": rows})
}

// GetIncomeStatement godoc
// @Summary      Income statement (P&L)
// @Tags         finance
// @Produce      json
// @Success      200 {object} services.IncomeStatement
// @Security     BearerAuth
// @Router       /finance/income-statement [get]
func (h *FinanceHandler) GetIncomeStatement(c *gin.Context) {
	from, to := parseRange(c)
	stmt, err := h.Ledger.IncomeStatement(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stmt)
}

// GetBalanceSheet godoc
// @Summary      Balance sheet
// @Tags         finance
// @Produce      json
// @Success      200 {object} services.BalanceSheet
// @Security     BearerAuth
// @Router       /finance/balance-sheet [get]
func (h *FinanceHandler) GetBalanceSheet(c *gin.Context) {
	from, to := parseRange(c)
	bs, err := h.Ledger.BalanceSheet(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bs)
}

// GetARAging godoc
// @Summary      Accounts receivable aging report
// @Tags         finance
// @Produce      json
// @Success      200 {array} services.ARAgingBucket
// @Security     BearerAuth
// @Router       /finance/ar-aging [get]
func (h *FinanceHandler) GetARAging(c *gin.Context) {
	buckets, err := h.AR.ARAgingReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"buckets": buckets})
}

// GetOutstandingAR godoc
// @Summary      List outstanding (unpaid) receivables
// @Tags         finance
// @Produce      json
// @Param        customerId query int false "filter by customer"
// @Success      200 {array} domain.AccountsReceivable
// @Security     BearerAuth
// @Router       /finance/receivables [get]
func (h *FinanceHandler) GetOutstandingAR(c *gin.Context) {
	var customerID *uint
	if v := c.Query("customerId"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			uid := uint(id)
			customerID = &uid
		}
	}
	ars, err := h.AR.ListOutstanding(customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"receivables": ars})
}

// RecordPaymentRequest is the body for recording a customer payment.
type RecordPaymentRequest struct {
	CustomerID  uint                                `json:"customerId" binding:"required"`
	Amount      float64                             `json:"amount" binding:"required,gt=0"`
	Method      string                              `json:"method" binding:"required"`
	Notes       string                              `json:"notes"`
	Allocations []services.PaymentAllocationInput   `json:"allocations"`
}

// RecordPayment godoc
// @Summary      Record a customer payment against receivables
// @Tags         finance
// @Accept       json
// @Produce      json
// @Param        request body RecordPaymentRequest true "payment"
// @Success      201 {object} domain.Payment
// @Security     BearerAuth
// @Router       /finance/payments [post]
func (h *FinanceHandler) RecordPayment(c *gin.Context) {
	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetUint("user_id")
	payment, err := h.AR.RecordPayment(services.RecordPaymentRequest{
		CustomerID:  req.CustomerID,
		Amount:      req.Amount,
		Method:      req.Method,
		Notes:       req.Notes,
		Allocations: req.Allocations,
	}, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, payment)
}

// GetAPAging godoc
// @Summary      Accounts payable aging report
// @Tags         finance
// @Produce      json
// @Success      200 {array} services.APAgingBucket
// @Security     BearerAuth
// @Router       /finance/ap-aging [get]
func (h *FinanceHandler) GetAPAging(c *gin.Context) {
	buckets, err := h.AP.APAgingReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"buckets": buckets})
}

// GetOutstandingAP godoc
// @Summary      List outstanding (unpaid) payables to suppliers
// @Tags         finance
// @Produce      json
// @Param        supplierId query int false "filter by supplier"
// @Success      200 {array} domain.AccountsPayable
// @Security     BearerAuth
// @Router       /finance/payables [get]
func (h *FinanceHandler) GetOutstandingAP(c *gin.Context) {
	var supplierID *uint
	if v := c.Query("supplierId"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			uid := uint(id)
			supplierID = &uid
		}
	}
	aps, err := h.AP.ListOutstanding(supplierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payables": aps})
}

// RecordSupplierPaymentRequest is the body for recording a payment to a supplier.
type RecordSupplierPaymentRequest struct {
	SupplierID  uint                                      `json:"supplierId" binding:"required"`
	Amount      float64                                   `json:"amount" binding:"required,gt=0"`
	Method      string                                    `json:"method" binding:"required"`
	Notes       string                                    `json:"notes"`
	Allocations []services.SupplierPaymentAllocationInput `json:"allocations"`
}

// RecordSupplierPayment godoc
// @Summary      Record a payment to a supplier against payables
// @Tags         finance
// @Accept       json
// @Produce      json
// @Param        request body RecordSupplierPaymentRequest true "payment"
// @Success      201 {object} domain.SupplierPayment
// @Security     BearerAuth
// @Router       /finance/supplier-payments [post]
func (h *FinanceHandler) RecordSupplierPayment(c *gin.Context) {
	var req RecordSupplierPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetUint("user_id")
	payment, err := h.AP.RecordPayment(services.RecordSupplierPaymentRequest{
		SupplierID:  req.SupplierID,
		Amount:      req.Amount,
		Method:      req.Method,
		Notes:       req.Notes,
		Allocations: req.Allocations,
	}, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, payment)
}
