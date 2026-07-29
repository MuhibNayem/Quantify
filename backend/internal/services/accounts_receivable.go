package services

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"inventory/backend/internal/domain"
)

// AccountsReceivableService manages credit sales (money owed) and the payments
// that settle them, keeping the ledger balanced at every step.
type AccountsReceivableService struct {
	DB     *gorm.DB
	Ledger *LedgerService
}

func NewAccountsReceivableService(db *gorm.DB) *AccountsReceivableService {
	return &AccountsReceivableService{DB: db, Ledger: NewLedgerService(db)}
}

// RecordPaymentRequest is a payment received from a customer. If Allocations is
// empty, the amount is applied FIFO against the customer's outstanding AR.
type RecordPaymentRequest struct {
	CustomerID  uint
	Amount      float64
	Method      string
	Notes       string
	Allocations []PaymentAllocationInput // optional explicit AR allocation
}

type PaymentAllocationInput struct {
	AccountsReceivableID uint
	Amount               float64
}

// RecordPayment records a customer payment, applies it to outstanding AR, and
// posts the balanced journal entry (DR Cash, CR Accounts Receivable).
func (s *AccountsReceivableService) RecordPayment(req RecordPaymentRequest, receivedBy uint) (*domain.Payment, error) {
	amount := decimal.NewFromFloat(req.Amount).Round(4)
	if !amount.IsPositive() {
		return nil, fmt.Errorf("payment amount must be positive")
	}

	var payment *domain.Payment
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		payment = &domain.Payment{
			CustomerID: req.CustomerID,
			Amount:     amount,
			Method:     req.Method,
			ReceivedAt: time.Now(),
			ReceivedBy: receivedBy,
			Notes:      req.Notes,
		}
		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		// Post DR Cash / CR AR.
		entry, err := s.Ledger.PostPayment(tx, payment)
		if err != nil {
			return fmt.Errorf("failed to post payment to ledger: %w", err)
		}
		payment.JournalEntryID = entry.ID

		// Determine allocations: explicit, or FIFO against outstanding AR.
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
			if err := s.applyToAR(tx, payment.ID, alloc.AccountsReceivableID, apply); err != nil {
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

// buildAllocations returns explicit allocations or a FIFO plan over outstanding AR.
func (s *AccountsReceivableService) buildAllocations(tx *gorm.DB, req RecordPaymentRequest, amount decimal.Decimal) ([]PaymentAllocationInput, error) {
	if len(req.Allocations) > 0 {
		return req.Allocations, nil
	}
	// FIFO: oldest unpaid AR first.
	var ars []domain.AccountsReceivable
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("customer_id = ? AND status != ?", req.CustomerID, "PAID").
		Order("created_at ASC").Find(&ars).Error; err != nil {
		return nil, err
	}
	var plan []PaymentAllocationInput
	remaining := amount
	for _, ar := range ars {
		if remaining.IsZero() {
			break
		}
		outstanding := ar.Outstanding()
		apply := outstanding
		if apply.GreaterThan(remaining) {
			apply = remaining
		}
		plan = append(plan, PaymentAllocationInput{AccountsReceivableID: ar.ID, Amount: apply.InexactFloat64()})
		remaining = remaining.Sub(apply)
	}
	return plan, nil
}

// applyToAR applies an amount to a specific AR entry, updating its paid amount
// and status, and records the allocation.
func (s *AccountsReceivableService) applyToAR(tx *gorm.DB, paymentID uint, arID uint, amount decimal.Decimal) error {
	var ar domain.AccountsReceivable
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ar, arID).Error; err != nil {
		return fmt.Errorf("accounts receivable %d not found: %w", arID, err)
	}
	if amount.GreaterThan(ar.Outstanding()) {
		return fmt.Errorf("allocation %s exceeds outstanding %s on AR %d", amount, ar.Outstanding(), arID)
	}
	ar.PaidAmount = ar.PaidAmount.Add(amount)
	if ar.PaidAmount.GreaterThanOrEqual(ar.Amount) {
		ar.Status = "PAID"
	} else {
		ar.Status = "PARTIAL"
	}
	if err := tx.Save(&ar).Error; err != nil {
		return err
	}
	return tx.Create(&domain.PaymentAllocation{
		PaymentID:            paymentID,
		AccountsReceivableID: arID,
		Amount:               amount,
	}).Error
}

// OutstandingByCustomer returns the total outstanding AR for a customer.
func (s *AccountsReceivableService) OutstandingByCustomer(customerID uint) (decimal.Decimal, error) {
	var ars []domain.AccountsReceivable
	if err := s.DB.Where("customer_id = ? AND status != ?", customerID, "PAID").Find(&ars).Error; err != nil {
		return decimal.Zero, err
	}
	total := decimal.Zero
	for _, ar := range ars {
		total = total.Add(ar.Outstanding())
	}
	return total, nil
}

// ARAgingBucket is one aging band of outstanding receivables.
type ARAgingBucket struct {
	Label   string          `json:"Label"`   // Current, 1-30, 31-60, 61-90, 90+
	Amount  decimal.Decimal `json:"Amount"`
	Count   int             `json:"Count"`
}

// ARAgingReport buckets all outstanding receivables by age (days since the order).
func (s *AccountsReceivableService) ARAgingReport() ([]ARAgingBucket, error) {
	var ars []domain.AccountsReceivable
	if err := s.DB.Preload("Order").Where("status != ?", "PAID").Find(&ars).Error; err != nil {
		return nil, err
	}
	buckets := []ARAgingBucket{
		{Label: "Current", Amount: decimal.Zero},
		{Label: "1-30 days", Amount: decimal.Zero},
		{Label: "31-60 days", Amount: decimal.Zero},
		{Label: "61-90 days", Amount: decimal.Zero},
		{Label: "90+ days", Amount: decimal.Zero},
	}
	now := time.Now()
	for _, ar := range ars {
		outstanding := ar.Outstanding()
		days := 0
		if !ar.Order.OrderDate.IsZero() {
			days = int(now.Sub(ar.Order.OrderDate).Hours() / 24)
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

// ListOutstanding returns all non-paid AR entries (optionally for one customer).
func (s *AccountsReceivableService) ListOutstanding(customerID *uint) ([]domain.AccountsReceivable, error) {
	q := s.DB.Preload("Customer").Preload("Order").Where("status != ?", "PAID")
	if customerID != nil {
		q = q.Where("customer_id = ?", *customerID)
	}
	var ars []domain.AccountsReceivable
	err := q.Order("created_at ASC").Find(&ars).Error
	return ars, err
}
