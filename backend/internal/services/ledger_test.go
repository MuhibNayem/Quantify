package services

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func setupLedgerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.Account{}, &domain.JournalEntry{}, &domain.JournalLine{}); err != nil {
		t.Fatal(err)
	}
	// Seed minimal chart of accounts.
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

func accID(t *testing.T, db *gorm.DB, code string) uint {
	t.Helper()
	var a domain.Account
	if err := db.Where("code = ?", code).First(&a).Error; err != nil {
		t.Fatal(err)
	}
	return a.ID
}

func TestLedger_RejectsUnbalancedEntry(t *testing.T) {
	db := setupLedgerTestDB(t)
	svc := NewLedgerService(db)
	cash := accID(t, db, domain.AccountCodeCash)
	revenue := accID(t, db, domain.AccountCodeSalesRevenue)

	entry := &domain.JournalEntry{
		Date: time.Now(), Reference: "BAD-1", SourceType: "SALE",
		Lines: []domain.JournalLine{
			{AccountID: cash, Debit: decimal.NewFromInt(100)},
			{AccountID: revenue, Credit: decimal.NewFromInt(90)}, // != 100
		},
	}
	if err := svc.Post(db, entry); err == nil {
		t.Fatal("expected unbalanced entry to be rejected")
	}
}

func TestLedger_PostsBalancedSale(t *testing.T) {
	db := setupLedgerTestDB(t)
	svc := NewLedgerService(db)
	order := &domain.Order{OrderNumber: "ORD-1", TotalAmount: 100.0}
	order.ID = 1

	entry, err := svc.PostSale(db, order, 60.0, 0, false, 1)
	if err != nil {
		t.Fatalf("balanced sale should post: %v", err)
	}
	// Verify Σ debit == Σ credit on the persisted entry.
	var lines []domain.JournalLine
	db.Where("journal_entry_id = ?", entry.ID).Find(&lines)
	d, c := decimal.Zero, decimal.Zero
	for _, l := range lines {
		d = d.Add(l.Debit)
		c = c.Add(l.Credit)
	}
	if !d.Equal(c) {
		t.Fatalf("persisted entry unbalanced: debit %s credit %s", d, c)
	}
	// Cash sale: DR Cash 100, CR Revenue 100, DR COGS 60, CR Inventory 60 → 160 each.
	if !d.Equal(decimal.NewFromInt(160)) {
		t.Fatalf("expected total 160, got %s", d)
	}
}

func TestLedger_CreditSaleDebitsAR(t *testing.T) {
	db := setupLedgerTestDB(t)
	svc := NewLedgerService(db)
	order := &domain.Order{OrderNumber: "ORD-2", TotalAmount: 200.0}
	order.ID = 2

	_, err := svc.PostSale(db, order, 0, 0, true, 1) // on credit, no cost
	if err != nil {
		t.Fatalf("credit sale should post: %v", err)
	}
	// AR should be debited 200.
	arID := accID(t, db, domain.AccountCodeAccountsReceivable)
	var line domain.JournalLine
	if err := db.Where("account_id = ?", arID).First(&line).Error; err != nil {
		t.Fatal(err)
	}
	if !line.Debit.Equal(decimal.NewFromInt(200)) {
		t.Fatalf("AR debit expected 200, got %s", line.Debit)
	}
}

func TestLedger_TrialBalanceBalances(t *testing.T) {
	db := setupLedgerTestDB(t)
	svc := NewLedgerService(db)

	// Post several entries: a cash sale, a credit sale, a payment.
	o1 := &domain.Order{OrderNumber: "O1", TotalAmount: 100.0}
	o1.ID = 1
	if _, err := svc.PostSale(db, o1, 60.0, 0, false, 1); err != nil {
		t.Fatal(err)
	}
	o2 := &domain.Order{OrderNumber: "O2", TotalAmount: 50.0}
	o2.ID = 2
	if _, err := svc.PostSale(db, o2, 0, 0, true, 1); err != nil {
		t.Fatal(err)
	}
	pmt := &domain.Payment{CustomerID: 1, Amount: decimal.NewFromInt(50), ReceivedAt: time.Now(), ReceivedBy: 1}
	pmt.ID = 1
	if _, err := svc.PostPayment(db, pmt); err != nil {
		t.Fatal(err)
	}

	rows, err := svc.TrialBalance(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	totalDebit, totalCredit := decimal.Zero, decimal.Zero
	for _, r := range rows {
		totalDebit = totalDebit.Add(r.TotalDebit)
		totalCredit = totalCredit.Add(r.TotalCredit)
	}
	if !totalDebit.Equal(totalCredit) {
		t.Fatalf("trial balance does not balance: debit %s credit %s", totalDebit, totalCredit)
	}
	if totalDebit.IsZero() {
		t.Fatal("trial balance is empty")
	}
}
