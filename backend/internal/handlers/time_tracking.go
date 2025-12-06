package handlers

import (
	"inventory/backend/internal/services"
	"net/http"
	"strconv"

	appErrors "inventory/backend/internal/errors"

	"github.com/gin-gonic/gin"
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

func (h *TimeTrackingHandler) StartBreak(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	timeClock, err := h.timeTrackingService.StartBreak(uint(userID))
	if err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusBadRequest, err))
		return
	}
	c.JSON(http.StatusOK, timeClock)
}

func (h *TimeTrackingHandler) EndBreak(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	timeClock, err := h.timeTrackingService.EndBreak(uint(userID))
	if err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusBadRequest, err))
		return
	}
	c.JSON(http.StatusOK, timeClock)
}

func (h *TimeTrackingHandler) GetHistory(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	history, err := h.timeTrackingService.GetHistory(uint(userID), 5) // Limit 5 for now
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch history", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"history": history})
}

func (h *TimeTrackingHandler) GetTeamStatus(c *gin.Context) {
	status, err := h.timeTrackingService.GetTeamStatus()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch team status", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *TimeTrackingHandler) GetRecentActivities(c *gin.Context) {
	activities, err := h.timeTrackingService.GetRecentActivities(10) // Limit 10
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch activities", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, activities)
}

func (h *TimeTrackingHandler) GetWeeklySummary(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid User ID", http.StatusBadRequest, err))
		return
	}
	summary, err := h.timeTrackingService.GetWeeklySummary(uint(userId))
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch weekly summary", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *TimeTrackingHandler) GetTeamOverview(c *gin.Context) {
	overview, err := h.timeTrackingService.GetTeamOverview()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch team overview", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, overview)
}
