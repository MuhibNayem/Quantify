package consumers

import (
	"bytes"
	"errors"
	"testing"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/services"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBulkConsumer_DoesNotAutoFinalize(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:bulk-consumer?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Job{},
		&domain.Product{},
		&domain.Category{},
		&domain.SubCategory{},
		&domain.Supplier{},
		&domain.Location{},
	))

	job := domain.Job{Type: "BULK_IMPORT", Status: "QUEUED"}
	require.NoError(t, db.Create(&job).Error)

	categoryRepo := repository.NewCategoryRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	locationRepo := repository.NewLocationRepository(db)
	consumer := NewBulkConsumer(
		db,
		repository.NewJobRepository(db),
		nil,
		categoryRepo,
		supplierRepo,
		locationRepo,
		nil,
		services.NewBulkImportService(categoryRepo, supplierRepo, locationRepo),
		nil,
		nil,
		nil,
	)

	csv := "sku,name,description,category,subcategory,supplier,brand,purchasePrice,sellingPrice,location,status\nSKU-1,Product 1,Description,Category,Subcategory,Supplier,Brand,10,20,Warehouse,Active\n"

	validateErr := func() (retErr error) {
		defer func() {
			if r := recover(); r != nil {
				retErr = errors.New("consumer attempted to finalize during validation")
			}
		}()
		return consumer.processValidationPass(&job, bytes.NewReader([]byte(csv)), amqp091.Delivery{})
	}()
	require.NoError(t, validateErr, "consumer must not run finalizeImport during validation")

	var persisted domain.Job
	require.NoError(t, db.First(&persisted, job.ID).Error)
	assert.Equal(t, "PENDING_CONFIRMATION", persisted.Status, "consumer must not auto-finalize")
	var productCount int64
	require.NoError(t, db.Model(&domain.Product{}).Count(&productCount).Error)
	assert.Zero(t, productCount, "consumer must not insert products before confirmation")
}