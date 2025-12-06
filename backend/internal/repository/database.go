package repository

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"gorm.io/driver/postgres"

	"gorm.io/gorm"

	"inventory/backend/internal/config"

	"inventory/backend/internal/domain"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",

		cfg.DBHost,

		cfg.DBUser,

		cfg.DBPassword,

		cfg.DBName,

		cfg.DBPort,
	)

	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {

		logrus.Fatalf("Failed to connect to database: %v", err)

	}

	logrus.Info("Database connection established")

	// Temporarily Auto-migrate the schema to apply the unique index

	AutoMigrate()
	ensureBarcodeUniqueIndex()

	ensureDefaultAlertSubscriptions()

}

func AutoMigrate() {

	err := DB.AutoMigrate(

		&domain.User{},

		&domain.Product{},

		&domain.Category{},

		&domain.SubCategory{},

		&domain.Supplier{},

		&domain.Location{},

		&domain.StockAdjustment{},

		&domain.Batch{},

		&domain.PurchaseOrder{},

		&domain.PurchaseOrderItem{},

		&domain.ReorderSuggestion{},

		&domain.DemandForecast{},

		&domain.Alert{},

		&domain.UserNotificationSettings{},
		&domain.ProductAlertSettings{},

		&domain.StockTransfer{},

		&domain.Transaction{},

		&domain.LoyaltyAccount{},

		&domain.TimeClock{},

		&domain.Job{},

		&domain.Notification{},

		&domain.AlertRoleSubscription{},

		&domain.GlobalSearchEntry{},
		&domain.SystemSetting{},
		&domain.Permission{},
		&domain.RolePermission{},
	)

	if err != nil {

		logrus.Fatalf("Database schema auto-migration failed: %v", err)

	}

	logrus.Info("Database schema auto-migrated")

	// Fix schema first (drop legacy columns) before seeding
	fixRolePermissionsSchema()
	seedData()
}

func fixRolePermissionsSchema() {
	if DB == nil {
		return
	}
	// The role_permissions table might have a legacy 'role' column from previous migrations.
	// This column has a NOT NULL constraint which breaks inserts since we now use foreign keys.
	// We safely drop it if it exists.
	query := `ALTER TABLE role_permissions DROP COLUMN IF EXISTS role;`
	if err := DB.Exec(query).Error; err != nil {
		logrus.Warnf("Failed to drop legacy 'role' column from role_permissions (this is expected if table doesn't exist yet): %v", err)
	}
}

