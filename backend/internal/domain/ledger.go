package domain

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AccountType is one of the five accounting categories. The type determines the
// account's normal balance (whether it increases on debit or credit).
type AccountType string

const (
	AccountAsset     AccountType = "ASSET"
	AccountLiability AccountType = "LIABILITY"
	AccountEquity    AccountType = "EQUITY"
	AccountRevenue   AccountType = "REVENUE"
	AccountExpense   AccountType = "EXPENSE"
)

// NormalBalance returns "DEBIT" for asset/expense accounts, "CREDIT" otherwise.
func (t AccountType) NormalBalance() string {
	switch t {
	case AccountAsset, AccountExpense:
		return "DEBIT"
	default: // LIABILITY, EQUITY, REVENUE
		return "CREDIT"
	}
}

// Well-known system account codes used by the automated posting engine.
const (
	AccountCodeCash                    = "1000"
	AccountCodeAccountsReceivable      = "1100"
	AccountCodeInventory               = "1200"
	AccountCodeAccountsPayable         = "2000"
	AccountCodeSalesTaxPayable         = "2100" // liability: tax collected at checkout
	AccountCodeSalesRevenue            = "4000"
	AccountCodeSalesDiscounts          = "4100" // contra-revenue: manual/loyalty discounts
	AccountCodeSalesReturns            = "4200" // contra-revenue: customer refunds
	AccountCodeInventoryAdjustmentGain = "4900" // other income: stock gains
	AccountCodeCOGS                    = "5000"
	AccountCodeShrinkageExpense        = "5100" // expense: damage/loss write-offs
)

// Account is a single ledger account in the chart of accounts.
type Account struct {
	gorm.Model
	Code        string      `gorm:"uniqueIndex;not null" json:"Code"` // "1000", "4000"
	Name        string      `gorm:"not null" json:"Name"`             // "Cash", "Sales Revenue"
	Type        AccountType `gorm:"not null;index" json:"Type"`
	Description string      `json:"Description"`
	IsActive    bool        `gorm:"default:true" json:"IsActive"`
	System      bool        `gorm:"default:false" json:"System"` // auto-created, cannot delete
}

// JournalEntry is a balanced accounting transaction (the header). The sum of its
// lines' debits must equal the sum of its credits — enforced by the ledger
// service, never violated.
type JournalEntry struct {
	gorm.Model
	Date        time.Time     `gorm:"index;not null" json:"Date"`
	Reference   string        `gorm:"index" json:"Reference"` // "SO-00001", "PMT-00001"
	Description string        `json:"Description"`
	SourceType  string        `gorm:"index" json:"SourceType"` // SALE, PAYMENT, PURCHASE, ADJUSTMENT
	SourceID    uint          `json:"SourceID"`                // link to Order/Payment/etc.
	PostedBy    uint          `json:"PostedBy"`
	Lines       []JournalLine `gorm:"foreignKey:JournalEntryID" json:"Lines"`
}

// JournalLine is one leg of a journal entry — a debit OR a credit to an account.
// Amounts use shopspring/decimal so the books balance exactly (no float drift).
type JournalLine struct {
	gorm.Model
	JournalEntryID uint            `gorm:"index;not null" json:"JournalEntryID"`
	AccountID      uint            `gorm:"index;not null" json:"AccountID"`
	Account        Account         `json:"Account"`
	Debit          decimal.Decimal `gorm:"type:decimal(19,4);default:0" json:"Debit"`
	Credit         decimal.Decimal `gorm:"type:decimal(19,4);default:0" json:"Credit"`
	Description    string          `json:"Description"`
}

// AccountsReceivable tracks money a customer owes from an unpaid (credit)
// delivery. Settled over time via Payments.
type AccountsReceivable struct {
	gorm.Model
	CustomerID     uint            `gorm:"index;not null" json:"CustomerID"`
	Customer       User            `json:"Customer"`
	OrderID        uint            `gorm:"index" json:"OrderID"`
	Order          Order           `json:"Order"`
	Amount         decimal.Decimal `gorm:"type:decimal(19,4);not null" json:"Amount"` // original amount owed
	PaidAmount     decimal.Decimal `gorm:"type:decimal(19,4);default:0" json:"PaidAmount"`
	Status         string          `gorm:"index;default:'UNPAID'" json:"Status"` // UNPAID, PARTIAL, PAID
	DueDate        *time.Time      `json:"DueDate"`
	JournalEntryID uint            `json:"JournalEntryID"` // entry that created this AR
}

