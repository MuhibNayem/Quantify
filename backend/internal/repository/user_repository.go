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
		if err := tx.First(user, user.ID).Error; err != nil {
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
