package handlers

import (
	"fmt"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	repo repository.NotificationRepository
}

func NewNotificationHandler(repo repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

// ListNotifications godoc
// @Summary Get a list of notifications for a user
// @Description Retrieves a list of notifications for a specific user, with optional filtering by read status.
// @Tags notifications
// @Produce json
// @Param userId path int true "User ID"
// @Param isRead query bool false "Filter by read status"
// @Param limit query int false "Limit number of notifications"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} domain.Notification
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{userId}/notifications [get]
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	// Authorization check
	authUserID := c.GetUint("user_id")
	authUserRole := c.GetString("role")
	fmt.Println("authUserID:", authUserID)
	fmt.Println("authUserRole:", authUserRole)
	fmt.Println("userID:", userID)

	if authUserRole != "Admin" && authUserID != uint(userID) {
		c.Error(appErrors.NewAppError("Forbidden", http.StatusForbidden, nil))
		return
	}

	var isRead *bool
	if isReadQuery, ok := c.GetQuery("isRead"); ok {
		val, err := strconv.ParseBool(isReadQuery)
		if err != nil {
			c.Error(appErrors.NewAppError("Invalid isRead query parameter", http.StatusBadRequest, err))
			return
		}
		isRead = &val
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	notifications, err := h.repo.GetNotificationsByUserID(uint(userID), isRead, limit, offset)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch notifications", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// GetUnreadNotificationCount godoc
// @Summary Get the count of unread notifications for a user
// @Description Retrieves the count of unread notifications for the authenticated user.
// @Tags notifications
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{userId}/notifications/unread/count [get]
func (h *NotificationHandler) GetUnreadNotificationCount(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	// Authorization check
	authUserID := c.GetUint("user_id")
	if authUserID != uint(userID) {
		c.Error(appErrors.NewAppError("Forbidden", http.StatusForbidden, nil))
		return
	}

	count, err := h.repo.GetUnreadNotificationCountByUserID(uint(userID))
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch unread notification count", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkNotificationAsRead godoc
// @Summary Mark a notification as read
// @Description Marks a specific notification as read for the authenticated user.
// @Tags notifications
// @Produce json
// @Param userId path int true "User ID"
// @Param notificationId path int true "Notification ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{userId}/notifications/{notificationId}/read [patch]
func (h *NotificationHandler) MarkNotificationAsRead(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}
	notificationID, err := strconv.ParseUint(c.Param("notificationId"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid notification ID", http.StatusBadRequest, err))
		return
	}

	// Authorization check
	authUserID := c.GetUint("user_id")
	if authUserID != uint(userID) {
		c.Error(appErrors.NewAppError("Forbidden", http.StatusForbidden, nil))
		return
	}

	if err := h.repo.MarkNotificationAsRead(uint(notificationID), uint(userID)); err != nil {
		c.Error(appErrors.NewAppError("Failed to mark notification as read", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllNotificationsAsRead godoc
// @Summary Mark all notifications as read
// @Description Marks all unread notifications as read for the authenticated user.
// @Tags notifications
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{userId}/notifications/read-all [patch]
func (h *NotificationHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	// Authorization check
	authUserID := c.GetUint("user_id")
	if authUserID != uint(userID) {
		c.Error(appErrors.NewAppError("Forbidden", http.StatusForbidden, nil))
		return
	}

	if err := h.repo.MarkAllNotificationsAsRead(uint(userID)); err != nil {
		c.Error(appErrors.NewAppError("Failed to mark all notifications as read", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}
