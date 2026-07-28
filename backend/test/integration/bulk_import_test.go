//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/services"
)

func TestBulkImportEndToEnd(t *testing.T) {
	db := SetupPostgres(t)
	Migrate(t, db)

	loc := domain.Location{Name: "Warehouse A"}
	require.NoError(t, db.Create(&loc).Error)

	cat := domain.Category{Name: "Groceries"}
	require.NoError(t, db.Create(&cat).Error)

	sup := domain.Supplier{Name: "FreshFarms"}
	require.NoError(t, db.Create(&sup).Error)

	categoryRepo := repository.NewCategoryRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	locationRepo := repository.NewLocationRepository(db)

	svc := services.NewBulkImportService(categoryRepo, supplierRepo, locationRepo)

	csv := `SKU,Name,Description,Category,SubCategory,Supplier,Brand,PurchasePrice,SellingPrice,Location,Status
SKU-100,Apple,Fresh apple,Groceries,Fruits,FreshFarms,NatureFarm,0.50,1.20,Warehouse A,Active
SKU-101,Banana,Ripe banana,Groceries,Fruits,FreshFarms,NatureFarm,0.30,0.80,Warehouse A,Active
SKU-102,,Missing name,Groceries,,FreshFarms,,0.10,0.20,Warehouse A,Active
`

	result, err := svc.ProcessBulkImport(strings.NewReader(csv))
	require.NoError(t, err)

	assert.Equal(t, 3, result.TotalRecords)
	assert.Equal(t, 2, result.ValidRecords)
	assert.Equal(t, 1, result.InvalidRecords)
	assert.Len(t, result.ValidProducts, 2)
	assert.Equal(t, "SKU-100", result.ValidProducts[0].SKU)
	assert.Equal(t, "SKU-101", result.ValidProducts[1].SKU)

	assert.True(t, result.NewEntities.SubCategories["Fruits"])
}
