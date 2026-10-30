package migrations

import (
	"crypto/rand"
	"encoding/base64"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GenerateRandomPassword returns a cryptographically random URL-safe password (256-bit entropy).
func GenerateRandomPassword() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(b)
}

func buildAIUser(hashedPassword string, adminRole domain.Role) domain.User {
	return domain.User{
		Username:   "ai-agent",
		Email:      "ai-agent@quantify.com",
		Password:   hashedPassword,
		RoleID:     adminRole.ID,
		Role:       adminRole, // Explicitly set association
		RoleName:   adminRole.Name,
		LegacyRole: adminRole.Name, // Populate legacy column
		// The AI service authenticates as this user, so it must be active.
		IsActive:    true,
		FirstName:   "AI",
		LastName:    "Agent",
		PhoneNumber: "000-000-0000",
		Address:     "System",
	}
}

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

	// Create AI Agent user. Prefer the operator-provided password so the
	// backend and ai-service can share it via environment configuration;
	// otherwise generate one and print it exactly once.
	password := os.Getenv("AI_AGENT_PASSWORD")
	generated := password == ""
	if generated {
		password = GenerateRandomPassword()
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user = buildAIUser(string(hashedPassword), adminRole)

	if err := repository.NewUserRepository(db).CreateUser(&user); err != nil {
		logrus.Errorf("Failed to create AI Agent user: %v", err)
		return err
	}

	if generated {
		logrus.Infof("AI Agent user created successfully. One-time password (store securely and set AI_AGENT_PASSWORD): %s", password)
	} else {
		logrus.Info("AI Agent user created successfully using AI_AGENT_PASSWORD.")
	}
	return nil
}
