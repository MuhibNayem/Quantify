package repository

import (
	"inventory/backend/internal/domain"

	"gorm.io/gorm"
)

type LocationRepository struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) *LocationRepository {
	return &LocationRepository{db: db}
}

func (r *LocationRepository) GetAll() ([]domain.Location, error) {
	var locations []domain.Location
	if err := r.db.Find(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}

func (r *LocationRepository) GetByNames(names []string, locations *[]domain.Location) error {
	return r.db.Where("name IN ?", names).Find(locations).Error
}
