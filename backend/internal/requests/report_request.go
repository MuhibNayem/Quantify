package requests

import "time"

// SalesTrendsReportRequest represents parameters for sales trends report.
type SalesTrendsReportRequest struct {
	StartDate  time.Time `json:"startDate" binding:"required"`
	EndDate    time.Time `json:"endDate" binding:"required"`
	CategoryID *uint     `json:"categoryId"`
	LocationID *uint     `json:"locationId"`
	GroupBy    string    `json:"groupBy" binding:"required,oneof=daily weekly monthly"`
}

// InventoryTurnoverReportRequest represents parameters for inventory turnover report.
type InventoryTurnoverReportRequest struct {
	StartDate  time.Time `json:"startDate" binding:"required"`
	EndDate    time.Time `json:"endDate" binding:"required"`
	CategoryID *uint     `json:"categoryId"`
	LocationID *uint     `json:"locationId"`
}

// ProfitMarginReportRequest represents parameters for profit margin report.
type ProfitMarginReportRequest struct {
	StartDate  time.Time `json:"startDate" binding:"required"`
	EndDate    time.Time `json:"endDate" binding:"required"`
	CategoryID *uint     `json:"categoryId"`
	LocationID *uint     `json:"locationId"`
}
