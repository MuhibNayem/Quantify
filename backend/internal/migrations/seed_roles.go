package migrations

import (
	"inventory/backend/internal/domain"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SeedRoles ensures default roles exist in the database.
func SeedRoles(db *gorm.DB) error {
	roles := []string{"Admin", "Manager", "Staff", "Customer"}

	for _, roleName := range roles {
		var role domain.Role
		err := db.Where("name = ?", roleName).First(&role).Error
		if err == nil {
			continue // Role already exists
		}
		if err != gorm.ErrRecordNotFound {
			logrus.Errorf("Failed to check for role %s: %v", roleName, err)
			return err
		}

		// Create role
		role = domain.Role{Name: roleName}
		if err := db.Create(&role).Error; err != nil {
			logrus.Errorf("Failed to create role %s: %v", roleName, err)
			return err
		}
		logrus.Infof("Role %s created successfully.", roleName)
	}

	return nil
}
