package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"inventory/backend/internal/config"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"net/http"

	"strconv"

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
	GetChurnRisk(customerID uint) (*domain.ChurnRisk, error)
}

type crmService struct {
	repo     repository.CRMRepository
	db       *gorm.DB
	settings SettingsService
	cfg      *config.Config
}

func NewCRMService(repo repository.CRMRepository, db *gorm.DB, settings SettingsService, cfg *config.Config) CRMService {
	return &crmService{
		repo:     repo,
		db:       db,
		settings: settings,
		cfg:      cfg,
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
		Role:        role, // Explicitly set association
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
func (s *crmService) calculateTier(points int) string {
	silverThreshold := 500
	goldThreshold := 2500
	platinumThreshold := 10000

	if val, err := s.settings.GetSetting("loyalty_tier_silver"); err == nil {
		if v, err := strconv.Atoi(val); err == nil {
			silverThreshold = v
		}
	}
	if val, err := s.settings.GetSetting("loyalty_tier_gold"); err == nil {
		if v, err := strconv.Atoi(val); err == nil {
			goldThreshold = v
		}
	}
	if val, err := s.settings.GetSetting("loyalty_tier_platinum"); err == nil {
		if v, err := strconv.Atoi(val); err == nil {
			platinumThreshold = v
		}
	}

	if points >= platinumThreshold {
		return "Platinum"
	} else if points >= goldThreshold {
		return "Gold"
	} else if points >= silverThreshold {
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
	newTier := s.calculateTier(account.Points)
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
	newTier := s.calculateTier(account.Points)
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

func (s *crmService) GetChurnRisk(customerID uint) (*domain.ChurnRisk, error) {
	// 1. Fetch Customer Data
	customer, err := s.repo.GetUserByID(customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customer: %w", err)
	}

	// 2. Fetch Purchase History (Last 5 orders)
	type OrderHistory struct {
		Date     string  `json:"date"`
		Total    float64 `json:"total"`
		Items    int     `json:"items"`
		Discount float64 `json:"discount"`
	}
	var orders []domain.Order
	if err := s.db.Where("user_id = ?", customerID).Order("created_at desc").Limit(5).Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	var purchaseHistory []OrderHistory
	for _, order := range orders {
		purchaseHistory = append(purchaseHistory, OrderHistory{
			Date:     order.CreatedAt.Format("2006-01-02"),
			Total:    order.TotalAmount,
			Items:    len(order.OrderItems), // Fixed field name
			Discount: order.DiscountAmount,  // Fixed field name
		})
	}

	// 3. Prepare Payload
	payload := map[string]interface{}{
		"customer_data": map[string]interface{}{
			"id":         customer.ID,
			"name":       customer.FirstName + " " + customer.LastName,
			"email":      customer.Email,
			"created_at": customer.CreatedAt.Format("2006-01-02"),
		},
		"purchase_history": purchaseHistory,
	}

	// 4. Call AI Service
	if s.cfg.AIServiceURL == "" {
		// Fallback if AI service is not configured
		return &domain.ChurnRisk{
			ChurnRiskScore:    0.0,
			RiskLevel:         "Unknown",
			PrimaryFactors:    []string{"AI Service not configured"},
			RetentionStrategy: "N/A",
		}, nil
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(s.cfg.AIServiceURL+"/predict-churn", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ai service returned status: %d", resp.StatusCode)
	}

	var result domain.ChurnRisk
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode AI response: %w", err)
	}

	return &result, nil
}
