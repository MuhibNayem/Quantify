package handlers

import (
	"inventory/backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appErrors "inventory/backend/internal/errors"
)

type TimeTrackingHandler struct {
	timeTrackingService services.TimeTrackingService
}

func NewTimeTrackingHandler(timeTrackingService services.TimeTrackingService) *TimeTrackingHandler {
	return &TimeTrackingHandler{timeTrackingService: timeTrackingService}
}

func (h *TimeTrackingHandler) ClockIn(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	timeClock, err := h.timeTrackingService.ClockIn(uint(userID), req.Notes)
	if err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusBadRequest, err))
		return
	}
	c.JSON(http.StatusCreated, timeClock)
}

func (h *TimeTrackingHandler) ClockOut(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	timeClock, err := h.timeTrackingService.ClockOut(uint(userID), req.Notes)
	if err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusBadRequest, err))
		return
	}
	c.JSON(http.StatusOK, timeClock)
}

func (h *TimeTrackingHandler) GetLastTimeClock(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	timeClock, err := h.timeTrackingService.GetLastTimeClock(uint(userID))
	if err != nil {
		c.Error(appErrors.NewAppError("Time clock entry not found", http.StatusNotFound, err))
		return
	}
	c.JSON(http.StatusOK, timeClock)
}

func (h *TimeTrackingHandler) GetLastTimeClockByUsername(c *gin.Context) {
	username := c.Param("username")
	timeClock, err := h.timeTrackingService.GetLastTimeClockByUsername(username)
	if err != nil {
		c.Error(appErrors.NewAppError("Time clock entry not found", http.StatusNotFound, err))
		return
	}
	c.JSON(http.StatusOK, timeClock)
}
