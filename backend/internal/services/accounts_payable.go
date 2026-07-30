package services

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"inventory/backend/internal/domain"
)

// AccountsPayableService manages money owed to suppliers (from received POs)
// and the payments that settle it, keeping the ledger balanced.
type AccountsPayableService struct {
	DB     *gorm.DB
	Ledger *LedgerService
}

func NewAccountsPayableService(db *gorm.DB) *AccountsPayableService {
	return &AccountsPayableService{DB: db, Ledger: NewLedgerService(db)}
}

// RecordSupplierPaymentRequest is a payment made to a supplier. If Allocations
// is empty, the amount is applied FIFO against the supplier's outstanding AP.
type RecordSupplierPaymentRequest struct {
	SupplierID  uint
	Amount      float64
	Method      string
	Notes       string
	Allocations []SupplierPaymentAllocationInput
}

type SupplierPaymentAllocationInput struct {
	AccountsPayableID uint
	Amount            float64
}

// RecordPayment records a payment to a supplier, applies it to outstanding AP,
// and posts the balanced journal entry (DR Accounts Payable, CR Cash).
func (s *AccountsPayableService) RecordPayment(req RecordSupplierPaymentRequest, paidBy uint) (*domain.SupplierPayment, error) {
	amount := decimal.NewFromFloat(req.Amount).Round(4)
	if !amount.IsPositive() {
		return nil, fmt.Errorf("payment amount must be positive")
	}

	var payment *domain.SupplierPayment
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		payment = &domain.SupplierPayment{
			SupplierID: req.SupplierID,
			Amount:     amount,
			Method:     req.Method,
			PaidAt:     time.Now(),
			PaidBy:     paidBy,
			Notes:      req.Notes,
		}
		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		entry, err := s.Ledger.PostSupplierPayment(tx, payment)
		if err != nil {
			return fmt.Errorf("failed to post supplier payment to ledger: %w", err)
		}
		payment.JournalEntryID = entry.ID

		allocs, err := s.buildAllocations(tx, req, amount)
		if err != nil {
			return err
		}

		remaining := amount
		for _, alloc := range allocs {
			if remaining.IsZero() {
				break
			}
			apply := decimal.NewFromFloat(alloc.Amount).Round(4)
			if apply.GreaterThan(remaining) {
				apply = remaining
			}
			if err := s.applyToAP(tx, payment.ID, alloc.AccountsPayableID, apply); err != nil {
				return err
			}
			remaining = remaining.Sub(apply)
		}
		return tx.Save(payment).Error
	})
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *AccountsPayableService) buildAllocations(tx *gorm.DB, req RecordSupplierPaymentRequest, amount decimal.Decimal) ([]SupplierPaymentAllocationInput, error) {
	if len(req.Allocations) > 0 {
		return req.Allocations, nil
	}
	var aps []domain.AccountsPayable
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("supplier_id = ? AND status != ?", req.SupplierID, "PAID").
		Order("created_at ASC").Find(&aps).Error; err != nil {
		return nil, err
	}
	var plan []SupplierPaymentAllocationInput
	remaining := amount
	for _, ap := range aps {
		if remaining.IsZero() {
			break
		}
		outstanding := ap.Outstanding()
		apply := outstanding
		if apply.GreaterThan(remaining) {
			apply = remaining
		}
		plan = append(plan, SupplierPaymentAllocationInput{AccountsPayableID: ap.ID, Amount: apply.InexactFloat64()})
		remaining = remaining.Sub(apply)
	}
	return plan, nil
}

func (s *AccountsPayableService) applyToAP(tx *gorm.DB, paymentID uint, apID uint, amount decimal.Decimal) error {
	var ap domain.AccountsPayable
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ap, apID).Error; err != nil {
		return fmt.Errorf("accounts payable %d not found: %w", apID, err)
	}
	if amount.GreaterThan(ap.Outstanding()) {
		return fmt.Errorf("allocation %s exceeds outstanding %s on AP %d", amount, ap.Outstanding(), apID)
	}
	ap.PaidAmount = ap.PaidAmount.Add(amount)
	if ap.PaidAmount.GreaterThanOrEqual(ap.Amount) {
		ap.Status = "PAID"
	} else {
		ap.Status = "PARTIAL"
	}
	if err := tx.Save(&ap).Error; err != nil {
		return err
	}
	return tx.Create(&domain.SupplierPaymentAllocation{
		SupplierPaymentID: paymentID,
		AccountsPayableID: apID,
		Amount:            amount,
	}).Error
}

// OutstandingBySupplier returns the total outstanding AP for a supplier.
func (s *AccountsPayableService) OutstandingBySupplier(supplierID uint) (decimal.Decimal, error) {
	var aps []domain.AccountsPayable
	if err := s.DB.Where("supplier_id = ? AND status != ?", supplierID, "PAID").Find(&aps).Error; err != nil {
		return decimal.Zero, err
	}
	total := decimal.Zero
	for _, ap := range aps {
		total = total.Add(ap.Outstanding())
	}
	return total, nil
}

// APAgingBucket is one aging band of outstanding payables.
type APAgingBucket struct {
	Label  string          `json:"Label"`
	Amount decimal.Decimal `json:"Amount"`
	Count  int             `json:"Count"`
}

// APAgingReport buckets all outstanding payables by age (days since receipt).
func (s *AccountsPayableService) APAgingReport() ([]APAgingBucket, error) {
	var aps []domain.AccountsPayable
	if err := s.DB.Preload("PurchaseOrder").Where("status != ?", "PAID").Find(&aps).Error; err != nil {
		return nil, err
	}
	buckets := []APAgingBucket{
		{Label: "Current", Amount: decimal.Zero},
		{Label: "1-30 days", Amount: decimal.Zero},
		{Label: "31-60 days", Amount: decimal.Zero},
		{Label: "61-90 days", Amount: decimal.Zero},
		{Label: "90+ days", Amount: decimal.Zero},
	}
	now := time.Now()
	for _, ap := range aps {
		outstanding := ap.Outstanding()
		days := 0
		if !ap.PurchaseOrder.OrderDate.IsZero() {
			days = int(now.Sub(ap.PurchaseOrder.OrderDate).Hours() / 24)
		}
		var idx int
		switch {
		case days <= 0:
			idx = 0
		case days <= 30:
			idx = 1
		case days <= 60:
			idx = 2
		case days <= 90:
			idx = 3
		default:
			idx = 4
		}
		buckets[idx].Amount = buckets[idx].Amount.Add(outstanding)
		buckets[idx].Count++
	}
	return buckets, nil
}

// ListOutstanding returns all non-paid AP entries (optionally for one supplier).
func (s *AccountsPayableService) ListOutstanding(supplierID *uint) ([]domain.AccountsPayable, error) {
	q := s.DB.Preload("Supplier").Preload("PurchaseOrder").Where("status != ?", "PAID")
	if supplierID != nil {
		q = q.Where("supplier_id = ?", *supplierID)
	}
	var aps []domain.AccountsPayable
	err := q.Order("created_at ASC").Find(&aps).Error
	return aps, err
}
