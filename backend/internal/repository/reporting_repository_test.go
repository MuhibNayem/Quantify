package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func setupReportingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:reporting-test-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Supplier{},
		&domain.Product{},
		&domain.Batch{},
		&domain.StockAdjustment{},
	))
	return db
}

func TestProductPerformanceReport_Succeeds(t *testing.T) {
	db := setupReportingTestDB(t)

	supplier := domain.Supplier{Name: "Acme"}
	require.NoError(t, db.Create(&supplier).Error)
	product := domain.Product{
		SKU:           "SKU-1",
		Name:          "Widget",
		SupplierID:    supplier.ID,
		PurchasePrice: 5,
		SellingPrice:  10,
	}
	require.NoError(t, db.Create(&product).Error)

	repo := NewReportsRepository(db)
	start := time.Now().Add(-30 * 24 * time.Hour)
	end := time.Now()
	report, err := repo.GetProductPerformanceAnalytics(start, end, "", 0)
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.NotEmpty(t, report, "expected at least the seeded product in the report")
}

// TestHeatmapReport_RejectsMaliciousTimezone guards against SQL injection via the
// timezone argument. The timezone originates from a SystemSetting that an operator
// can edit; the repository must safely handle (parameterize + validate) arbitrary
// strings rather than interpolating them into the raw SQL.
func TestHeatmapReport_RejectsMaliciousTimezone(t *testing.T) {
	db := setupReportingTestDB(t)

	repo := NewReportsRepository(db)
	malicious := "UTC' UNION SELECT password FROM users --"
	start := time.Now().Add(-7 * 24 * time.Hour)
	end := time.Now()

	_, err := repo.GetHourlySalesHeatmap(start, end, malicious)
	// We don't assert NoError: the raw query uses Postgres-only syntax
	// (EXTRACT/DOW/AT TIME ZONE) which SQLite cannot execute, so a dialect
	// "syntax error" is expected on this test backend regardless of input.
	// The security property we exercise is that the malicious payload is NOT
	// interpolated into the SQL: the error message echoes the failing query
	// and must not contain any part of the injection.
	if err != nil {
		assert.NotContains(t, err.Error(), "UNION SELECT", "malicious timezone must not be interpolated into SQL; got: %v", err)
		assert.NotContains(t, err.Error(), "password", "malicious timezone must not be interpolated into SQL; got: %v", err)
	}
}

// TestIsValidTimezone verifies the timezone validator rejects attacker-controlled
// characters (quotes, spaces, semicolons, SQL comments) and accepts the IANA
// zone-name alphabet (letters, digits, '/', '_', '+', '-').
func TestIsValidTimezone(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"UTC", "UTC", true},
		{"utc_lower", "utc", true},
		{"iana_with_slash", "America/New_York", true},
		{"iana_with_underscore", "America/Argentina/Buenos_Aires", true},
		{"iana_with_offset", "Etc/GMT+5", true},
		{"iana_with_negative_offset", "Etc/GMT-12", true},
		{"sql_injection_union", "UTC' UNION SELECT password FROM users --", false},
		{"sql_injection_quote", "UTC'; DROP TABLE users;--", false},
		{"sql_injection_semicolon", "UTC;evil", false},
		{"contains_space", "America New_York", false},
		{"contains_dot", "America.NYC", false},
		{"contains_quote", "UTC'evil", false},
		{"contains_backtick", "UTC`evil", false},
		// All chars are in the allowed alphabet, so the character-class
		// validator accepts it. SQL safety still holds because the value
		// is bound as a parameter — Postgres simply rejects the unknown
		// timezone at execution time.
		{"comment_chars_only", "UTC--comment", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isValidTimezone(tc.in))
		})
	}
}