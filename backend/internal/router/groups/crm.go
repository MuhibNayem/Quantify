package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterCRM(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	crmHandler := c.CRMHandler

	crm := rg.Group("/crm")
	{
		crm.GET("/customers", middleware.RequirePermission(roleRepo, "customers.read"), crmHandler.ListCustomers)
		crm.POST("/customers", middleware.RequirePermission(roleRepo, "customers.write"), crmHandler.CreateCustomer)
		crm.GET("/customers/:identifier", middleware.RequirePermission(roleRepo, "customers.read"), crmHandler.GetCustomer)
		crm.GET("/customers/username/:username", middleware.RequirePermission(roleRepo, "customers.read"), crmHandler.GetCustomerByUsername)
		crm.GET("/customers/email/:email", middleware.RequirePermission(roleRepo, "customers.read"), crmHandler.GetCustomerByEmail)
		crm.GET("/customers/phone/:phone", middleware.RequirePermission(roleRepo, "customers.read"), crmHandler.GetCustomerByPhone)
		crm.PUT("/customers/:identifier", middleware.RequirePermission(roleRepo, "customers.write"), crmHandler.UpdateCustomer)
		crm.DELETE("/customers/:identifier", middleware.RequirePermission(roleRepo, "customers.write"), crmHandler.DeleteCustomer)
		crm.GET("/loyalty/:userId", middleware.RequirePermission(roleRepo, "loyalty.read"), crmHandler.GetLoyaltyAccount)
		crm.POST("/loyalty/:userId/points", middleware.RequirePermission(roleRepo, "loyalty.write"), crmHandler.AddLoyaltyPoints)
		crm.POST("/loyalty/:userId/redeem", middleware.RequirePermission(roleRepo, "loyalty.write"), crmHandler.RedeemLoyaltyPoints)
		crm.GET("/customers/:identifier/churn-risk", middleware.RequirePermission(roleRepo, "customers.read"), crmHandler.GetChurnRisk)
	}
}
