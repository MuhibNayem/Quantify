package repository

import (
	"inventory/backend/internal/domain"
	"gorm.io/gorm"
)

type CRMRepository interface {
	CreateLoyaltyAccount(account *domain.LoyaltyAccount) error
	GetLoyaltyAccountByUserID(userID uint) (*domain.LoyaltyAccount, error)
	UpdateLoyaltyAccount(account *domain.LoyaltyAccount) error
	GetUserByUsername(username string) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	GetUserByID(userID uint) (*domain.User, error)
	GetUserByPhone(phone string) (*domain.User, error)
	CreateUser(user *domain.User) error
	UpdateUser(user *domain.User) error
	DeleteUser(userID uint) error
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
