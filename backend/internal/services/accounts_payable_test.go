package services

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func setupAPTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(
		&domain.Account{}, &domain.JournalEntry{}, &domain.JournalLine{},
		&domain.AccountsPayable{}, &domain.SupplierPayment{}, &domain.SupplierPaymentAllocation{},
		&domain.PurchaseOrder{}, &domain.Supplier{},
	)
	if err != nil {
		t.Fatal(err)
	}
	accounts := []domain.Account{
		{Code: domain.AccountCodeCash, Name: "Cash", Type: domain.AccountAsset},
		{Code: domain.AccountCodeInventory, Name: "Inventory", Type: domain.AccountAsset},
		{Code: domain.AccountCodeAccountsPayable, Name: "AP", Type: domain.AccountLiability},
	}
	for i := range accounts {
		if err := db.Create(&accounts[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

// TestPurchaseReceipt_FullCycle proves receiving stock creates accounts payable
// and that paying the supplier settles it, keeping the ledger balanced throughout.
func TestPurchaseReceipt_FullCycle(t *testing.T) {
	db := setupAPTestDB(t)
	ledger := NewLedgerService(db)
	apSvc := NewAccountsPayableService(db)

	// 1. Receive goods worth $500 on credit → DR Inventory 500 / CR AP 500.
	po := &domain.PurchaseOrder{SupplierID: 1, Status: "SENT", OrderDate: time.Now()}
	po.ID = 1
	entry, err := ledger.PostPurchaseReceipt(db, po, 500.0, 1)
	if err != nil {
		t.Fatalf("purchase receipt should post: %v", err)
	}

	// 2. Record the payable (we owe $500).
	ap := domain.AccountsPayable{
		SupplierID: 1, PurchaseOrderID: po.ID,
		Amount: decimal.NewFromInt(500), PaidAmount: decimal.Zero,
		Status: "UNPAID", JournalEntryID: entry.ID,
	}
	if err := db.Create(&ap).Error; err != nil {
		t.Fatal(err)
	}
	outstanding, _ := apSvc.OutstandingBySupplier(1)
	if !outstanding.Equal(decimal.NewFromInt(500)) {
		t.Fatalf("outstanding expected 500, got %s", outstanding)
	}
	assertAPBalanced(t, ledger)

	// 3. Pay $200 → PARTIAL, $300 outstanding.
	if _, err := apSvc.RecordPayment(RecordSupplierPaymentRequest{SupplierID: 1, Amount: 200, Method: "bank"}, 1); err != nil {
		t.Fatalf("partial supplier payment failed: %v", err)
	}
	db.First(&ap, ap.ID)
	if ap.Status != "PARTIAL" {
		t.Fatalf("expected PARTIAL, got %s", ap.Status)
	}
	outstanding, _ = apSvc.OutstandingBySupplier(1)
	if !outstanding.Equal(decimal.NewFromInt(300)) {
		t.Fatalf("outstanding after partial expected 300, got %s", outstanding)
	}
	assertAPBalanced(t, ledger)

	// 4. Pay remaining $300 → PAID, $0 outstanding.
	if _, err := apSvc.RecordPayment(RecordSupplierPaymentRequest{SupplierID: 1, Amount: 300, Method: "bank"}, 1); err != nil {
		t.Fatalf("final supplier payment failed: %v", err)
	}
	db.First(&ap, ap.ID)
	if ap.Status != "PAID" {
		t.Fatalf("expected PAID, got %s", ap.Status)
	}
	outstanding, _ = apSvc.OutstandingBySupplier(1)
	if !outstanding.IsZero() {
		t.Fatalf("outstanding after settlement expected 0, got %s", outstanding)
	}
	assertAPBalanced(t, ledger)
}

func assertAPBalanced(t *testing.T, ledger *LedgerService) {
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
