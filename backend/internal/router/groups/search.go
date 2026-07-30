package groups

import (
	"inventory/backend/internal/app"

	"github.com/gin-gonic/gin"
)

func RegisterSearch(rg *gin.RouterGroup, c *app.Container) {
	searchHandler := c.SearchHandler

	rg.GET("/search", searchHandler.Search)
}
