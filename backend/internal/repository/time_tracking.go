package repository

import (
	"inventory/backend/internal/domain"
	"gorm.io/gorm"
)

type TimeTrackingRepository interface {
	CreateTimeClock(timeClock *domain.TimeClock) error
	GetLastTimeClock(userID uint) (*domain.TimeClock, error)
	UpdateTimeClock(timeClock *domain.TimeClock) error
	GetUserByUsername(username string) (*domain.User, error)
}

type timeTrackingRepository struct {
	db *gorm.DB
}

func NewTimeTrackingRepository(db *gorm.DB) TimeTrackingRepository {
	return &timeTrackingRepository{db: db}
}

func (r *timeTrackingRepository) CreateTimeClock(timeClock *domain.TimeClock) error {
	return r.db.Create(timeClock).Error
}

func (r *timeTrackingRepository) GetLastTimeClock(userID uint) (*domain.TimeClock, error) {
	var timeClock domain.TimeClock
	if err := r.db.Where("user_id = ?", userID).Order("clock_in desc").First(&timeClock).Error; err != nil {
		return nil, err
	}
	return &timeClock, nil
}

func (r *timeTrackingRepository) UpdateTimeClock(timeClock *domain.TimeClock) error {
	return r.db.Save(timeClock).Error
}

func (r *timeTrackingRepository) GetUserByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
