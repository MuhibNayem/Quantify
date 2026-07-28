package handlers

import (
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NotificationHandler struct {
	repo repository.NotificationRepository
	db   *gorm.DB
}

func NewNotificationHandler(repo repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

// WithDB injects a *gorm.DB into the handler for endpoints that need direct
// access (e.g. user-facing subscription management that has no repository
// layer yet). Returns the receiver for chaining.
func (h *NotificationHandler) WithDB(db *gorm.DB) *NotificationHandler {
	h.db = db
	return h
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
	authUserID := c.GetUint("user_id")
	if authUserID == 0 {
		c.Error(appErrors.NewAppError("Unauthenticated", http.StatusUnauthorized, nil))
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	// Authorization check
	authUserRole := c.GetString("role")
	logrus.WithFields(logrus.Fields{
		"auth_user_id": authUserID,
		"auth_role":    authUserRole,
		"user_id":      userID,
	}).Debug("ListNotifications authorization check")

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
	authUserID := c.GetUint("user_id")
	if authUserID == 0 {
		c.Error(appErrors.NewAppError("Unauthenticated", http.StatusUnauthorized, nil))
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	// Authorization check
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
	authUserID := c.GetUint("user_id")
	if authUserID == 0 {
		c.Error(appErrors.NewAppError("Unauthenticated", http.StatusUnauthorized, nil))
		return
	}

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
	authUserID := c.GetUint("user_id")
	if authUserID == 0 {
		c.Error(appErrors.NewAppError("Unauthenticated", http.StatusUnauthorized, nil))
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	// Authorization check
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

// BroadcastNotificationRequest defines the payload for broadcasting a notification.
type BroadcastNotificationRequest struct {
	Permission string `json:"permission" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Message    string `json:"message" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Payload    string `json:"payload"`
}

// BroadcastNotification godoc
// @Summary Broadcast a notification to users with a specific permission
// @Description Sends a notification to all users who have the specified permission.
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body BroadcastNotificationRequest true "Broadcast request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications/broadcast [post]
func (h *NotificationHandler) BroadcastNotification(c *gin.Context) {
	var req BroadcastNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	notification := domain.Notification{
		Type:    req.Type,
		Title:   req.Title,
		Message: req.Message,
		Payload: req.Payload,
	}

	if err := h.repo.CreateNotificationsForPermission(req.Permission, notification); err != nil {
		c.Error(appErrors.NewAppError("Failed to broadcast notification", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification broadcasted successfully"})
}

// UpdateSubscription godoc
// @Summary Set a user's opt-in/out for a specific alert type
// @Description Upserts the user's subscription for the given alert type. When
// @Description `enabled` is false the user will no longer receive alerts of
// @Description that type from the alert consumer.
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param type path string true "Alert type (e.g. LOW_STOCK, OUT_OF_STOCK)"
// @Param request body object{enabled=bool} true "Subscription payload"
// @Success 200 {object} domain.UserNotificationSubscription
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id}/notifications/subscriptions/{type} [put]
func (h *NotificationHandler) UpdateSubscription(c *gin.Context) {
	authUserID := c.GetUint("user_id")
	if authUserID == 0 {
		c.Error(appErrors.NewAppError("Unauthenticated", http.StatusUnauthorized, nil))
		return
	}

	pathUserID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	// Authorization: users may only manage their own subscriptions. Admins
	// can override (consistent with other notification endpoints).
	if authUserID != uint(pathUserID) && c.GetString("role") != "Admin" {
		c.Error(appErrors.NewAppError("Forbidden", http.StatusForbidden, nil))
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	if h.db == nil {
		c.Error(appErrors.NewAppError("Database not configured", http.StatusInternalServerError, nil))
		return
	}

	alertType := c.Param("type")
	sub := domain.UserNotificationSubscription{
		UserID:  uint(pathUserID),
		Type:    alertType,
		Enabled: req.Enabled,
	}
	if err := h.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "type"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled", "updated_at"}),
	}).Create(&sub).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to save subscription", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, sub)
}

// ListSubscriptions godoc
// @Summary List a user's notification subscriptions
// @Description Returns all subscription rows for the user (one per alert type).
// @Tags notifications
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {array} domain.UserNotificationSubscription
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id}/notifications/subscriptions [get]
func (h *NotificationHandler) ListSubscriptions(c *gin.Context) {
	authUserID := c.GetUint("user_id")
	if authUserID == 0 {
		c.Error(appErrors.NewAppError("Unauthenticated", http.StatusUnauthorized, nil))
		return
	}

	pathUserID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	if authUserID != uint(pathUserID) && c.GetString("role") != "Admin" {
		c.Error(appErrors.NewAppError("Forbidden", http.StatusForbidden, nil))
		return
	}

	if h.db == nil {
		c.Error(appErrors.NewAppError("Database not configured", http.StatusInternalServerError, nil))
		return
	}

	var subs []domain.UserNotificationSubscription
	if err := h.db.Where("user_id = ?", uint(pathUserID)).
		Order("type ASC").
		Find(&subs).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch subscriptions", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, subs)
}
