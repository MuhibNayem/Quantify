package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
)

// CreateLocation godoc
// @Summary Create a new location
// @Description Create a new inventory location
// @Tags locations
// @Accept json
// @Produce json
// @Param location body requests.LocationCreateRequest true "Location creation request"
// @Success 201 {object} domain.Location
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 409 {object} map[string]interface{} "Conflict: Location with this name already exists"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /locations [post]
func CreateLocation(c *gin.Context) {
	var req requests.LocationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	location := domain.Location{
		Name:    req.Name,
		Address: req.Address,
	}

	if err := repository.DB.Create(&location).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			c.Error(appErrors.NewAppError("Location with this name already exists", http.StatusConflict, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to create location", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusCreated, location)
}

// ListLocations godoc
// @Summary Get a list of locations
// @Description Get a list of all inventory locations
// @Tags locations
// @Accept json
// @Produce json
// @Success 200 {array} domain.Location
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /locations [get]
func ListLocations(c *gin.Context) {
	var locations []domain.Location
	if err := repository.DB.Find(&locations).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch locations", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, locations)
}

// GetLocation godoc
// @Summary Get a location by ID
// @Description Get a single location by its ID
// @Tags locations
// @Accept json
// @Produce json
// @Param id path int true "Location ID"
// @Success 200 {object} domain.Location
// @Failure 404 {object} map[string]interface{} "Location not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /locations/{id} [get]
func GetLocation(c *gin.Context) {
	id := c.Param("id")
	var location domain.Location
	if err := repository.DB.First(&location, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Location not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch location", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, location)
}

// UpdateLocation godoc
// @Summary Update an existing location
// @Description Update an existing inventory location
// @Tags locations
// @Accept json
// @Produce json
// @Param id path int true "Location ID"
// @Param location body requests.LocationUpdateRequest true "Location update request"
// @Success 200 {object} domain.Location
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Location not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Location with this name already exists"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /locations/{id} [put]
func UpdateLocation(c *gin.Context) {
	id := c.Param("id")
	var req requests.LocationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var location domain.Location
	if err := repository.DB.First(&location, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Location not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch location", http.StatusInternalServerError, err))
		return
	}

	if err := repository.DB.Model(&location).Updates(domain.Location{
		Name:    req.Name,
		Address: req.Address,
	}).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			c.Error(appErrors.NewAppError("Location with this name already exists", http.StatusConflict, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to update location", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, location)
}

// DeleteLocation godoc
// @Summary Delete a location
// @Description Delete an inventory location by its ID. Cannot delete if products, batches, or stock adjustments are associated.
// @Tags locations
// @Accept json
// @Produce json
// @Param id path int true "Location ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{} "Location not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Location has associated data"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /locations/{id} [delete]
func DeleteLocation(c *gin.Context) {
	id := c.Param("id")
	var location domain.Location
	if err := repository.DB.First(&location, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Location not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch location", http.StatusInternalServerError, err))
		return
	}

	// Check for associated products
	var productCount int64
	repository.DB.Model(&domain.Product{}).Where("location_id = ?", id).Count(&productCount)
	if productCount > 0 {
		c.Error(appErrors.NewAppError("Cannot delete location: products are associated", http.StatusConflict, nil))
		return
	}

	// Check for associated batches
	var batchCount int64
	repository.DB.Model(&domain.Batch{}).Where("location_id = ?", id).Count(&batchCount)
	if batchCount > 0 {
		c.Error(appErrors.NewAppError("Cannot delete location: batches are associated", http.StatusConflict, nil))
		return
	}

	// Check for associated stock adjustments
	var adjustmentCount int64
	repository.DB.Model(&domain.StockAdjustment{}).Where("location_id = ?", id).Count(&adjustmentCount)
	if adjustmentCount > 0 {
		c.Error(appErrors.NewAppError("Cannot delete location: stock adjustments are associated", http.StatusConflict, nil))
		return
	}

	if err := repository.DB.Delete(&location).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to delete location", http.StatusInternalServerError, err))
		return
	}

	c.Status(http.StatusNoContent)
}
