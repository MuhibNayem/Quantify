package repository

import (
	"inventory/backend/internal/domain"

	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) CreateCategory(category *domain.Category) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(category).Error; err != nil {
			return err
		}
		return UpdateSearchIndex(tx, category)
	})
}

func (r *categoryRepository) UpdateCategory(category *domain.Category, updates map[string]interface{}) error {
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

func (r *categoryRepository) DeleteCategory(category *domain.Category) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(category).Error; err != nil {
			return err
		}
		return DeleteFromSearchIndex(tx, category.GetEntityType(), category.GetID())
	})
}

func (r *categoryRepository) CreateSubCategory(subCategory *domain.SubCategory) error {
	return r.db.Create(subCategory).Error
}

func (r *categoryRepository) GetAll() ([]domain.Category, error) {
	var categories []domain.Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepository) GetAllSubCategories() ([]domain.SubCategory, error) {
	var subCategories []domain.SubCategory
	if err := r.db.Find(&subCategories).Error; err != nil {
		return nil, err
	}
	return subCategories, nil
}

func (r *categoryRepository) GetByNames(names []string, categories *[]domain.Category) error {
	return r.db.Where("name IN ?", names).Find(categories).Error
}

func (r *categoryRepository) GetSubCategoriesByNames(names []string, subCategories *[]domain.SubCategory) error {
	return r.db.Where("name IN ?", names).Find(subCategories).Error
}

func (r *categoryRepository) GetCategoryByName(name string) (*domain.Category, error) {
	var category domain.Category
	if err := r.db.Where("name = ?", name).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}
