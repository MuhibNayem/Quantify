package handlers

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"inventory/backend/internal/config"
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/services"
	"inventory/backend/internal/websocket"
)

func computeRefund(unitPrice float64, qty float64, taxRate float64) float64 {
	subtotal := unitPrice * qty
	tax := subtotal * taxRate
	return math.Round((subtotal+tax)*100) / 100
}

func computeLoyaltyReversal(pointsRedeemed int, itemQty, returnQty float64) int {
	if pointsRedeemed == 0 || itemQty == 0 {
		return 0
	}
	return int(float64(pointsRedeemed) * returnQty / itemQty)
}

type ReturnHandler struct {
	DB               *gorm.DB
	Cfg              *config.Config
	Settings         services.SettingsService
	Hub              *websocket.Hub
	NotificationRepo repository.NotificationRepository
	ReportingService *services.ReportingService
}

func NewReturnHandler(db *gorm.DB, cfg *config.Config, settings services.SettingsService, hub *websocket.Hub, notificationRepo repository.NotificationRepository, reportingService *services.ReportingService) *ReturnHandler {
	return &ReturnHandler{DB: db, Cfg: cfg, Settings: settings, Hub: hub, NotificationRepo: notificationRepo, ReportingService: reportingService}
}

// ListReturns godoc
// @Summary List returns
// @Description Retrieves a list of returns, optionally filtered by status
// @Tags returns
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (PENDING, APPROVED, REJECTED)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of returns"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /returns [get]
func (h *ReturnHandler) ListReturns(c *gin.Context) {
	status := c.Query("status")
	var returns []domain.Return

	query := h.DB.Preload("ReturnItems").Preload("ReturnItems.Product").Order("created_at desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&returns).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch returns", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"returns": returns})
}

// GetReturn godoc
// @Summary Get return details
// @Description Fetch details of a specific return request
// @Tags returns
// @Accept json
// @Produce json
// @Param id path int true "Return ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Return details"
// @Failure 404 {object} map[string]interface{} "Return not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /returns/{id} [get]
func (h *ReturnHandler) GetReturn(c *gin.Context) {
	id := c.Param("id")
	var returnRecord domain.Return

	if err := h.DB.Preload("ReturnItems.Product").Preload("Order").Preload("User").First(&returnRecord, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Return request not found", http.StatusNotFound, err))
		} else {
			c.Error(appErrors.NewAppError("Failed to fetch return details", http.StatusInternalServerError, err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"return": returnRecord})
}
