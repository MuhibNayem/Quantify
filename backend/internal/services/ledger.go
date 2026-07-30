// Package services — ledger.go implements a double-entry general ledger.
//
// Every financial event posts a JournalEntry whose lines sum to zero
// (Σ debits == Σ credits). This invariant is enforced on every Post and is the
// mathematical guarantee that the books always balance — the foundation for a
// correct trial balance, income statement, and balance sheet.
package services

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

var zero = decimal.Zero

// LedgerService posts balanced journal entries and produces financial reports.
type LedgerService struct {
	DB *gorm.DB
}

func NewLedgerService(db *gorm.DB) *LedgerService {
	return &LedgerService{DB: db}
}

// Post validates that a journal entry is balanced (Σ debit == Σ credit, both > 0)
// and persists it. It returns an error — and posts nothing — if unbalanced.
// Call within an outer transaction to keep the ledger atomic with the source event.
func (s *LedgerService) Post(tx *gorm.DB, entry *domain.JournalEntry) error {
	if len(entry.Lines) < 2 {
		return fmt.Errorf("journal entry must have at least 2 lines")
	}
	totalDebit := zero
	totalCredit := zero
	for i := range entry.Lines {
		l := &entry.Lines[i]
		if l.Debit.IsNegative() || l.Credit.IsNegative() {
			return fmt.Errorf("journal line amounts must be non-negative")
		}
		if l.Debit.IsPositive() && l.Credit.IsPositive() {
			return fmt.Errorf("a journal line cannot be both a debit and a credit")
		}
		if l.Debit.IsZero() && l.Credit.IsZero() {
			return fmt.Errorf("journal line must have a debit or credit amount")
		}
		totalDebit = totalDebit.Add(l.Debit)
		totalCredit = totalCredit.Add(l.Credit)
	}
	if !totalDebit.Equal(totalCredit) {
		return fmt.Errorf("unbalanced entry: debits %s != credits %s", totalDebit, totalCredit)
	}
	if totalDebit.IsZero() {
		return fmt.Errorf("journal entry has zero total")
	}
	if entry.Date.IsZero() {
		entry.Date = time.Now()
	}
	return tx.Create(entry).Error
}

// accountByCode resolves a system account code to its ID (cached within the tx).
func (s *LedgerService) accountID(tx *gorm.DB, code string) (uint, error) {
	var acc domain.Account
	if err := tx.Where("code = ?", code).First(&acc).Error; err != nil {
		return 0, fmt.Errorf("account %s not found: %w", code, err)
	}
	return acc.ID, nil
}

