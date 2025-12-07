package services

import (
	"fmt"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	ListCustomers(page, limit int, search string) ([]domain.User, int64, error)
}

type crmService struct {
	repo repository.CRMRepository
	db   *gorm.DB
}

func NewCRMService(repo repository.CRMRepository, db *gorm.DB) CRMService {
	return &crmService{
		repo: repo,
		db:   db,
	}
}

func (s *crmService) CreateCustomer(req *requests.CreateCustomerRequest) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var role domain.Role
	if err := s.db.Where("name = ?", "Customer").First(&role).Error; err != nil {
		return nil, fmt.Errorf("customer role not found: %w", err)
	}

	user := &domain.User{
		Username:    req.Username,
		Password:    string(hashedPassword),
		RoleID:      role.ID,
		IsActive:    true,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	}

	// Wrap in transaction to ensure atomicity
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create user
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// Create loyalty account
		account := &domain.LoyaltyAccount{
			UserID: user.ID,
		}
		if err := tx.Create(account).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
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

// calculateTier determines the loyalty tier based on points (bidirectional)
func calculateTier(points int) string {
	if points >= 10000 {
		return "Platinum"
	} else if points >= 5000 {
		return "Gold"
	} else if points >= 1000 {
		return "Silver"
	}
	return "Bronze"
}

func (s *crmService) AddLoyaltyPoints(userID uint, points int) (*domain.LoyaltyAccount, error) {
	// Use atomic operation to prevent race conditions
	if err := s.repo.AddLoyaltyPointsAtomic(userID, points); err != nil {
		return nil, err
	}

	// Fetch updated account
	account, err := s.repo.GetLoyaltyAccountByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Update tier based on new balance
	newTier := calculateTier(account.Points)
	if account.Tier != newTier {
		account.Tier = newTier
		if err := s.repo.UpdateLoyaltyAccount(account); err != nil {
			return nil, err
		}
	}

	return account, nil
}

func (s *crmService) RedeemLoyaltyPoints(userID uint, points int) (*domain.LoyaltyAccount, error) {
	// First check if user has enough points
	account, err := s.repo.GetLoyaltyAccountByUserID(userID)
	if err != nil {
		return nil, err
	}

	if account.Points < points {
		return nil, fmt.Errorf("not enough points to redeem")
	}

	// Use atomic operation to prevent race conditions
	if err := s.repo.AddLoyaltyPointsAtomic(userID, -points); err != nil {
		return nil, err
	}

	// Fetch updated account
	account, err = s.repo.GetLoyaltyAccountByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Update tier based on new balance (supports downgrades)
	newTier := calculateTier(account.Points)
	if account.Tier != newTier {
		account.Tier = newTier
		if err := s.repo.UpdateLoyaltyAccount(account); err != nil {
			return nil, err
		}
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

func (s *crmService) ListCustomers(page, limit int, search string) ([]domain.User, int64, error) {
	return s.repo.ListCustomers(page, limit, search)
}