// Outstanding returns Amount - PaidAmount.
func (ar AccountsReceivable) Outstanding() decimal.Decimal {
	return ar.Amount.Sub(ar.PaidAmount)
}

// Payment is money received from a customer, applied against one or more
// AccountsReceivable entries via Allocations.
type Payment struct {
	gorm.Model
	CustomerID     uint                `gorm:"index;not null" json:"CustomerID"`
	Customer       User                `json:"Customer"`
	Amount         decimal.Decimal     `gorm:"type:decimal(19,4);not null" json:"Amount"`
	Method         string              `json:"Method"` // cash, card, bkash, bank
	ReceivedAt     time.Time           `json:"ReceivedAt"`
	ReceivedBy     uint                `json:"ReceivedBy"`
	Notes          string              `json:"Notes"`
	JournalEntryID uint                `json:"JournalEntryID"`
	Allocations    []PaymentAllocation `gorm:"foreignKey:PaymentID" json:"Allocations"`
}

// PaymentAllocation applies part of a Payment to a specific AccountsReceivable.
// A single payment can settle multiple AR entries (and vice versa).
type PaymentAllocation struct {
	gorm.Model
	PaymentID            uint               `gorm:"index;not null" json:"PaymentID"`
	AccountsReceivableID uint               `gorm:"index;not null" json:"AccountsReceivableID"`
	AccountsReceivable   AccountsReceivable `json:"AccountsReceivable"`
	Amount               decimal.Decimal    `gorm:"type:decimal(19,4);not null" json:"Amount"`
}

// AccountsPayable tracks money we owe a supplier from received purchase orders.
// Created when stock is received (DR Inventory / CR Accounts Payable), settled
// via SupplierPayments.
type AccountsPayable struct {
	gorm.Model
	SupplierID      uint            `gorm:"index;not null" json:"SupplierID"`
	Supplier        Supplier        `json:"Supplier"`
	PurchaseOrderID uint            `gorm:"index" json:"PurchaseOrderID"`
	PurchaseOrder   PurchaseOrder   `json:"PurchaseOrder"`
	Amount          decimal.Decimal `gorm:"type:decimal(19,4);not null" json:"Amount"` // total owed for this receipt
	PaidAmount      decimal.Decimal `gorm:"type:decimal(19,4);default:0" json:"PaidAmount"`
	Status          string          `gorm:"index;default:'UNPAID'" json:"Status"` // UNPAID, PARTIAL, PAID
	DueDate         *time.Time      `json:"DueDate"`
	JournalEntryID  uint            `json:"JournalEntryID"`
}

// Outstanding returns Amount - PaidAmount.
func (ap AccountsPayable) Outstanding() decimal.Decimal {
	return ap.Amount.Sub(ap.PaidAmount)
}

// SupplierPayment is money paid to a supplier, applied against one or more
// AccountsPayable entries via Allocations.
type SupplierPayment struct {
	gorm.Model
	SupplierID     uint                        `gorm:"index;not null" json:"SupplierID"`
	Supplier       Supplier                    `json:"Supplier"`
	Amount         decimal.Decimal             `gorm:"type:decimal(19,4);not null" json:"Amount"`
	Method         string                      `json:"Method"` // cash, card, bkash, bank
	PaidAt         time.Time                   `json:"PaidAt"`
	PaidBy         uint                        `json:"PaidBy"`
	Notes          string                      `json:"Notes"`
	JournalEntryID uint                        `json:"JournalEntryID"`
	Allocations    []SupplierPaymentAllocation `gorm:"foreignKey:SupplierPaymentID" json:"Allocations"`
}

// SupplierPaymentAllocation applies part of a SupplierPayment to an AccountsPayable.
type SupplierPaymentAllocation struct {
	gorm.Model
	SupplierPaymentID uint            `gorm:"index;not null" json:"SupplierPaymentID"`
	AccountsPayableID uint            `gorm:"index;not null" json:"AccountsPayableID"`
	AccountsPayable   AccountsPayable `json:"AccountsPayable"`
	Amount            decimal.Decimal `gorm:"type:decimal(19,4);not null" json:"Amount"`
}
