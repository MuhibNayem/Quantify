package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/services"
)

type BarcodeHandler struct {
	barcodeService services.BarcodeService
}

func NewBarcodeHandler(barcodeService services.BarcodeService) *BarcodeHandler {
	return &BarcodeHandler{barcodeService: barcodeService}
}

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
func (h *BarcodeHandler) GenerateBarcode(c *gin.Context) {
	sku := c.Query("sku")
	productIDStr := c.Query("productId")

	var productID uint64
	var err error

	if productIDStr != "" {
		productID, err = strconv.ParseUint(productIDStr, 10, 64)
		if err != nil {
			c.Error(appErrors.NewAppError("Invalid product ID", http.StatusBadRequest, err))
			return
		}
	}

	if sku == "" && productID == 0 {
		c.Error(appErrors.NewAppError("Either 'sku' or 'productId' is required", http.StatusBadRequest, nil))
		return
	}

	buf, err := h.barcodeService.GenerateBarcode(sku, productID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to generate barcode", http.StatusInternalServerError, err))
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
func (h *BarcodeHandler) LookupProductByBarcode(c *gin.Context) {
	barcodeValue := c.Query("barcode")
	if barcodeValue == "" {
		c.Error(appErrors.NewAppError("Barcode value is required", http.StatusBadRequest, nil))
		return
	}

	product, err := h.barcodeService.LookupProductByBarcode(barcodeValue)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found with provided barcode", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to lookup product", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, product)
}
