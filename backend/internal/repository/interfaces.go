package repository

import (
	"time"

	"inventory/backend/internal/domain"
)

// UserRepository defines methods for interacting with users and their roles.
type UserRepository interface {
	CreateUser(user *domain.User) error
	UpdateUser(user *domain.User) error
	DeleteUser(user *domain.User) error
	GetAllUsers() ([]domain.User, error)
	ListUsers(page, limit int, status, search string) ([]domain.User, int64, error)
	GetUserByID(id string) (*domain.User, error)
	GetUserByUsername(username string) (*domain.User, error)
	CountUsers() (int64, error)
	GetRoleByName(name string) (*domain.Role, error)
	GetRoleByID(id uint) (*domain.Role, error)
}

// ProductRepository defines methods for interacting with products.
type ProductRepository interface {
	CreateProduct(product *domain.Product) error
	UpdateProduct(product *domain.Product, updates map[string]interface{}) error
	DeleteProduct(product *domain.Product) error
	GetProductBySKU(sku string) (*domain.Product, error)
	GetProductByBarcode(barcode string) (*domain.Product, error)
	CreateBulk(products *[]domain.Product) error
	GetAll() ([]domain.Product, error)
	GetFiltered(filters map[string]interface{}) ([]domain.Product, error)
}

// CategoryRepository defines methods for interacting with categories and sub-categories.
type CategoryRepository interface {
	CreateCategory(category *domain.Category) error
	UpdateCategory(category *domain.Category, updates map[string]interface{}) error
	DeleteCategory(category *domain.Category) error
	CreateSubCategory(subCategory *domain.SubCategory) error
	GetAll() ([]domain.Category, error)
	GetAllSubCategories() ([]domain.SubCategory, error)
	GetByNames(names []string, categories *[]domain.Category) error
	GetSubCategoriesByNames(names []string, subCategories *[]domain.SubCategory) error
	GetCategoryByName(name string) (*domain.Category, error)
}

// SupplierRepository defines methods for interacting with suppliers.
type SupplierRepository interface {
	CreateSupplier(supplier *domain.Supplier) error
	UpdateSupplier(supplier *domain.Supplier, updates map[string]interface{}) error
	DeleteSupplier(supplier *domain.Supplier) error
	GetAll() ([]domain.Supplier, error)
	GetByNames(names []string, suppliers *[]domain.Supplier) error
	GetSupplierByName(name string) (*domain.Supplier, error)
	GetSupplierPerformance(supplierID uint) (float64, float64, error)
}

// LocationRepository defines methods for interacting with locations.
type LocationRepository interface {
	GetAll() ([]domain.Location, error)
	GetByNames(names []string, locations *[]domain.Location) error
}

// JobRepository defines methods for interacting with background jobs.
type JobRepository interface {
	CreateJob(job *domain.Job) error
	GetJob(jobID uint) (*domain.Job, error)
	UpdateJob(job *domain.Job) error
	ListJobs(limit int) ([]domain.Job, error)
}

// ReportsRepository defines methods for generating analytics and reports.
type ReportsRepository interface {
	GetSalesTrends(startDate, endDate time.Time, categoryID, locationID, productID *uint, groupBy, timezone string) ([]SalesTrend, []TopSellingProduct, error)
	GetInventoryTurnover(startDate, endDate time.Time, categoryID, locationID *uint) (float64, float64, error)
	GetProfitMargin(startDate, endDate time.Time, categoryID, locationID *uint) (float64, float64, error)
	GetDailySalesSummary(date time.Time) (float64, error)
	GetStockAgingReport() (map[string][]StockAgingItem, error)
	GetDeadStockReport(daysThreshold int) ([]DeadStockItem, error)
	GetSupplierPerformanceReport(startDate, endDate time.Time) ([]SupplierScorecard, error)
	GetHourlySalesHeatmap(startDate, endDate time.Time, timezone string) ([]HeatmapPoint, error)
	GetSalesByEmployeeReport(startDate, endDate time.Time) ([]EmployeeSalesStats, error)
	GetCategoryDrillDownReport(startDate, endDate time.Time) ([]CategoryPerformance, error)
	GetCOGSAndGMROIReport(startDate, endDate time.Time) (GMROIStats, error)
	GetVoidDiscountAuditReport(startDate, endDate time.Time) ([]domain.AuditLog, error)
	GetTaxLiabilityReport(startDate, endDate time.Time) ([]TaxLiabilityStats, error)
	GetCashDrawerReconciliationReport(startDate, endDate time.Time) ([]domain.CashDrawerSession, error)
	GetBasketAnalysisReport(startDate, endDate time.Time, page, limit int) ([]BasketAnalysisItem, int, error)
	GetPOAnalysisReport(startDate, endDate time.Time) ([]domain.PurchaseOrder, error)
	GetCustomerInsightsReport(startDate, endDate time.Time) ([]CustomerInsight, error)
	GetProductPerformanceAnalytics(startDate, endDate time.Time, supplierName string, minStock int) ([]ProductPerformanceAnalytics, error)
	GetShrinkageReport(startDate, endDate time.Time) ([]ShrinkageItem, error)
	GetCustomerReturnAnalysisReport(startDate, endDate time.Time) ([]ReturnAnalysisItem, error)
}

// SearchRepository defines methods for performing global search.
type SearchRepository interface {
	Search(query string) ([]SearchResult, error)
}

// DashboardRepository defines methods for retrieving dashboard data.
type DashboardRepository interface {
	GetStats() (*DashboardStats, error)
	GetRecentProducts(limit int) ([]domain.Product, error)
	GetRecentAlerts(limit int) ([]domain.Alert, error)
	GetRecentReorderSuggestions(limit int) ([]domain.ReorderSuggestion, error)
	GetSalesChartData() ([]float64, error)
	GetSalesTrend() (*SalesTrendData, error)
}
