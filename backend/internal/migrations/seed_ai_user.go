package migrations

import (
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedAIUser ensures the AI Agent user exists with Admin privileges.
func SeedAIUser(db *gorm.DB) error {
	var adminRole domain.Role
	if err := db.Where("name = ?", "Admin").First(&adminRole).Error; err != nil {
		logrus.Errorf("Admin role not found: %v", err)
		return err
	}

	var user domain.User
	err := db.Where("email = ?", "ai-agent@quantify.com").First(&user).Error
	if err == nil {
		logrus.Info("AI Agent user already exists.")
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	// Create AI Agent user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("rZQ$4Rs!6{QHaR{5Sra{]z_%n"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user = domain.User{
		Username:    "ai-agent",
		Email:       "ai-agent@quantify.com",
		Password:    string(hashedPassword),
		RoleID:      adminRole.ID,
		IsActive:    true,
		FirstName:   "AI",
		LastName:    "Agent",
		PhoneNumber: "000-000-0000",
		Address:     "System",
	}

	if err := repository.NewUserRepository(db).CreateUser(&user); err != nil {
		logrus.Errorf("Failed to create AI Agent user: %v", err)
		return err
	}

	logrus.Info("AI Agent user created successfully.")
	return nil
}