func seedData() {
	// Seed Permissions
	permissions := []domain.Permission{
		// Inventory - Products
		{Name: "products.read", Group: "Product Management", Description: "View products"},
		{Name: "products.write", Group: "Product Management", Description: "Create/Edit products"},
		{Name: "products.delete", Group: "Product Management", Description: "Delete products"},
		// Inventory - Categories
		{Name: "categories.read", Group: "Product Management", Description: "View categories"},
		{Name: "categories.write", Group: "Product Management", Description: "Manage categories"},
		// Inventory - Locations
		{Name: "locations.read", Group: "Inventory", Description: "View locations"},
		{Name: "locations.write", Group: "Inventory", Description: "Manage locations"},
		// Suppliers
		{Name: "suppliers.read", Group: "Inventory", Description: "View suppliers"},
		{Name: "suppliers.write", Group: "Inventory", Description: "Manage suppliers"},
		// Barcode
		{Name: "barcode.read", Group: "Inventory", Description: "Lookup barcodes"},
		// Replenishment
		{Name: "replenishment.read", Group: "Inventory", Description: "View forecasts/suggestions"},
		{Name: "replenishment.write", Group: "Inventory", Description: "Generate forecasts and manage POs"},
		// CRM
		{Name: "customers.read", Group: "CRM", Description: "View customers"},
		{Name: "customers.write", Group: "CRM", Description: "Manage customers"},
		{Name: "loyalty.read", Group: "CRM", Description: "View loyalty info"},
		{Name: "loyalty.write", Group: "CRM", Description: "Manage loyalty points"},
		// POS / Sales
		{Name: "pos.access", Group: "POS", Description: "Access Point of Sale terminal"},
		// Reports
		{Name: "reports.sales", Group: "Reports", Description: "View sales reports"},
		{Name: "reports.inventory", Group: "Reports", Description: "View inventory reports"},
		{Name: "reports.financial", Group: "Reports", Description: "View financial reports"},
		// Bulk Operations
		{Name: "bulk.import", Group: "System", Description: "Import data in bulk"},
		{Name: "bulk.export", Group: "System", Description: "Export data in bulk"},
		// Settings & Access
		{Name: "settings.view", Group: "Settings", Description: "View system settings"},
		{Name: "settings.manage", Group: "Settings", Description: "Edit system settings"},
		{Name: "users.view", Group: "Access Control", Description: "View users"},
		{Name: "users.manage", Group: "Access Control", Description: "Manage users"},
		{Name: "roles.view", Group: "Access Control", Description: "View roles"},
		{Name: "roles.manage", Group: "Access Control", Description: "Manage roles"},
		// Alerts
		{Name: "alerts.view", Group: "Inventory", Description: "View alerts"},
		{Name: "alerts.manage", Group: "Inventory", Description: "Resolve alerts"},
	}

	permMap := make(map[string]domain.Permission)
	for _, p := range permissions {
		DB.FirstOrCreate(&p, domain.Permission{Name: p.Name})
		permMap[p.Name] = p
	}

	// Seed Roles
	roles := []domain.Role{
		{Name: "Admin", Description: "Full system access", IsSystem: true},
		{Name: "Manager", Description: "Store management access", IsSystem: true},
		{Name: "Staff", Description: "Standard employee access", IsSystem: true},
		{Name: "Customer", Description: "Customer portal access", IsSystem: true},
	}

	for _, r := range roles {
		var role domain.Role
		if err := DB.FirstOrCreate(&role, domain.Role{Name: r.Name}).Error; err != nil {
			logrus.Errorf("Failed to seed role %s: %v", r.Name, err)
			continue
		}

		// Assign Default Permissions
		var permsToAssign []domain.Permission
		switch r.Name {
		case "Admin":
			// Admin gets ALL permissions
			for _, p := range permMap {
				permsToAssign = append(permsToAssign, p)
			}
		case "Manager":
			// Manager gets almost everything except maybe system-level destructive role management?
			// Giving Manager broad access for now as per previous logic
			permsToAssign = append(permsToAssign,
				permMap["products.read"], permMap["products.write"], permMap["products.delete"],
				permMap["categories.read"], permMap["categories.write"],
				permMap["locations.read"], permMap["locations.write"],
				permMap["suppliers.read"], permMap["suppliers.write"],
				permMap["barcode.read"],
				permMap["replenishment.read"], permMap["replenishment.write"],
				permMap["customers.read"], permMap["customers.write"],
				permMap["loyalty.read"], permMap["loyalty.write"],
				permMap["pos.access"],
				permMap["reports.sales"], permMap["reports.inventory"], permMap["reports.financial"],
				permMap["alerts.view"], permMap["alerts.manage"],
				permMap["bulk.import"], permMap["bulk.export"],
				permMap["users.view"], permMap["users.manage"],
				permMap["settings.view"],
			)
		case "Staff":
			permsToAssign = append(permsToAssign,
				permMap["products.read"],
				permMap["categories.read"],
				permMap["locations.read"],
				permMap["suppliers.read"],
				permMap["barcode.read"],
				permMap["customers.read"],
				permMap["pos.access"],
				permMap["alerts.view"],
			)
		}

		if len(permsToAssign) > 0 {
			if err := DB.Model(&role).Association("Permissions").Replace(permsToAssign); err != nil {
				logrus.Errorf("Failed to assign permissions to role %s: %v", r.Name, err)
			}
		}
	}

	// Seed Settings
	settings := []domain.SystemSetting{
		{Key: "business_name", Value: "Quantify Business", Group: "General", Type: "string"},
		{Key: "currency_symbol", Value: "$", Group: "General", Type: "string"},
		{Key: "timezone", Value: "UTC", Group: "General", Type: "string"},
	}

	for _, s := range settings {
		DB.FirstOrCreate(&s, domain.SystemSetting{Key: s.Key})
	}

	logrus.Info("System Data (Roles, Permissions, Settings) seeded")
}

func ensureDefaultAlertSubscriptions() {

	defaultTypes := []string{"LOW_STOCK", "OUT_OF_STOCK", "OVERSTOCK", "EXPIRY_ALERT"}

	for _, alertType := range defaultTypes {

		var count int64

		if err := DB.Model(&domain.AlertRoleSubscription{}).
			Where("alert_type = ? AND role = ?", alertType, "Admin").
			Count(&count).Error; err != nil {

			logrus.Errorf("Failed to check default alert subscription for %s: %v", alertType, err)

			continue

		}

		if count == 0 {

			subscription := domain.AlertRoleSubscription{

				AlertType: alertType,

				Role: "Admin",
			}

			if err := DB.Create(&subscription).Error; err != nil {

				logrus.Errorf("Failed to seed default alert subscription for %s: %v", alertType, err)

			} else {

				logrus.Infof("Seeded default alert subscription %s -> Admin", alertType)

			}

		}

	}

}

func ensureBarcodeUniqueIndex() {
	if DB == nil {
		return
	}

	dropConstraint := `
DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'products_barcode_upc_key'
	) THEN
		ALTER TABLE products DROP CONSTRAINT products_barcode_upc_key;
	END IF;
END $$;
`
	if err := DB.Exec(dropConstraint).Error; err != nil {
		logrus.Errorf("Failed to drop legacy barcode constraint: %v", err)
	}

	createPartialIndex := `
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_barcode_upc_not_null
ON products (barcode_upc)
WHERE barcode_upc IS NOT NULL AND barcode_upc <> '';
`
	if err := DB.Exec(createPartialIndex).Error; err != nil {
		logrus.Errorf("Failed to ensure partial barcode index: %v", err)
	}
}

// CloseDB closes the database connection.

func CloseDB() {

	sqlDB, err := DB.DB()

	if err != nil {

		logrus.Errorf("Error getting underlying DB: %v", err)

		return

	}

	if err := sqlDB.Close(); err != nil {

		logrus.Errorf("Error closing database connection: %v", err)

	}

	logrus.Info("Database connection closed")

}
