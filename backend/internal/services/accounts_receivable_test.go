package services

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func setupARTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(
		&domain.Account{}, &domain.JournalEntry{}, &domain.JournalLine{},
		&domain.AccountsReceivable{}, &domain.Payment{}, &domain.PaymentAllocation{},
		&domain.Order{}, &domain.User{},
	)
	if err != nil {
		t.Fatal(err)
	}
	accounts := []domain.Account{
		{Code: domain.AccountCodeCash, Name: "Cash", Type: domain.AccountAsset},
		{Code: domain.AccountCodeAccountsReceivable, Name: "AR", Type: domain.AccountAsset},
		{Code: domain.AccountCodeInventory, Name: "Inventory", Type: domain.AccountAsset},
		{Code: domain.AccountCodeSalesRevenue, Name: "Revenue", Type: domain.AccountRevenue},
		{Code: domain.AccountCodeCOGS, Name: "COGS", Type: domain.AccountExpense},
	}
	for i := range accounts {
		if err := db.Create(&accounts[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

// TestCreditSale_FullCycle proves an unpaid delivery is tracked as accounts
// receivable and is mathematically ledger-bound from sale through settlement.
func TestCreditSale_FullCycle(t *testing.T) {
	db := setupARTestDB(t)
	ledger := NewLedgerService(db)
	arSvc := NewAccountsReceivableService(db)

	// 1. Credit sale (unpaid delivery) of $100, cost $60.
	order := &domain.Order{OrderNumber: "ORD-C1", TotalAmount: 100.0, CustomerID: ptrUint(1)}
	order.ID = 1
	entry, err := ledger.PostSale(db, order, 60.0, 0, true, 1)
	if err != nil {
		t.Fatalf("credit sale should post: %v", err)
	}

	// 2. Record the receivable (customer owes $100).
	ar := domain.AccountsReceivable{
		CustomerID: 1, OrderID: order.ID,
		Amount: decimal.NewFromInt(100), PaidAmount: decimal.Zero,
		Status: "UNPAID", JournalEntryID: entry.ID,
	}
	if err := db.Create(&ar).Error; err != nil {
		t.Fatal(err)
	}
	outstanding, _ := arSvc.OutstandingByCustomer(1)
	if !outstanding.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("outstanding expected 100, got %s", outstanding)
	}

	// 3. Books balance after the sale.
	assertBalanced(t, ledger)

	// 4. Partial payment of $40.
	if _, err := arSvc.RecordPayment(RecordPaymentRequest{CustomerID: 1, Amount: 40, Method: "cash"}, 1); err != nil {
		t.Fatalf("partial payment failed: %v", err)
	}
	db.First(&ar, ar.ID)
	if ar.Status != "PARTIAL" {
		t.Fatalf("expected PARTIAL, got %s", ar.Status)
	}
	outstanding, _ = arSvc.OutstandingByCustomer(1)
	if !outstanding.Equal(decimal.NewFromInt(60)) {
		t.Fatalf("outstanding after partial expected 60, got %s", outstanding)
	}
	assertBalanced(t, ledger)

	// 5. Final payment of $60 settles it.
	if _, err := arSvc.RecordPayment(RecordPaymentRequest{CustomerID: 1, Amount: 60, Method: "cash"}, 1); err != nil {
		t.Fatalf("final payment failed: %v", err)
	}
	db.First(&ar, ar.ID)
	if ar.Status != "PAID" {
		t.Fatalf("expected PAID, got %s", ar.Status)
	}
	outstanding, _ = arSvc.OutstandingByCustomer(1)
	if !outstanding.IsZero() {
		t.Fatalf("outstanding after settlement expected 0, got %s", outstanding)
	}

	// 6. Books still balance after all payments.
	assertBalanced(t, ledger)

	// 7. AR aging reflects nothing outstanding.
	buckets, err := arSvc.ARAgingReport()
	if err != nil {
		t.Fatal(err)
	}
	total := decimal.Zero
	for _, b := range buckets {
		total = total.Add(b.Amount)
	}
	if !total.IsZero() {
		t.Fatalf("aging total expected 0 after settlement, got %s", total)
	}
}

func assertBalanced(t *testing.T, ledger *LedgerService) {
	t.Helper()
	rows, err := ledger.TrialBalance(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	d, c := decimal.Zero, decimal.Zero
	for _, r := range rows {
		d = d.Add(r.TotalDebit)
		c = c.Add(r.TotalCredit)
	}
	if !d.Equal(c) {
		t.Fatalf("ledger unbalanced: debit %s credit %s", d, c)
	}
}

var _ = time.Now
