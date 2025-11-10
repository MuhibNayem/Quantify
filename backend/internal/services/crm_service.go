package services

import (
	"fmt"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"golang.org/x/crypto/bcrypt"
)

type CRMService interface {
	CreateLoyaltyAccount(userID uint) (*domain.LoyaltyAccount, error)
	AddLoyaltyPoints(userID uint, points int) (*domain.LoyaltyAccount, error)
	RedeemLoyaltyPoints(userID uint, points int) (*domain.LoyaltyAccount, error)
	GetLoyaltyAccount(userID uint) (*domain.LoyaltyAccount, error)
	GetCustomerByUsername(username string) (*domain.User, error)
	GetCustomerByEmail(email string) (*domain.User, error)
	GetCustomerByID(userID uint) (*domain.User, error)
	GetCustomerByPhone(phone string) (*domain.User, error)
	CreateCustomer(req *requests.CreateCustomerRequest) (*domain.User, error)
	UpdateCustomer(userID uint, req *requests.UpdateCustomerRequest) (*domain.User, error)
	DeleteCustomer(userID uint) error
}

type crmService struct {
	repo repository.CRMRepository
}

func NewCRMService(repo repository.CRMRepository) CRMService {
	return &crmService{repo: repo}
}

func (s *crmService) CreateCustomer(req *requests.CreateCustomerRequest) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:    req.Username,
		Password:    string(hashedPassword),
		Role:        "Customer",
		IsActive:    true,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	if _, err := s.CreateLoyaltyAccount(user.ID); err != nil {
		// TODO: Handle user creation rollback
		return nil, err
	}

	return user, nil
}

func (s *crmService) UpdateCustomer(userID uint, req *requests.UpdateCustomerRequest) (*domain.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *crmService) DeleteCustomer(userID uint) error {
	return s.repo.DeleteUser(userID)
}

func (s *crmService) CreateLoyaltyAccount(userID uint) (*domain.LoyaltyAccount, error) {
	account := &domain.LoyaltyAccount{
		UserID: userID,
	}
	if err := s.repo.CreateLoyaltyAccount(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *crmService) AddLoyaltyPoints(userID uint, points int) (*domain.LoyaltyAccount, error) {
	account, err := s.repo.GetLoyaltyAccountByUserID(userID)
	if err != nil {
		return nil, err
	}

	account.Points += points
	// Add logic for tier upgrades
	if account.Points >= 10000 {
		account.Tier = "Platinum"
	} else if account.Points >= 5000 {
		account.Tier = "Gold"
	} else if account.Points >= 1000 {
		account.Tier = "Silver"
	}

	if err := s.repo.UpdateLoyaltyAccount(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *crmService) RedeemLoyaltyPoints(userID uint, points int) (*domain.LoyaltyAccount, error) {
	account, err := s.repo.GetLoyaltyAccountByUserID(userID)
	if err != nil {
		return nil, err
	}

	if account.Points < points {
		return nil, fmt.Errorf("not enough points to redeem")
	}

	account.Points -= points
	if err := s.repo.UpdateLoyaltyAccount(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *crmService) GetLoyaltyAccount(userID uint) (*domain.LoyaltyAccount, error) {
	return s.repo.GetLoyaltyAccountByUserID(userID)
}

func (s *crmService) GetCustomerByUsername(username string) (*domain.User, error) {
	return s.repo.GetUserByUsername(username)
}

func (s *crmService) GetCustomerByEmail(email string) (*domain.User, error) {
	return s.repo.GetUserByEmail(email)
}

func (s *crmService) GetCustomerByID(userID uint) (*domain.User, error) {
	return s.repo.GetUserByID(userID)
}

func (s *crmService) GetCustomerByPhone(phone string) (*domain.User, error) {
	return s.repo.GetUserByPhone(phone)
}
