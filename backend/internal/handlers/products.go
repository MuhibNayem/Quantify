package handlers

import (
	"encoding/json"
	"fmt"
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ProductHandler holds the repository dependencies for product-related handlers.
type ProductHandler struct {
	productRepo *repository.ProductRepository
	db          *gorm.DB // Keep db for now for existing functions
}

// NewProductHandler creates a new ProductHandler with the given repository.
func NewProductHandler(productRepo *repository.ProductRepository, db *gorm.DB) *ProductHandler {
	return &ProductHandler{
		productRepo: productRepo,
		db:          db,
	}
}

// ProductEventPayload defines the payload for product-related events.
type ProductEventPayload struct {
	ProductID uint   `json:"productId"`
	SKU       string `json:"sku"`
	Name      string `json:"name"`
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with the provided details
// @Tags products
// @Accept json
// @Produce json
// @Param product body requests.ProductCreateRequest true "Product creation request"
// @Success 201 {object} domain.Product
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req requests.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	product := domain.Product{
		SKU:           req.SKU,
		Name:          req.Name,
		Description:   req.Description,
		CategoryID:    req.CategoryID,
		SubCategoryID: req.SubCategoryID,
		SupplierID:    req.SupplierID,
		Brand:         req.Brand,
		PurchasePrice: req.PurchasePrice,
		SellingPrice:  req.SellingPrice,
		BarcodeUPC:    req.BarcodeUPC,
		ImageURLs:     req.ImageURLs,
		Status:        "Active", // Default status
		LocationID:    req.LocationID,
	}

	if err := h.db.Create(&product).Error; err != nil {
		// Check for unique constraint violation (e.g., SKU, BarcodeUPC)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			c.Error(appErrors.NewAppError("Product with this SKU or BarcodeUPC already exists", http.StatusConflict, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to create product", http.StatusInternalServerError, err))
		return
	}

	// Invalidate relevant caches
	repository.DeleteCache(fmt.Sprintf("product:%d", product.ID))
	repository.DeleteCache("products:*") // Invalidate all product list caches

	// Publish ProductCreatedEvent
	payload := ProductEventPayload{
		ProductID: product.ID,
		SKU:       product.SKU,
		Name:      product.Name,
	}
	if err := message_broker.Publish(c.Request.Context(), "inventory", "product.created", payload); err != nil {
		logrus.Errorf("Failed to publish product created event: %v", err)
	}

	c.JSON(http.StatusCreated, product)
}

// ListProducts godoc
// @Summary Get a list of products
// @Description Get a paginated, searchable, and filterable list of products
// @Tags products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param search query string false "Search term for Product Name, SKU, or Barcode"
// @Param category query int false "Filter by Category ID"
// @Param supplier query int false "Filter by Supplier ID"
// @Param status query string false "Filter by Status (Active, Archived, Discontinued)"
// @Success 200 {object} map[string]interface{} "List of products"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	var products []domain.Product
	db := h.db.Preload("Category").Preload("SubCategory").Preload("Supplier").Preload("Location")

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Search
	if search := c.Query("search"); search != "" {
		db = db.Where("name ILIKE ? OR sku ILIKE ? OR barcode_upc ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Filters
	if categoryID := c.Query("category"); categoryID != "" {
		db = db.Where("category_id = ?", categoryID)
	}
	if supplierID := c.Query("supplier"); supplierID != "" {
		db = db.Where("supplier_id = ?", supplierID)
	}
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	if locationID := c.Query("location"); locationID != "" {
		db = db.Where("location_id = ?", locationID)
	}

	cacheKey := fmt.Sprintf("products:page:%d:limit:%d:search:%s:category:%s:supplier:%s:status:%s:location:%s",
		page, limit, c.Query("search"), c.Query("category"), c.Query("supplier"), c.Query("status"), c.Query("location"))

	// Try to get from cache first
	if cachedProducts, err := repository.GetCache(cacheKey); err == nil && cachedProducts != "" {
		var cachedResponse gin.H
		if err := json.Unmarshal([]byte(cachedProducts), &cachedResponse); err == nil {
			c.JSON(http.StatusOK, cachedResponse)
			return
		}
		logrus.Errorf("Failed to unmarshal cached products for key %s: %v", cacheKey, err)
	}

	var total int64
	h.db.Model(&domain.Product{}).Count(&total)

	if err := db.Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch products", http.StatusInternalServerError, err))
		return
	}

	response := gin.H{
		"products":     products,
		"totalItems":   total,
		"currentPage":  page,
		"totalPages":   (total + int64(limit) - 1) / int64(limit),
		"itemsPerPage": limit,
	}

	// Set to cache
	if responseJSON, err := json.Marshal(response); err == nil {
		repository.SetCache(cacheKey, responseJSON, time.Minute*5) // Cache for 5 minutes
	} else {
		logrus.Errorf("Failed to marshal products response for cache key %s: %v", cacheKey, err)
	}

	c.JSON(http.StatusOK, response)
}

// GetProduct godoc
// @Summary Get a product by ID
// @Description Get a single product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} domain.Product
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("productId")
	var product domain.Product
	cacheKey := fmt.Sprintf("product:%s", id)

	// Try to get from cache first
	if cachedProduct, err := repository.GetCache(cacheKey); err == nil && cachedProduct != "" {
		if err := json.Unmarshal([]byte(cachedProduct), &product); err == nil {
			c.JSON(http.StatusOK, product)
			return
		}
		logrus.Errorf("Failed to unmarshal cached product for key %s: %v", cacheKey, err)
	}

	if err := h.db.Preload("Category").Preload("SubCategory").Preload("Supplier").Preload("Location").First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}

	// Set to cache
	if productJSON, err := json.Marshal(product); err == nil {
		repository.SetCache(cacheKey, productJSON, time.Minute*10) // Cache for 10 minutes
	} else {
		logrus.Errorf("Failed to marshal product for cache key %s: %v", cacheKey, err)
	}

	c.JSON(http.StatusOK, product)
}

