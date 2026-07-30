package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterJobs(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	reportHandler := c.Report

	jobs := rg.Group("/jobs")
	jobs.Use(middleware.AuthorizeMiddleware("Admin", "Manager"))
	{
		jobs.POST("/:jobId/cancel", reportHandler.CancelJob)
	}
}
