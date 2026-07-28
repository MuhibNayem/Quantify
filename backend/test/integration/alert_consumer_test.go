//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"inventory/backend/internal/domain"
)

func TestAlertConsumerSubscribedUsers(t *testing.T) {
	db := SetupPostgres(t)
	Migrate(t, db)

	role := domain.Role{Name: "Manager"}
	require.NoError(t, db.Create(&role).Error)

	user := domain.User{
		Username: "manager1",
		Password: "pass",
		Email:    "manager1@test.com",
		RoleID:   role.ID,
		IsActive: true,
	}
	require.NoError(t, db.Create(&user).Error)

	sub := domain.AlertRoleSubscription{AlertType: "LOW_STOCK", Role: "Manager"}
	require.NoError(t, db.Create(&sub).Error)

	notifSettings := domain.UserNotificationSettings{
		UserID:                    user.ID,
		EmailNotificationsEnabled: true,
		EmailAddress:              "manager1@test.com",
	}
	require.NoError(t, db.Create(&notifSettings).Error)

	var subscriptions []domain.AlertRoleSubscription
	require.NoError(t, db.Where("alert_type = ?", "LOW_STOCK").Find(&subscriptions).Error)
	require.Len(t, subscriptions, 1)

	roleNames := make([]string, 0, len(subscriptions))
	for _, s := range subscriptions {
		roleNames = append(roleNames, s.Role)
	}

	var users []domain.User
	err := db.
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name IN ?", roleNames).
		Where("NOT EXISTS (SELECT 1 FROM user_notification_subscriptions s WHERE s.user_id = users.id AND s.type = ? AND s.enabled = false)", "LOW_STOCK").
		Find(&users).Error
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "manager1", users[0].Username)

	optOut := domain.UserNotificationSubscription{
		UserID:  user.ID,
		Type:    "LOW_STOCK",
		Enabled: false,
	}
	require.NoError(t, db.Create(&optOut).Error)

	var usersAfterOptOut []domain.User
	err = db.
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name IN ?", roleNames).
		Where("NOT EXISTS (SELECT 1 FROM user_notification_subscriptions s WHERE s.user_id = users.id AND s.type = ? AND s.enabled = false)", "LOW_STOCK").
		Find(&usersAfterOptOut).Error
	require.NoError(t, err)
	assert.Len(t, usersAfterOptOut, 0)
}
