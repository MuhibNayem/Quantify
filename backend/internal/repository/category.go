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

func (r *CategoryRepository) CreateCategory(category *domain.Category) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(category).Error; err != nil {
			return err
		}
		return UpdateSearchIndex(tx, category)
	})
}

func (r *CategoryRepository) UpdateCategory(category *domain.Category, updates map[string]interface{}) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(category).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.First(category, category.ID).Error; err != nil {
			return err
		}
		return UpdateSearchIndex(tx, category)
	})
}

func (r *CategoryRepository) DeleteCategory(category *domain.Category) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(category).Error; err != nil {
			return err
		}
		return DeleteFromSearchIndex(tx, category.GetEntityType(), category.GetID())
	})
}

func (r *CategoryRepository) CreateSubCategory(subCategory *domain.SubCategory) error {
	return r.db.Create(subCategory).Error
}

func (r *CategoryRepository) GetAll() ([]domain.Category, error) {
	var categories []domain.Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepository) GetAllSubCategories() ([]domain.SubCategory, error) {
	var subCategories []domain.SubCategory
	if err := r.db.Find(&subCategories).Error; err != nil {
		return nil, err
	}
	return subCategories, nil
}

func (r *CategoryRepository) GetCategoryByName(name string) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.Where("name = ?", name).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}
