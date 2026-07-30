package services

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

type Ledger interface {
	Post(tx *gorm.DB, entry *domain.JournalEntry) error
	PostSale(tx *gorm.DB, order *domain.Order, cost float64, manualDiscount float64, onCredit bool, postedBy uint) (*domain.JournalEntry, error)
	PostPayment(tx *gorm.DB, payment *domain.Payment) (*domain.JournalEntry, error)
	PostPurchaseReceipt(tx *gorm.DB, po *domain.PurchaseOrder, value float64, postedBy uint) (*domain.JournalEntry, error)
	PostSupplierPayment(tx *gorm.DB, payment *domain.SupplierPayment) (*domain.JournalEntry, error)
	TrialBalance(from, to *time.Time) ([]TrialBalanceRow, error)
	IncomeStatement(from, to *time.Time) (*IncomeStatement, error)
	BalanceSheet(from, to *time.Time) (*BalanceSheet, error)
}

type AccountsReceivable interface {
	RecordPayment(req RecordPaymentRequest, receivedBy uint) (*domain.Payment, error)
	OutstandingByCustomer(customerID uint) (decimal.Decimal, error)
	ARAgingReport() ([]ARAgingBucket, error)
	ListOutstanding(customerID *uint) ([]domain.AccountsReceivable, error)
}

type AccountsPayable interface {
	RecordPayment(req RecordSupplierPaymentRequest, paidBy uint) (*domain.SupplierPayment, error)
	OutstandingBySupplier(supplierID uint) (decimal.Decimal, error)
	APAgingReport() ([]APAgingBucket, error)
	ListOutstanding(supplierID *uint) ([]domain.AccountsPayable, error)
}

var (
	_ Ledger             = (*LedgerService)(nil)
	_ AccountsReceivable = (*AccountsReceivableService)(nil)
	_ AccountsPayable    = (*AccountsPayableService)(nil)
)