// GetProductBySKU godoc
// @Summary Get a product by SKU
// @Description Get a single product by its SKU
// @Tags products
// @Accept json
// @Produce json
// @Param sku path string true "Product SKU"
// @Success 200 {object} domain.Product
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/sku/{sku} [get]
func (h *ProductHandler) GetProductBySKU(c *gin.Context) {
	sku := c.Param("sku")
	product, err := h.productRepo.GetProductBySKU(sku)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, product)
}

// GetProductByBarcode godoc
// @Summary Get a product by Barcode
// @Description Get a single product by its Barcode
// @Tags products
// @Accept json
// @Produce json
// @Param barcode path string true "Product Barcode"
// @Success 200 {object} domain.Product
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/barcode/{barcode} [get]
func (h *ProductHandler) GetProductByBarcode(c *gin.Context) {
	barcode := c.Param("barcode")
	product, err := h.productRepo.GetProductByBarcode(barcode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, product)
}

// UpdateProduct godoc
// @Summary Update an existing product
// @Description Update an existing product with the provided details
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body requests.ProductUpdateRequest true "Product update request"
// @Success 200 {object} domain.Product
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("productId")
	var req requests.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var product domain.Product
	if err := h.db.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}

	// Update fields
	if err := h.db.Model(&product).Updates(domain.Product{
		Name:          req.Name,
		Description:   req.Description,
		CategoryID:    req.CategoryID,
		SubCategoryID: req.SubCategoryID,
		SupplierID:    req.SupplierID,
		Brand:         req.Brand,
		PurchasePrice: req.PurchasePrice,
		SellingPrice:  req.SellingPrice,
		BarcodeUPC:    req.BarcodeUPC,
		ImageURLs:     req.ImageURLs,
		Status:        req.Status,
		LocationID:    req.LocationID,
	}).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to update product", http.StatusInternalServerError, err))
		return
	}

	// Invalidate relevant caches
	repository.DeleteCache(fmt.Sprintf("product:%d", product.ID))
	repository.DeleteCache("products:*") // Invalidate all product list caches

	// Publish ProductUpdatedEvent
	payload := ProductEventPayload{
		ProductID: product.ID,
		SKU:       product.SKU,
		Name:      product.Name,
	}
	if err := message_broker.Publish(c.Request.Context(), "inventory", "product.updated", payload); err != nil {
		logrus.Errorf("Failed to publish product updated event: %v", err)
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product by its ID. Restricted if product has associated sales or stock history.
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Product has associated data"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("productId")
	var product domain.Product
	if err := h.db.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}

	// Check for associated batches
	var batchCount int64
	h.db.Model(&domain.Batch{}).Where("product_id = ?", id).Count(&batchCount)
	if batchCount > 0 {
		c.Error(appErrors.NewAppError("Cannot delete product: associated stock batches exist", http.StatusConflict, nil))
		return
	}

	// Check for associated stock adjustments
	var adjustmentCount int64
	h.db.Model(&domain.StockAdjustment{}).Where("product_id = ?", id).Count(&adjustmentCount)
	if adjustmentCount > 0 {
		c.Error(appErrors.NewAppError("Cannot delete product: associated stock adjustments exist", http.StatusConflict, nil))
		return
	}

	// For now, a simple delete
	if err := h.db.Delete(&product).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to delete product", http.StatusInternalServerError, err))
		return
	}

	// Invalidate relevant caches
	repository.DeleteCache(fmt.Sprintf("product:%d", product.ID))
	repository.DeleteCache("products:*") // Invalidate all product list caches

	// Publish ProductDeletedEvent
	payload := ProductEventPayload{
		ProductID: product.ID,
		SKU:       product.SKU,
		Name:      product.Name,
	}
	if err := message_broker.Publish(c.Request.Context(), "inventory", "product.deleted", payload); err != nil {
		logrus.Errorf("Failed to publish product deleted event: %v", err)
	}

	c.Status(http.StatusNoContent)
}

// ArchiveProduct godoc
// @Summary Archive a product
// @Description Archive a product by setting its status to 'Archived'
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} domain.Product
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/{id}/archive [patch]
func (h *ProductHandler) ArchiveProduct(c *gin.Context) {
	id := c.Param("id")
	var req requests.ProductArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var product domain.Product
	if err := h.db.First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}

	if err := h.db.Model(&product).Update("Status", req.Status).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to archive product", http.StatusInternalServerError, err))
		return
	}

	// Invalidate relevant caches
	repository.DeleteCache(fmt.Sprintf("product:%d", product.ID))
	repository.DeleteCache("products:*") // Invalidate all product list caches

	c.JSON(http.StatusOK, product)
}