// PostSale records a completed sale. For a cash sale, cash is debited; for a
// credit (unpaid) sale, accounts receivable is debited instead. Revenue is
// always credited, a manual discount (if any) is recorded as a contra-revenue
// (Sales Discounts), and COGS is recognized against inventory.
//
// Sale with a manual discount:
//
//	DR Cash / AR            netTotal
//	DR Sales Discounts      manualDiscount   (contra-revenue)
//	CR Sales Revenue        netTotal + manualDiscount   (gross)
//	DR COGS                 cost
//	CR Inventory            cost
func (s *LedgerService) PostSale(tx *gorm.DB, order *domain.Order, cost float64, manualDiscount float64, onCredit bool, postedBy uint) (*domain.JournalEntry, error) {
	netTotal := decimal.NewFromFloat(order.TotalAmount).Round(4)
	discount := decimal.NewFromFloat(manualDiscount).Round(4)
	grossRevenue := netTotal.Add(discount)
	costDec := decimal.NewFromFloat(cost).Round(4)

	debitAccount := domain.AccountCodeCash
	refPrefix := "RCPT"
	if onCredit {
		debitAccount = domain.AccountCodeAccountsReceivable
		refPrefix = "SO"
	}

	cashID, err := s.accountID(tx, debitAccount)
	if err != nil {
		return nil, err
	}
	revenueID, err := s.accountID(tx, domain.AccountCodeSalesRevenue)
	if err != nil {
		return nil, err
	}
	cogsID, err := s.accountID(tx, domain.AccountCodeCOGS)
	if err != nil {
		return nil, err
	}
	inventoryID, err := s.accountID(tx, domain.AccountCodeInventory)
	if err != nil {
		return nil, err
	}

	entry := &domain.JournalEntry{
		Date:        time.Now(),
		Reference:   fmt.Sprintf("%s-%05d", refPrefix, order.ID),
		Description: fmt.Sprintf("Sale %s", order.OrderNumber),
		SourceType:  "SALE",
		SourceID:    order.ID,
		PostedBy:    postedBy,
		Lines:       []domain.JournalLine{},
	}
	// Cash/AR leg — omitted for a fully-discounted (zero-total) sale.
	if netTotal.IsPositive() {
		entry.Lines = append(entry.Lines,
			domain.JournalLine{AccountID: cashID, Debit: netTotal, Description: debitAccount},
		)
	}
	entry.Lines = append(entry.Lines,
		domain.JournalLine{AccountID: revenueID, Credit: grossRevenue, Description: "Sales Revenue"},
	)
	// Record the manual discount as a contra-revenue (debit to Sales Discounts).
	if discount.IsPositive() {
		discountID, err := s.accountID(tx, domain.AccountCodeSalesDiscounts)
		if err != nil {
			return nil, err
		}
		entry.Lines = append(entry.Lines,
			domain.JournalLine{AccountID: discountID, Debit: discount, Description: "Sales Discounts"},
		)
	}
	// Recognize cost of goods sold only if we have a cost figure.
	if costDec.IsPositive() {
		entry.Lines = append(entry.Lines,
			domain.JournalLine{AccountID: cogsID, Debit: costDec, Description: "COGS"},
			domain.JournalLine{AccountID: inventoryID, Credit: costDec, Description: "Inventory"},
		)
	}
	if err := s.Post(tx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// PostPayment records money received from a customer against accounts receivable.
//
//	DR Cash                  amount
//	CR Accounts Receivable   amount
func (s *LedgerService) PostPayment(tx *gorm.DB, payment *domain.Payment) (*domain.JournalEntry, error) {
	amount := payment.Amount.Round(4)
	cashID, err := s.accountID(tx, domain.AccountCodeCash)
	if err != nil {
		return nil, err
	}
	arID, err := s.accountID(tx, domain.AccountCodeAccountsReceivable)
	if err != nil {
		return nil, err
	}
	entry := &domain.JournalEntry{
		Date:        payment.ReceivedAt,
		Reference:   fmt.Sprintf("PMT-%05d", payment.ID),
		Description: fmt.Sprintf("Payment received (customer %d)", payment.CustomerID),
		SourceType:  "PAYMENT",
		SourceID:    payment.ID,
		PostedBy:    payment.ReceivedBy,
		Lines: []domain.JournalLine{
			{AccountID: cashID, Debit: amount, Description: "Cash"},
			{AccountID: arID, Credit: amount, Description: "Accounts Receivable"},
		},
	}
	if err := s.Post(tx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// PostPurchaseReceipt records stock received from a supplier on credit: inventory
// increases and we now owe the supplier (accounts payable).
//
//	DR Inventory          value
//	CR Accounts Payable   value
func (s *LedgerService) PostPurchaseReceipt(tx *gorm.DB, po *domain.PurchaseOrder, value float64, postedBy uint) (*domain.JournalEntry, error) {
	amount := decimal.NewFromFloat(value).Round(4)
	if !amount.IsPositive() {
		return nil, fmt.Errorf("purchase receipt value must be positive")
	}
	inventoryID, err := s.accountID(tx, domain.AccountCodeInventory)
	if err != nil {
		return nil, err
	}
	apID, err := s.accountID(tx, domain.AccountCodeAccountsPayable)
	if err != nil {
		return nil, err
	}
	entry := &domain.JournalEntry{
		Date:        time.Now(),
		Reference:   fmt.Sprintf("GRN-%05d", po.ID),
		Description: fmt.Sprintf("Goods received (PO #%d)", po.ID),
		SourceType:  "PURCHASE",
		SourceID:    po.ID,
		PostedBy:    postedBy,
		Lines: []domain.JournalLine{
			{AccountID: inventoryID, Debit: amount, Description: "Inventory"},
			{AccountID: apID, Credit: amount, Description: "Accounts Payable"},
		},
	}
	if err := s.Post(tx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// PostSupplierPayment records money paid to a supplier against accounts payable.
//
//	DR Accounts Payable   amount
//	CR Cash               amount
func (s *LedgerService) PostSupplierPayment(tx *gorm.DB, payment *domain.SupplierPayment) (*domain.JournalEntry, error) {
	amount := payment.Amount.Round(4)
	apID, err := s.accountID(tx, domain.AccountCodeAccountsPayable)
	if err != nil {
		return nil, err
	}
	cashID, err := s.accountID(tx, domain.AccountCodeCash)
	if err != nil {
		return nil, err
	}
	entry := &domain.JournalEntry{
		Date:        payment.PaidAt,
		Reference:   fmt.Sprintf("SPMT-%05d", payment.ID),
		Description: fmt.Sprintf("Payment to supplier %d", payment.SupplierID),
		SourceType:  "SUPPLIER_PAYMENT",
		SourceID:    payment.ID,
		PostedBy:    payment.PaidBy,
		Lines: []domain.JournalLine{
			{AccountID: apID, Debit: amount, Description: "Accounts Payable"},
			{AccountID: cashID, Credit: amount, Description: "Cash"},
		},
	}
	if err := s.Post(tx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// TrialBalanceRow is one account's summed debits and credits across all posted entries.
type TrialBalanceRow struct {
	AccountID   uint            `json:"AccountID"`
	Code        string          `json:"Code"`
	Name        string          `json:"Name"`
	Type        string          `json:"Type"`
	TotalDebit  decimal.Decimal `json:"TotalDebit"`
	TotalCredit decimal.Decimal `json:"TotalCredit"`
	// Balance is the net (debit-positive for asset/expense, credit-positive otherwise).
	Balance decimal.Decimal `json:"Balance"`
}

// TrialBalance aggregates all posted journal lines per account. The grand total
// of debits must equal the grand total of credits — the books balance.
func (s *LedgerService) TrialBalance(from, to *time.Time) ([]TrialBalanceRow, error) {
	q := s.DB.Table("journal_lines").
		Select(`accounts.id as account_id, accounts.code, accounts.name, accounts.type,
			COALESCE(SUM(journal_lines.debit), 0) as total_debit,
			COALESCE(SUM(journal_lines.credit), 0) as total_credit`).
		Joins("JOIN accounts ON accounts.id = journal_lines.account_id").
		Joins("JOIN journal_entries ON journal_entries.id = journal_lines.journal_entry_id")
	if from != nil {
		q = q.Where("journal_entries.date >= ?", *from)
	}
	if to != nil {
		q = q.Where("journal_entries.date <= ?", *to)
	}
	q = q.Group("accounts.id, accounts.code, accounts.name, accounts.type").
		Order("accounts.code")

	var rows []TrialBalanceRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	for i := range rows {
		// Net balance: assets/expenses are debit-normal, others credit-normal.
		if rows[i].Type == string(domain.AccountAsset) || rows[i].Type == string(domain.AccountExpense) {
			rows[i].Balance = rows[i].TotalDebit.Sub(rows[i].TotalCredit)
		} else {
			rows[i].Balance = rows[i].TotalCredit.Sub(rows[i].TotalDebit)
		}
	}
	return rows, nil
}

// IncomeStatement summarizes revenue and expenses for a period.
type IncomeStatement struct {
	Revenue   decimal.Decimal `json:"Revenue"`
	COGS      decimal.Decimal `json:"COGS"`
	GrossProfit decimal.Decimal `json:"GrossProfit"`
	Expenses  decimal.Decimal `json:"Expenses"` // non-COGS expenses
	NetIncome decimal.Decimal `json:"NetIncome"`
}

// IncomeStatement computes revenue, COGS, and net income for the period.
func (s *LedgerService) IncomeStatement(from, to *time.Time) (*IncomeStatement, error) {
	rows, err := s.TrialBalance(from, to)
	if err != nil {
		return nil, err
	}
	stmt := &IncomeStatement{}
	cogs := decimal.Zero
	expenses := decimal.Zero
	for _, r := range rows {
		switch r.Type {
		case string(domain.AccountRevenue):
			stmt.Revenue = stmt.Revenue.Add(r.Balance)
		case string(domain.AccountExpense):
			if r.Code == domain.AccountCodeCOGS {
				cogs = cogs.Add(r.Balance)
			} else {
				expenses = expenses.Add(r.Balance)
			}
		}
	}
	stmt.COGS = cogs
	stmt.Expenses = expenses
	stmt.GrossProfit = stmt.Revenue.Sub(stmt.COGS)
	stmt.NetIncome = stmt.GrossProfit.Sub(stmt.Expenses)
	return stmt, nil
}

// BalanceSheet summarizes the accounting equation: Assets = Liabilities + Equity.
type BalanceSheet struct {
	Assets        decimal.Decimal `json:"Assets"`
	Liabilities   decimal.Decimal `json:"Liabilities"`
	Equity        decimal.Decimal `json:"Equity"`
	RetainedEarnings decimal.Decimal `json:"RetainedEarnings"` // net income rolled into equity
	TotalLiabilitiesAndEquity decimal.Decimal `json:"TotalLiabilitiesAndEquity"`
	Balanced      bool            `json:"Balanced"`
}

// BalanceSheet computes assets, liabilities, and equity as of the period end.
// Net income for the period is included in equity as retained earnings so the
// equation Assets = Liabilities + Equity holds.
func (s *LedgerService) BalanceSheet(from, to *time.Time) (*BalanceSheet, error) {
	rows, err := s.TrialBalance(from, to)
	if err != nil {
		return nil, err
	}
	bs := &BalanceSheet{}
	for _, r := range rows {
		switch r.Type {
		case string(domain.AccountAsset):
			bs.Assets = bs.Assets.Add(r.Balance)
		case string(domain.AccountLiability):
			bs.Liabilities = bs.Liabilities.Add(r.Balance)
		case string(domain.AccountEquity):
			bs.Equity = bs.Equity.Add(r.Balance)
		}
	}
	// Roll net income into equity (retained earnings) to balance the equation.
	income, err := s.IncomeStatement(from, to)
	if err != nil {
		return nil, err
	}
	bs.RetainedEarnings = income.NetIncome
	bs.Equity = bs.Equity.Add(income.NetIncome)
	bs.TotalLiabilitiesAndEquity = bs.Liabilities.Add(bs.Equity)
	bs.Balanced = bs.Assets.Equal(bs.TotalLiabilitiesAndEquity)
	return bs, nil
}
