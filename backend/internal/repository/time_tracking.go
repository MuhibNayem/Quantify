package repository

import (
	"inventory/backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

type TimeTrackingRepository interface {
	CreateTimeClock(timeClock *domain.TimeClock) error
	GetLastTimeClock(userID uint) (*domain.TimeClock, error)
	UpdateTimeClock(timeClock *domain.TimeClock) error
	GetUserByUsername(username string) (*domain.User, error)
	GetRecentTimeClocks(userID uint, limit int) ([]domain.TimeClock, error)
	GetAllRecentTimeClocks(limit int) ([]domain.TimeClock, error)
	GetActiveTimeClocks() ([]domain.TimeClock, error)
	GetTimeClocksByDateRange(userID uint, start, end time.Time) ([]domain.TimeClock, error)
	GetAllTimeClocksByDateRange(start, end time.Time) ([]domain.TimeClock, error)
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

func (r *timeTrackingRepository) GetRecentTimeClocks(userID uint, limit int) ([]domain.TimeClock, error) {
	var clocks []domain.TimeClock
	if err := r.db.Where("user_id = ?", userID).Order("clock_in desc").Limit(limit).Find(&clocks).Error; err != nil {
		return nil, err
	}
	return clocks, nil
}

func (r *timeTrackingRepository) GetAllRecentTimeClocks(limit int) ([]domain.TimeClock, error) {
	var clocks []domain.TimeClock
	if err := r.db.Preload("User").Order("clock_in desc").Limit(limit).Find(&clocks).Error; err != nil {
		return nil, err
	}
	return clocks, nil
}

func (r *timeTrackingRepository) GetActiveTimeClocks() ([]domain.TimeClock, error) {
	var clocks []domain.TimeClock
	if err := r.db.Preload("User").Where("status IN ?", []string{"CLOCKED_IN", "ON_BREAK"}).Find(&clocks).Error; err != nil {
		return nil, err
	}
	return clocks, nil
}

func (r *timeTrackingRepository) GetTimeClocksByDateRange(userID uint, start, end time.Time) ([]domain.TimeClock, error) {
	var clocks []domain.TimeClock
	if err := r.db.Where("user_id = ? AND clock_in >= ? AND clock_in <= ?", userID, start, end).Order("clock_in asc").Find(&clocks).Error; err != nil {
		return nil, err
	}
	return clocks, nil
}

func (r *timeTrackingRepository) GetAllTimeClocksByDateRange(start, end time.Time) ([]domain.TimeClock, error) {
	var clocks []domain.TimeClock
	if err := r.db.Preload("User").Where("clock_in >= ? AND clock_in <= ?", start, end).Order("clock_in asc").Find(&clocks).Error; err != nil {
		return nil, err
	}
	return clocks, nil
}
