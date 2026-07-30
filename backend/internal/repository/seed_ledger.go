package repository

import (
	"inventory/backend/internal/domain"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// chartOfAccounts is the default chart of accounts. Codes follow a simple
// convention: 1xxx assets, 2xxx liabilities, 3xxx equity, 4xxx revenue, 5xxx expenses.
var chartOfAccounts = []domain.Account{
	{Code: domain.AccountCodeCash, Name: "Cash", Type: domain.AccountAsset, Description: "Cash on hand and bank", System: true, IsActive: true},
	{Code: domain.AccountCodeAccountsReceivable, Name: "Accounts Receivable", Type: domain.AccountAsset, Description: "Money owed by customers (credit sales)", System: true, IsActive: true},
	{Code: domain.AccountCodeInventory, Name: "Inventory", Type: domain.AccountAsset, Description: "Value of stock on hand", System: true, IsActive: true},
	{Code: domain.AccountCodeAccountsPayable, Name: "Accounts Payable", Type: domain.AccountLiability, Description: "Money owed to suppliers", System: true, IsActive: true},
	{Code: "3000", Name: "Owner's Equity", Type: domain.AccountEquity, Description: "Owner's stake in the business", System: true, IsActive: true},
	{Code: domain.AccountCodeSalesRevenue, Name: "Sales Revenue", Type: domain.AccountRevenue, Description: "Revenue from sales", System: true, IsActive: true},
	{Code: domain.AccountCodeSalesDiscounts, Name: "Sales Discounts", Type: domain.AccountRevenue, Description: "Manual discounts given (contra-revenue)", System: true, IsActive: true},
	{Code: domain.AccountCodeCOGS, Name: "Cost of Goods Sold", Type: domain.AccountExpense, Description: "Direct cost of sold inventory", System: true, IsActive: true},
}

// SeedChartOfAccounts idempotently inserts the default chart of accounts.
func SeedChartOfAccounts(db *gorm.DB) error {
	for _, acc := range chartOfAccounts {
		a := acc
		if err := db.Where("code = ?", a.Code).
			Assign(domain.Account{Name: a.Name, Type: a.Type, Description: a.Description, System: a.System, IsActive: true}).
			FirstOrCreate(&a).Error; err != nil {
			return err
		}
	}
	logrus.WithField("accounts", len(chartOfAccounts)).Info("Seeded chart of accounts")
	return nil
}
