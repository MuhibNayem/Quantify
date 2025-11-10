package services

import (
	"fmt"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"time"
)

type TimeTrackingService interface {
	ClockIn(userID uint, notes string) (*domain.TimeClock, error)
	ClockOut(userID uint, notes string) (*domain.TimeClock, error)
	GetLastTimeClock(userID uint) (*domain.TimeClock, error)
	GetLastTimeClockByUsername(username string) (*domain.TimeClock, error)
}

type timeTrackingService struct {
	repo repository.TimeTrackingRepository
}

func NewTimeTrackingService(repo repository.TimeTrackingRepository) TimeTrackingService {
	return &timeTrackingService{repo: repo}
}

func (s *timeTrackingService) ClockIn(userID uint, notes string) (*domain.TimeClock, error) {
	lastTimeClock, err := s.repo.GetLastTimeClock(userID)
	if err == nil && lastTimeClock.ClockOut == nil {
		return nil, fmt.Errorf("user is already clocked in")
	}

	timeClock := &domain.TimeClock{
		UserID:  userID,
		ClockIn: time.Now(),
		Notes:   notes,
	}

	if err := s.repo.CreateTimeClock(timeClock); err != nil {
		return nil, err
	}
	return timeClock, nil
}

func (s *timeTrackingService) ClockOut(userID uint, notes string) (*domain.TimeClock, error) {
	lastTimeClock, err := s.repo.GetLastTimeClock(userID)
	if err != nil {
		return nil, fmt.Errorf("user is not clocked in")
	}

	if lastTimeClock.ClockOut != nil {
		return nil, fmt.Errorf("user is already clocked out")
	}

	now := time.Now()
	lastTimeClock.ClockOut = &now
	lastTimeClock.Notes = notes

	if err := s.repo.UpdateTimeClock(lastTimeClock); err != nil {
		return nil, err
	}
	return lastTimeClock, nil
}

func (s *timeTrackingService) GetLastTimeClock(userID uint) (*domain.TimeClock, error) {
	return s.repo.GetLastTimeClock(userID)
}

func (s *timeTrackingService) GetLastTimeClockByUsername(username string) (*domain.TimeClock, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	return s.repo.GetLastTimeClock(user.ID)
}
