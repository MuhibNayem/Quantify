package consumers

import (
	"context"
	"testing"

	"inventory/backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAlertConsumerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Role{},
		&domain.User{},
		&domain.AlertRoleSubscription{},
	))
	return db
}

func TestLoadSubscribedUsers_FindsByRoleName(t *testing.T) {
	db := setupAlertConsumerTestDB(t)

	managerRole := domain.Role{Name: "Manager"}
	require.NoError(t, db.Create(&managerRole).Error)

	require.NoError(t, db.Create(&domain.User{
		Username:   "alice",
		Password:   "x",
		Email:      "alice@example.com",
		PhoneNumber: "1111111111",
		RoleID:     managerRole.ID,
	}).Error)
	require.NoError(t, db.Create(&domain.User{
		Username:   "bob",
		Password:   "x",
		Email:      "bob@example.com",
		PhoneNumber: "2222222222",
		RoleID:     managerRole.ID,
	}).Error)

	clerkRole := domain.Role{Name: "Clerk"}
	require.NoError(t, db.Create(&clerkRole).Error)
	require.NoError(t, db.Create(&domain.User{
		Username:   "carol",
		Password:   "x",
		Email:      "carol@example.com",
		PhoneNumber: "3333333333",
		RoleID:     clerkRole.ID,
	}).Error)

	require.NoError(t, db.Create(&domain.AlertRoleSubscription{
		AlertType: "LOW_STOCK",
		Role:      "Manager",
	}).Error)

	users, err := loadSubscribedUsers(context.Background(), db, "LOW_STOCK")
	require.NoError(t, err)

	usernames := make([]string, 0, len(users))
	for _, u := range users {
		usernames = append(usernames, u.Username)
	}
	assert.ElementsMatch(t, []string{"alice", "bob"}, usernames,
		"subscribed Managers must be returned (Clerk 'carol' must be excluded)")
}

func TestLoadSubscribedUsers_NoSubscriptionsReturnsNil(t *testing.T) {
	db := setupAlertConsumerTestDB(t)

	users, err := loadSubscribedUsers(context.Background(), db, "UNSUBSCRIBED_TYPE")
	require.NoError(t, err)
	assert.Empty(t, users)
}
