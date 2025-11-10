package repository

import (
	"inventory/backend/internal/domain"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	CreateTransaction(transaction *domain.Transaction) error
	GetTransactionByTranID(tranID string) (*domain.Transaction, error)
	UpdateTransaction(transaction *domain.Transaction) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) CreateTransaction(transaction *domain.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *paymentRepository) GetTransactionByTranID(tranID string) (*domain.Transaction, error) {
	var transaction domain.Transaction
	if err := r.db.Where("order_id = ?", tranID).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *paymentRepository) UpdateTransaction(transaction *domain.Transaction) error {
	return r.db.Save(transaction).Error
}
