package handlers

import (
	"net/http"

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
// @Description Generates a barcode image (PNG) for a product SKU
// @Tags barcodes
// @Accept json
// @Produce image/png
// @Param sku query string true "Product SKU"
// @Security BearerAuth
// @Success 200 {file} image/png "Barcode image"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /barcode/generate [get]
func (h *BarcodeHandler) GenerateBarcode(c *gin.Context) {
	sku := c.Query("sku")
	if sku == "" {
		c.Error(appErrors.NewAppError("Product SKU is required", http.StatusBadRequest, nil))
		return
	}

	buf, err := h.barcodeService.GenerateBarcode(sku)
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
// @Security BearerAuth
// @Success 200 {object} domain.Product
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /barcode/lookup [get]
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

// ResolveBarcode godoc
// @Summary Resolve a barcode to a product and selling unit
// @Description Resolves a product barcode, SKU, or active selling-unit barcode.
// @Tags barcodes
// @Accept json
// @Produce json
// @Param barcode query string true "Barcode, UPC, or SKU"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "product and optional unit"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Router /barcode/resolve [get]
func (h *BarcodeHandler) ResolveBarcode(c *gin.Context) {
	barcodeValue := c.Query("barcode")
	if barcodeValue == "" {
		c.Error(appErrors.NewAppError("Barcode value is required", http.StatusBadRequest, nil))
		return
	}

	product, unit, err := h.barcodeService.ResolveBarcode(barcodeValue)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found with provided barcode", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to resolve barcode", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product, "unit": unit})
}
