package repository

import (
	"inventory/backend/internal/domain"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *domain.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return UpdateSearchIndex(tx, user)
	})
}

func (r *UserRepository) UpdateUser(user *domain.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		// Reload the user to get all fields for indexing
		if err := tx.Preload("Role").First(user, user.ID).Error; err != nil {
			return err
		}
		return UpdateSearchIndex(tx, user)
	})
}

func (r *UserRepository) DeleteUser(user *domain.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(user).Error; err != nil {
			return err
		}
		return DeleteFromSearchIndex(tx, user.GetEntityType(), user.GetID())
	})
}

func (r *UserRepository) GetAllUsers() ([]domain.User, error) {
	var users []domain.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) ListUsers(page, limit int, status, search string) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64
	offset := (page - 1) * limit

	db := r.db.Model(&domain.User{})

	if status == "approved" {
		db = db.Where("is_active = ?", true)
	} else if status == "pending" {
		db = db.Where("is_active = ?", false)
	}

	if search != "" {
		pattern := "%" + search + "%"
		db = db.Where("username ILIKE ? OR CAST(id AS TEXT) ILIKE ?", pattern, pattern)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("Role").Order("id ASC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) GetUserByID(id string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Preload("Role").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Preload("Role.Permissions").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserRepository) GetRoleByName(name string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *UserRepository) GetRoleByID(id uint) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
