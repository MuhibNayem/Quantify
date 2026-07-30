package router

import (
	"time"

	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/router/groups"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	_ "inventory/backend/api"
)

func SetupRouter(c *app.Container) *gin.Engine {
	r := gin.Default()

	corsConfig := cors.Config{
		AllowOrigins:     c.Cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Authorization", "X-Requested-With", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowWebSockets:  true,
	}

	r.Use(middleware.RequestID())
	r.Use(otelgin.Middleware("quantify-backend"))
	r.Use(middleware.Metrics())
	r.Use(cors.New(corsConfig))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.ErrorHandler())

	roleRepo := c.Roles

	groups.RegisterPublic(r, c, roleRepo)

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	api.Use(middleware.CSRFMiddleware())

	groups.RegisterJobs(api, c, roleRepo)
	groups.RegisterDashboard(api, c, roleRepo)
	groups.RegisterProducts(api, c, roleRepo)
	groups.RegisterUnits(api, c, roleRepo)
	groups.RegisterPromotions(api, c, roleRepo)
	groups.RegisterCategories(api, c, roleRepo)
	groups.RegisterSubCategories(api, c, roleRepo)
	groups.RegisterSuppliers(api, c, roleRepo)
	groups.RegisterLocations(api, c, roleRepo)
	groups.RegisterBarcode(api, c, roleRepo)
	groups.RegisterReplenishment(api, c, roleRepo)
	groups.RegisterSales(api, c, roleRepo)
	groups.RegisterReturns(api, c, roleRepo)
	groups.RegisterReports(api, c, roleRepo)
	groups.RegisterFinance(api, c, roleRepo)
	groups.RegisterAlerts(api, c, roleRepo)
	groups.RegisterBulk(api, c, roleRepo)
	groups.RegisterInventory(api, c, roleRepo)
	groups.RegisterUsers(api, c, roleRepo)
	groups.RegisterNotifications(api, c, roleRepo)
	groups.RegisterCRM(api, c, roleRepo)
	groups.RegisterTimeTracking(api, c, roleRepo)
	groups.RegisterSearch(api, c)
	groups.RegisterSettings(api, c, roleRepo)
	groups.RegisterRoles(api, c, roleRepo)
	groups.RegisterPermissions(api, c, roleRepo)

	return r
}
