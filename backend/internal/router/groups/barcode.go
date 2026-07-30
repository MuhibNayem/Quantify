package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterBarcode(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	barcodeHandler := c.BarcodeHandler

	barcode := rg.Group("/barcode")
	{
		barcode.GET("/lookup", middleware.RequirePermission(roleRepo, "barcode.read"), barcodeHandler.LookupProductByBarcode)
		barcode.GET("/generate", middleware.RequirePermission(roleRepo, "barcode.read"), barcodeHandler.GenerateBarcode)
	}
}
