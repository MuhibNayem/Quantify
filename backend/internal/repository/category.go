package repository

import (
	"inventory/backend/internal/domain"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetCategoryByName(name string) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.Where("name = ?", name).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}
