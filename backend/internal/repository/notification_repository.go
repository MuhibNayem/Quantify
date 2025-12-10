package repository

import (
	"inventory/backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

// NotificationRepository defines methods for interacting with notifications.
type NotificationRepository interface {
	CreateNotification(notification *domain.Notification) error
	GetNotificationsByUserID(userID uint, isRead *bool, limit, offset int) ([]domain.Notification, error)
	GetUnreadNotificationCountByUserID(userID uint) (int64, error)
	MarkNotificationAsRead(notificationID uint, userID uint) error
	MarkAllNotificationsAsRead(userID uint) error
	CreateNotificationsForPermission(permission string, notification domain.Notification) error
}

// notificationRepository implements NotificationRepository using GORM.
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new NotificationRepository.
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// CreateNotification creates a new notification record.
func (r *notificationRepository) CreateNotification(notification *domain.Notification) error {
	return r.db.Create(notification).Error
}

// GetNotificationsByUserID retrieves notifications for a specific user.
func (r *notificationRepository) GetNotificationsByUserID(userID uint, isRead *bool, limit, offset int) ([]domain.Notification, error) {
	var notifications []domain.Notification
	query := r.db.Where("user_id = ?", userID).Order("created_at DESC")

	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset >= 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&notifications).Error
	return notifications, err
}

// GetUnreadNotificationCountByUserID retrieves the count of unread notifications for a user.
func (r *notificationRepository) GetUnreadNotificationCountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&domain.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

// MarkNotificationAsRead marks a specific notification as read.
func (r *notificationRepository) MarkNotificationAsRead(notificationID uint, userID uint) error {
	return r.db.Model(&domain.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{"is_read": true, "read_at": time.Now()}).Error
}

// MarkAllNotificationsAsRead marks all unread notifications for a user as read.
func (r *notificationRepository) MarkAllNotificationsAsRead(userID uint) error {
	return r.db.Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{"is_read": true, "read_at": time.Now()}).Error
}

// CreateNotificationsForPermission creates a notification for all users with a specific permission.
func (r *notificationRepository) CreateNotificationsForPermission(permission string, notification domain.Notification) error {
	var userIDs []uint
	// Find all users who have a role containing the given permission
	// This assumes the schema: Users -> Role -> RolePermissions -> Permissions
	// We use a subquery or join to find matching user IDs.
	// Since GORM Many2Many is "role_permissions", we query:
	err := r.db.Table("users").
		Select("users.id").
		Joins("JOIN roles ON users.role_id = roles.id").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("permissions.name = ?", permission).
		Find(&userIDs).Error

	if err != nil {
		return err
	}

	if len(userIDs) == 0 {
		return nil
	}

	var notifications []domain.Notification
	for _, uid := range userIDs {
		n := notification // Copy struct
		n.ID = 0          // Reset ID for new creation
		n.UserID = uid
		notifications = append(notifications, n)
	}

	return r.db.CreateInBatches(notifications, 100).Error
}
