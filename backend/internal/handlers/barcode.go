package handlers

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128" // Example barcode type
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"image/png"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
)

// GenerateBarcode godoc
// @Summary Generate a barcode image for a product
// @Description Generates a barcode image (PNG) for a given product SKU or ID
// @Tags barcodes
// @Accept json
// @Produce image/png
// @Param sku query string false "Product SKU"
// @Param productId query int false "Product ID"
// @Success 200 {file} image/png "Barcode image"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /barcodes/generate [get]
func GenerateBarcode(c *gin.Context) {
	sku := c.Query("sku")
	productIDStr := c.Query("productId")

	var product domain.Product
	var err error
	locationID := c.Query("locationId")

	db := repository.DB
	if locationID != "" {
		db = db.Where("location_id = ?", locationID)
	}

	if sku != "" {
		err = db.Where("sku = ?", sku).First(&product).Error
	} else if productIDStr != "" {
		productID, parseErr := strconv.ParseUint(productIDStr, 10, 64)
		if parseErr != nil {
			c.Error(appErrors.NewAppError("Invalid product ID", http.StatusBadRequest, parseErr))
			return
		}
		err = db.First(&product, productID).Error
	} else {
		c.Error(appErrors.NewAppError("Either 'sku' or 'productId' is required", http.StatusBadRequest, nil))
		return
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}

	// Use SKU or BarcodeUPC for barcode content
	content := product.SKU
	if product.BarcodeUPC != "" {
		content = product.BarcodeUPC
	}

	// Create the barcode
	var bcode barcode.Barcode
	bcode, err = code128.Encode(content)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to encode barcode", http.StatusInternalServerError, err))
		return
	}

	// Scale the barcode
	bcode, err = barcode.Scale(bcode, 200, 50) // Width, Height
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to scale barcode", http.StatusInternalServerError, err))
		return
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, bcode); err != nil {
		c.Error(appErrors.NewAppError("Failed to encode barcode to PNG", http.StatusInternalServerError, err))
		return
	}

	c.Data(http.StatusOK, "image/png", buf.Bytes())
}

// LookupProductByBarcode godoc
// @Summary Lookup a product by barcode/UPC
// @Description Retrieves product details by scanning its barcode or UPC
// @Tags barcodes
// @Accept json
// @Produce json
// @Param barcode query string true "Barcode or UPC value"
// @Success 200 {object} domain.Product
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/lookup [get]
func LookupProductByBarcode(c *gin.Context) {
	barcodeValue := c.Query("barcode")
	if barcodeValue == "" {
		c.Error(appErrors.NewAppError("Barcode value is required", http.StatusBadRequest, nil))
		return
	}

	var product domain.Product
	db := repository.DB.Preload("Category").Preload("SubCategory").Preload("Supplier").Preload("Location")

	if locationID := c.Query("locationId"); locationID != "" {
		db = db.Where("location_id = ?", locationID)
	}

	if err := db.Where("sku = ? OR barcode_upc = ?", barcodeValue, barcodeValue).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found with provided barcode", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to lookup product", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, product)
}
