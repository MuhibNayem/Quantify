package handlers

import (
	"inventory/backend/internal/errors"
	"inventory/backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService *services.SearchService
}

func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

// Search godoc
// @Summary Perform a global search
// @Description Searches across products, users, suppliers, and categories.
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {array} repository.SearchResult
// @Failure 400 {object} map[string]interface{} "Bad Request: Query parameter 'q' is required"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.Error(errors.NewAppError("Query parameter 'q' is required", http.StatusBadRequest, nil))
		return
	}

	results, err := h.searchService.Search(query)
	if err != nil {
		c.Error(errors.NewAppError("Failed to perform search", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, results)
}
