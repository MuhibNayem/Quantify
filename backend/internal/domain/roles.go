package domain

import "gorm.io/gorm"

// Role represents a user role in the system.
type Role struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	IsSystem    bool         `gorm:"default:false"` // System roles (Admin, Manager) cannot be deleted
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

// Permission represents a granular system permission.
type Permission struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"` // e.g., "inventory.read"
	Description string
	Group       string // e.g., "Inventory", "CRM"
}

// RolePermission is the join table for Roles and Permissions.
// Explicit definition allows for custom fields if needed later.
// Note: GORM handles many2many automatically, but define this if we need hooks or extra data.
// For now, we rely on GORM's automatic join table or explicit if we want to query it directly easier.
// Explicit model is safer for seeding and direct manipulation.
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}
