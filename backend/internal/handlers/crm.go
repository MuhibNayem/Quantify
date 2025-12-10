package handlers

import (
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CRMHandler struct {
	crmService services.CRMService
}

func NewCRMHandler(crmService services.CRMService) *CRMHandler {
	return &CRMHandler{crmService: crmService}
}

func (h *CRMHandler) ListCustomers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 200 {
		limit = 200
	}
	search := c.Query("q")

	users, total, err := h.crmService.ListCustomers(page, limit, search)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to list customers", http.StatusInternalServerError, err))
		return
	}

	// Remove passwords
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"users":        users,
		"totalItems":   total,
		"currentPage":  page,
		"totalPages":   (total + int64(limit) - 1) / int64(limit),
		"itemsPerPage": limit,
	})
}

func (h *CRMHandler) CreateCustomer(c *gin.Context) {
	var req requests.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	user, err := h.crmService.CreateCustomer(&req)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to create customer", http.StatusInternalServerError, err))
		return
	}

	user.Password = ""
	c.JSON(http.StatusCreated, user)
}

func (h *CRMHandler) GetCustomer(c *gin.Context) {
	identifier := c.Param("identifier")
	var user *domain.User
	var err error

	// Try parsing as numeric ID first (most common case)
	if userID, parseErr := strconv.ParseUint(identifier, 10, 32); parseErr == nil {
		user, err = h.crmService.GetCustomerByID(uint(userID))
		if err == nil {
			user.Password = ""
			c.JSON(http.StatusOK, user)
			return
		}
		// If not found by ID, continue to other methods
	}

	// Check if identifier contains '@' (email)
	if strings.Contains(identifier, "@") && strings.Contains(identifier, ".") {
		user, err = h.crmService.GetCustomerByEmail(identifier)
		if err == nil {
			user.Password = ""
			c.JSON(http.StatusOK, user)
			return
		}
	}

	// Try phone number (digits with possible +, -, or spaces)
	cleanedPhone := strings.ReplaceAll(strings.ReplaceAll(identifier, "-", ""), " ", "")
	if len(cleanedPhone) >= 10 && (cleanedPhone[0] == '+' || (cleanedPhone[0] >= '0' && cleanedPhone[0] <= '9')) {
		user, err = h.crmService.GetCustomerByPhone(identifier)
		if err == nil {
			user.Password = ""
			c.JSON(http.StatusOK, user)
			return
		}
	}

	// Finally, try username
	user, err = h.crmService.GetCustomerByUsername(identifier)
	if err != nil {
		c.Error(appErrors.NewAppError("Customer not found", http.StatusNotFound, err))
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}

func (h *CRMHandler) GetCustomerByUsername(c *gin.Context) {
	username := c.Param("username")
	user, err := h.crmService.GetCustomerByUsername(username)
	if err != nil {
		c.Error(appErrors.NewAppError("Customer not found", http.StatusNotFound, err))
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

func (h *CRMHandler) GetCustomerByEmail(c *gin.Context) {
	email := c.Param("email")
	user, err := h.crmService.GetCustomerByEmail(email)
	if err != nil {
		c.Error(appErrors.NewAppError("Customer not found", http.StatusNotFound, err))
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

func (h *CRMHandler) GetCustomerByPhone(c *gin.Context) {
	phone := c.Param("phone")
	user, err := h.crmService.GetCustomerByPhone(phone)
	if err != nil {
		c.Error(appErrors.NewAppError("Customer not found", http.StatusNotFound, err))
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

func (h *CRMHandler) UpdateCustomer(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("identifier"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	var req requests.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	user, err := h.crmService.UpdateCustomer(uint(userID), &req)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to update customer", http.StatusInternalServerError, err))
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}

func (h *CRMHandler) DeleteCustomer(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("identifier"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	if err := h.crmService.DeleteCustomer(uint(userID)); err != nil {
		c.Error(appErrors.NewAppError("Failed to delete customer", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

func (h *CRMHandler) GetLoyaltyAccount(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	account, err := h.crmService.GetLoyaltyAccount(uint(userID))
	if err != nil {
		c.Error(appErrors.NewAppError("Loyalty account not found", http.StatusNotFound, err))
		return
	}
	c.JSON(http.StatusOK, account)
}

func (h *CRMHandler) AddLoyaltyPoints(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	var req struct {
		Points int `json:"points"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	account, err := h.crmService.AddLoyaltyPoints(uint(userID), req.Points)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to add loyalty points", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, account)
}

func (h *CRMHandler) RedeemLoyaltyPoints(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	var req struct {
		Points int `json:"points"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	account, err := h.crmService.RedeemLoyaltyPoints(uint(userID), req.Points)
	if err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusBadRequest, err))
		return
	}
	c.JSON(http.StatusOK, account)
}

func (h *CRMHandler) GetChurnRisk(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("identifier"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	risk, err := h.crmService.GetChurnRisk(uint(userID))
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to analyze churn risk", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, risk)
}
