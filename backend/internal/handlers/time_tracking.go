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

// ClockIn godoc
// @Summary Clock in
// @Description Records a clock-in event for a user
// @Tags time-tracking
// @Accept json
// @Produce json
// @Param userId path int true "User ID"
// @Param notes body object{notes=string} false "Optional notes"
// @Security BearerAuth
// @Success 201 {object} object
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Router /time-tracking/clock-in/{userId} [post]
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

// ClockOut godoc
// @Summary Clock out
// @Description Records a clock-out event for a user
// @Tags time-tracking
// @Accept json
// @Produce json
// @Param userId path int true "User ID"
// @Param notes body object{notes=string} false "Optional notes"
// @Security BearerAuth
// @Success 200 {object} object
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Router /time-tracking/clock-out/{userId} [post]
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

// GetLastTimeClock godoc
// @Summary Get last time clock entry
// @Description Retrieves the most recent time clock entry for a user
// @Tags time-tracking
// @Produce json
// @Param userId path int true "User ID"
// @Security BearerAuth
// @Success 200 {object} object
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Router /time-tracking/last-entry/{userId} [get]
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

// GetLastTimeClockByUsername godoc
// @Summary Get last time clock by username
// @Description Retrieves the most recent time clock entry by username
// @Tags time-tracking
// @Produce json
// @Param username path string true "Username"
// @Security BearerAuth
// @Success 200 {object} object
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Router /time-tracking/last-entry/username/{username} [get]
func (h *TimeTrackingHandler) GetLastTimeClockByUsername(c *gin.Context) {
	username := c.Param("username")
	timeClock, err := h.timeTrackingService.GetLastTimeClockByUsername(username)
	if err != nil {
		c.Error(appErrors.NewAppError("Time clock entry not found", http.StatusNotFound, err))
		return
	}
	c.JSON(http.StatusOK, timeClock)
}

// StartBreak godoc
// @Summary Start a break
// @Description Records a break start for a user
// @Tags time-tracking
// @Produce json
// @Param userId path int true "User ID"
// @Security BearerAuth
// @Success 200 {object} object
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Router /time-tracking/break-start/{userId} [post]
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

// EndBreak godoc
// @Summary End a break
// @Description Records a break end for a user
// @Tags time-tracking
// @Produce json
// @Param userId path int true "User ID"
// @Security BearerAuth
// @Success 200 {object} object
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Router /time-tracking/break-end/{userId} [post]
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

// GetHistory godoc
// @Summary Get time tracking history
// @Description Retrieves recent time tracking history for a user
// @Tags time-tracking
// @Produce json
// @Param userId path int true "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /time-tracking/history/{userId} [get]
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

// GetTeamStatus godoc
// @Summary Get team status
// @Description Retrieves current clock-in/out status for all team members
// @Tags time-tracking
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /time-tracking/team-status [get]
func (h *TimeTrackingHandler) GetTeamStatus(c *gin.Context) {
	status, err := h.timeTrackingService.GetTeamStatus()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch team status", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, status)
}

// GetRecentActivities godoc
// @Summary Get recent activities
// @Description Retrieves recent time tracking activities across all users
// @Tags time-tracking
// @Produce json
// @Security BearerAuth
// @Success 200 {array} object
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /time-tracking/activities [get]
func (h *TimeTrackingHandler) GetRecentActivities(c *gin.Context) {
	activities, err := h.timeTrackingService.GetRecentActivities(10) // Limit 10
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch activities", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, activities)
}

// GetWeeklySummary godoc
// @Summary Get weekly summary
// @Description Retrieves weekly time tracking summary for a user
// @Tags time-tracking
// @Produce json
// @Param userId path int true "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /time-tracking/weekly-summary/{userId} [get]
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

// GetTeamOverview godoc
// @Summary Get team overview
// @Description Retrieves comprehensive team time tracking overview
// @Tags time-tracking
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /time-tracking/team-overview [get]
func (h *TimeTrackingHandler) GetTeamOverview(c *gin.Context) {
	overview, err := h.timeTrackingService.GetTeamOverview()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch team overview", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, overview)
}
