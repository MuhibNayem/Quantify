package repository

import (
	"inventory/backend/internal/domain"

	"gorm.io/gorm"
)

type CRMRepository interface {
	CreateLoyaltyAccount(account *domain.LoyaltyAccount) error
	GetLoyaltyAccountByUserID(userID uint) (*domain.LoyaltyAccount, error)
	UpdateLoyaltyAccount(account *domain.LoyaltyAccount) error
	AddLoyaltyPointsAtomic(userID uint, points int) error
	GetUserByUsername(username string) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	GetUserByID(userID uint) (*domain.User, error)
	GetUserByPhone(phone string) (*domain.User, error)
	CreateUser(user *domain.User) error
	UpdateUser(user *domain.User) error
	DeleteUser(userID uint) error
	ListCustomers(page, limit int, search string) ([]domain.User, int64, error)
}

type crmRepository struct {
	db *gorm.DB
}

func NewCRMRepository(db *gorm.DB) CRMRepository {
	return &crmRepository{db: db}
}

func (r *crmRepository) CreateLoyaltyAccount(account *domain.LoyaltyAccount) error {
	return r.db.Create(account).Error
}

func (r *crmRepository) GetLoyaltyAccountByUserID(userID uint) (*domain.LoyaltyAccount, error) {
	var account domain.LoyaltyAccount
	if err := r.db.Where("user_id = ?", userID).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *crmRepository) UpdateLoyaltyAccount(account *domain.LoyaltyAccount) error {
	return r.db.Save(account).Error
}

// AddLoyaltyPointsAtomic atomically updates points to prevent race conditions
func (r *crmRepository) AddLoyaltyPointsAtomic(userID uint, points int) error {
	return r.db.Model(&domain.LoyaltyAccount{}).
		Where("user_id = ?", userID).
		Update("points", gorm.Expr("points + ?", points)).
		Error
}

func (r *crmRepository) GetUserByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *crmRepository) GetUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *crmRepository) GetUserByID(userID uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *crmRepository) GetUserByPhone(phone string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("phone_number = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *crmRepository) CreateUser(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *crmRepository) UpdateUser(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *crmRepository) DeleteUser(userID uint) error {
	return r.db.Delete(&domain.User{}, userID).Error
}

func (r *crmRepository) ListCustomers(page, limit int, search string) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64
	offset := (page - 1) * limit

	query := r.db.Model(&domain.User{}).Where("role = ?", "Customer")

	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ? OR phone_number ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?", pattern, pattern, pattern, pattern, pattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("LoyaltyAccount").Limit(limit).Offset(offset).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
