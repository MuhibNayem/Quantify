package consumers

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/services"

	"github.com/alicebob/miniredis/v2"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func zeroTime() time.Time {
	return time.Time{}
}

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

// TestBulkReconfirm_IsIdempotent verifies that re-confirming an already-completed
// bulk import job is a no-op: no duplicate Product rows are created, no error is
// surfaced, and the Job.Status remains COMPLETED.
//
// Regression target: P3.T2 audit found that re-confirming a job re-runs
// processImportConfirmDelivery and re-inserts already-created products, causing
// unique-constraint errors on SKU.
func TestBulkReconfirm_IsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:bulk-reconfirm?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Job{},
		&domain.Product{},
		&domain.Category{},
		&domain.SubCategory{},
		&domain.Supplier{},
		&domain.Location{},
	))

	// finalizeImport invalidates Redis cache; stand up an in-process miniredis
	// so the global repository.RedisClient is non-nil.
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	prevRedis := repository.RedisClient
	repository.RedisClient = &repository.UniversalClient{UniversalClient: rdb}
	t.Cleanup(func() { repository.RedisClient = prevRedis })

	// Pre-create entities so the products reference real IDs.
	category := domain.Category{Name: "Category"}
	require.NoError(t, db.Create(&category).Error)
	supplier := domain.Supplier{Name: "Supplier"}
	require.NoError(t, db.Create(&supplier).Error)
	location := domain.Location{Name: "Warehouse"}
	require.NoError(t, db.Create(&location).Error)

	// Build a valid BulkImportResult and serialize it onto the Job.
	result := services.BulkImportResult{
		TotalRecords:  2,
		ValidRecords:  2,
		ValidProducts: []domain.Product{
			{
				SKU:           "SKU-1",
				Name:          "Product 1",
				Description:   "Desc 1",
				CategoryID:    category.ID,
				SupplierID:    supplier.ID,
				LocationID:    location.ID,
				Brand:         "Brand",
				PurchasePrice: 10,
				SellingPrice:  20,
				Status:        "Active",
				Category:      domain.Category{Name: category.Name},
				Supplier:      domain.Supplier{Name: supplier.Name},
				Location:      domain.Location{Name: location.Name},
			},
			{
				SKU:           "SKU-2",
				Name:          "Product 2",
				Description:   "Desc 2",
				CategoryID:    category.ID,
				SupplierID:    supplier.ID,
				LocationID:    location.ID,
				Brand:         "Brand",
				PurchasePrice: 15,
				SellingPrice:  25,
				Status:        "Active",
				Category:      domain.Category{Name: category.Name},
				Supplier:      domain.Supplier{Name: supplier.Name},
				Location:      domain.Location{Name: location.Name},
			},
		},
	}
	resultBytes, err := json.Marshal(result)
	require.NoError(t, err)

	job := domain.Job{
		Type:   "BULK_IMPORT",
		Status: "PENDING_CONFIRMATION",
		Result: string(resultBytes),
	}
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

	confirmDelivery := func() (retErr error) {
		defer func() {
			if r := recover(); r != nil {
				retErr = errors.New("confirm panicked")
			}
		}()
		body, mErr := json.Marshal(map[string]interface{}{"jobId": float64(job.ID)})
		require.NoError(t, mErr)
		consumer.processImportConfirmDelivery(amqp091.Delivery{Body: body})
		return nil
	}

	// First confirm: should succeed and create products.
	require.NoError(t, confirmDelivery())

	var prodCount int64
	require.NoError(t, db.Model(&domain.Product{}).Count(&prodCount).Error)
	assert.Equal(t, int64(len(result.ValidProducts)), prodCount, "first confirm must insert products")

	var persisted domain.Job
	require.NoError(t, db.First(&persisted, job.ID).Error)
	require.Equal(t, "COMPLETED", persisted.Status, "first confirm must complete the job")

	// Second confirm: must be a no-op (idempotent).
	require.NoError(t, confirmDelivery())

	require.NoError(t, db.Model(&domain.Product{}).Count(&prodCount).Error)
	assert.Equal(t, int64(len(result.ValidProducts)), prodCount, "no duplicates after re-confirm")

	var persistedAgain domain.Job
	require.NoError(t, db.First(&persistedAgain, job.ID).Error)
	assert.Equal(t, "COMPLETED", persistedAgain.Status, "re-confirm must not change a completed job's status")
}

// TestBulkFinalizeImport_IsIdempotent verifies that finalizeImport can be invoked
// repeatedly against the same already-completed job without raising a
// unique-constraint error on Product.SKU.
//
// Regression target: P3.T2 audit found that finalizeImport unconditionally
// re-inserts already-created products. ON CONFLICT DO NOTHING at the bulk-insert
// site is the defense-in-depth guarantee.
func TestBulkFinalizeImport_IsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:bulk-finalize?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Job{},
		&domain.Product{},
		&domain.Category{},
		&domain.SubCategory{},
		&domain.Supplier{},
		&domain.Location{},
	))

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	prevRedis := repository.RedisClient
	repository.RedisClient = &repository.UniversalClient{UniversalClient: rdb}
	t.Cleanup(func() { repository.RedisClient = prevRedis })

	category := domain.Category{Name: "Category"}
	require.NoError(t, db.Create(&category).Error)
	supplier := domain.Supplier{Name: "Supplier"}
	require.NoError(t, db.Create(&supplier).Error)
	location := domain.Location{Name: "Warehouse"}
	require.NoError(t, db.Create(&location).Error)

	result := services.BulkImportResult{
		TotalRecords:  1,
		ValidRecords:  1,
		ValidProducts: []domain.Product{
			{
				SKU:           "SKU-1",
				Name:          "Product 1",
				CategoryID:    category.ID,
				SupplierID:    supplier.ID,
				LocationID:    location.ID,
				PurchasePrice: 10,
				SellingPrice:  20,
				Status:        "Active",
				Category:      domain.Category{Name: category.Name},
				Supplier:      domain.Supplier{Name: supplier.Name},
				Location:      domain.Location{Name: location.Name},
			},
		},
	}

	job := domain.Job{Type: "BULK_IMPORT", Status: "PROCESSING"}
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

	// First finalize: inserts products.
	require.NoError(t, consumer.finalizeImport(&job, &result))

	var prodCount int64
	require.NoError(t, db.Model(&domain.Product{}).Count(&prodCount).Error)
	assert.Equal(t, int64(1), prodCount, "first finalize must insert the product")

	// Reset IDs on the result so the second call simulates a fresh validation
	// payload (same SKU, new primary key), forcing the SKU unique index to be
	// the constraint the ON CONFLICT clause must catch.
	for i := range result.ValidProducts {
		result.ValidProducts[i].ID = 0
		result.ValidProducts[i].CreatedAt = zeroTime()
		result.ValidProducts[i].UpdatedAt = zeroTime()
	}

	// Second finalize on the same SKUs must not raise a unique-constraint error.
	require.NoError(t, consumer.finalizeImport(&job, &result), "finalizeImport must be idempotent across re-runs")

	require.NoError(t, db.Model(&domain.Product{}).Count(&prodCount).Error)
	assert.Equal(t, int64(1), prodCount, "second finalize must not duplicate the product")
}