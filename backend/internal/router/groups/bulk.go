package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterBulk(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	bulkHandler := c.Bulk

	bulk := rg.Group("/bulk")
	{
		bulk.GET("/products/template", middleware.RequirePermission(roleRepo, "bulk.import"), bulkHandler.GetProductImportTemplate)
		bulk.POST("/products/import", middleware.RequirePermission(roleRepo, "bulk.import"), bulkHandler.UploadProductImport)
		bulk.GET("/products/import/:jobId/status", middleware.RequirePermission(roleRepo, "bulk.import"), bulkHandler.GetBulkImportStatus)
		bulk.POST("/products/import/:jobId/confirm", middleware.RequirePermission(roleRepo, "bulk.import"), bulkHandler.ConfirmBulkImport)
		bulk.GET("/products/export", middleware.RequirePermission(roleRepo, "bulk.export"), bulkHandler.ExportProducts)
		bulk.GET("/jobs", middleware.RequirePermission(roleRepo, "bulk.import"), bulkHandler.ListBulkJobs)
		bulk.GET("/files/:bucket/:object", middleware.RequirePermission(roleRepo, "bulk.export"), bulkHandler.DownloadFile)
	}
}
