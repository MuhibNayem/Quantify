package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterPublic(r *gin.Engine, c *app.Container, roleRepo repository.RoleRepository) {
	readyHandler := handlers.NewReadyHandler(c.DB, c.Redis, c.AMQP, c.Uploader)
	paymentHandler := c.PaymentHandler
	userHandler := c.User

	publicRoutes := r.Group("/")
	{
		publicRoutes.GET("/health", handlers.HealthCheck)
		publicRoutes.GET("/ready", readyHandler.Ready)
		publicRoutes.GET("/ws", func(ctx *gin.Context) {
			handlers.ServeWs(c.Hub, ctx, c.Notifications, roleRepo, c.Cfg)
		})
		publicRoutes.GET("/metrics", gin.WrapH(promhttp.Handler()))
		publicRoutes.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	webhookRoutes := r.Group("/webhooks")
	{
		webhookRoutes.POST("/:name", c.Integrations.HandleWebhook)
	}

	paymentRoutes := r.Group("/payment")
	{
		paymentRoutes.POST("/create", paymentHandler.CreatePayment)
		paymentRoutes.POST("/success", paymentHandler.PaymentSuccess)
		paymentRoutes.POST("/fail", paymentHandler.PaymentFail)
		paymentRoutes.POST("/cancel", paymentHandler.PaymentCancel)
		paymentRoutes.POST("/ipn", paymentHandler.HandleSSLCommerzIPN)
	}

	authLimit := middleware.RateLimit(c.Cfg.AuthRateLimitRequests, c.Cfg.AuthRateLimitWindow, "auth")

	publicAPI := r.Group("/api/v1")
	{
		userRoutes := publicAPI.Group("/users")
		{
			userRoutes.POST("/register", authLimit, userHandler.RegisterUser)
			userRoutes.POST("/login", authLimit, userHandler.LoginUser)
			// Refresh and logout must work when the access cookie has expired;
			// both authenticate from the refresh cookie / session cookies.
			userRoutes.POST("/refresh-token", authLimit, userHandler.RefreshToken)
			userRoutes.POST("/logout", userHandler.LogoutUser)
		}
	}
}
