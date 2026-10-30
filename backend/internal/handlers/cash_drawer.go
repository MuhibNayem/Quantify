package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
)

// CashDrawerHandler manages cashier shifts: opening float, cash sales,
// refunds, safe drops, closing count, and variance.
type CashDrawerHandler struct {
	DB *gorm.DB
}

func NewCashDrawerHandler(db *gorm.DB) *CashDrawerHandler {
	return &CashDrawerHandler{DB: db}
}

type openDrawerRequest struct {
	StartingCash float64 `json:"startingCash" binding:"gte=0"`
	LocationID   uint    `json:"locationId"`
	Notes        string  `json:"notes"`
}

type closeDrawerRequest struct {
	EndingCash float64 `json:"endingCash" binding:"gte=0"`
	Notes      string  `json:"notes"`
}

type cashDropRequest struct {
	Amount float64 `json:"amount" binding:"gt=0"`
	Reason string  `json:"reason"`
}

func currentUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := value.(uint)
	return id, ok
}

func recalcSystemCash(s *domain.CashDrawerSession) {
	s.SystemCash = s.StartingCash + s.TotalSales - s.TotalRefunds - s.TotalDrops
}

// Open starts a shift. A cashier can only have one open drawer at a time.
func (h *CashDrawerHandler) Open(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}

	var req openDrawerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var session domain.CashDrawerSession
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var openCount int64
		if err := tx.Model(&domain.CashDrawerSession{}).
			Where("user_id = ? AND status = ?", userID, "OPEN").
			Count(&openCount).Error; err != nil {
			return err
		}
		if openCount > 0 {
			return fmt.Errorf("a cash drawer is already open for this user")
		}

		// location_id is NOT NULL with an FK. Default to the first configured
		// location when the POS doesn't send one.
		if req.LocationID == 0 {
			var location domain.Location
			if err := tx.First(&location).Error; err != nil {
				return fmt.Errorf("create a location before opening a cash drawer")
			}
			req.LocationID = location.ID
		}

		session = domain.CashDrawerSession{
			UserID:       userID,
			LocationID:   req.LocationID,
			StartTime:    time.Now(),
			StartingCash: req.StartingCash,
			SystemCash:   req.StartingCash,
			Status:       "OPEN",
			Notes:        req.Notes,
		}
		return tx.Create(&session).Error
	})
	if err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusConflict, err))
		return
	}

	h.DB.Preload("Location").First(&session, session.ID)
	c.JSON(http.StatusCreated, session)
}

// Current returns the caller's open drawer, if any.
func (h *CashDrawerHandler) Current(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}

	var session domain.CashDrawerSession
	err := h.DB.Preload("Location").
		Where("user_id = ? AND status = ?", userID, "OPEN").
		Order("start_time DESC").
		First(&session).Error
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, gin.H{"session": nil})
		return
	}
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to load cash drawer", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"session": session})
}

// Drop removes cash from the drawer (safe drop) and reduces expected cash.
func (h *CashDrawerHandler) Drop(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}

	var req cashDropRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var session domain.CashDrawerSession
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND status = ?", userID, "OPEN").
			Order("start_time DESC").First(&session).Error; err != nil {
			return fmt.Errorf("no open cash drawer")
		}

		drop := domain.CashDrop{
			SessionID: session.ID,
			Amount:    req.Amount,
			Reason:    req.Reason,
			DroppedBy: userID,
			DroppedAt: time.Now(),
		}
		if err := tx.Create(&drop).Error; err != nil {
			return err
		}

		session.TotalDrops += req.Amount
		recalcSystemCash(&session)
		return tx.Save(&session).Error
	})
	if err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusBadRequest, err))
		return
	}

	c.JSON(http.StatusOK, session)
}

// Close records the counted cash and freezes the drawer.
func (h *CashDrawerHandler) Close(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}

	var req closeDrawerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var session domain.CashDrawerSession
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND status = ?", userID, "OPEN").
			Order("start_time DESC").First(&session).Error; err != nil {
			return fmt.Errorf("no open cash drawer")
		}

		recalcSystemCash(&session)
		now := time.Now()
		session.EndTime = &now
		session.EndingCash = req.EndingCash
		session.Variance = req.EndingCash - session.SystemCash
		session.Status = "CLOSED"
		if req.Notes != "" {
			if session.Notes != "" {
				session.Notes += " | "
			}
			session.Notes += req.Notes
		}
		return tx.Save(&session).Error
	})
	if err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusBadRequest, err))
		return
	}

	c.JSON(http.StatusOK, session)
}

// List returns the caller's recent shifts.
func (h *CashDrawerHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}

	limit := 50
	var sessions []domain.CashDrawerSession
	if err := h.DB.Preload("Location").
		Where("user_id = ?", userID).
		Order("start_time DESC").
		Limit(limit).
		Find(&sessions).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to load cash drawer history", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}
