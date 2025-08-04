package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"map-memories-api/database"
	"map-memories-api/middleware"
	"map-memories-api/models"
	"map-memories-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LocationController struct{}

// CreateLocation godoc
// @Summary Create a new location
// @Description Create a new location with coordinates and details
// @Tags Locations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param location body models.LocationCreateRequest true "Location creation data"
// @Success 201 {object} models.APIResponse{data=models.LocationResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /locations [post]
func (lc *LocationController) CreateLocation(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			"Authentication required",
			"UNAUTHORIZED",
			nil,
		))
		return
	}

	var req models.LocationCreateRequest
	if err := utils.ValidateAndBindJSON(c, &req); err != nil {
		return
	}

	// Validate coordinates
	if !utils.IsValidLatitude(req.Latitude) || !utils.IsValidLongitude(req.Longitude) {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid coordinates",
			"INVALID_COORDINATES",
			nil,
		))
		return
	}

	// If custom marker is specified, validate user owns it
	if req.MarkerItemID != nil {
		var userItem models.UserItem
		if err := database.DB.Where("user_id = ? AND shop_item_id = ? AND quantity > 0",
			userID, *req.MarkerItemID).First(&userItem).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
					"You don't own this marker item",
					"MARKER_NOT_OWNED",
					nil,
				))
			} else {
				c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
					"Failed to verify marker ownership",
					"INTERNAL_ERROR",
					nil,
				))
			}
			return
		}
	}

	// Create location
	location := models.Location{
		Name:         req.Name,
		Description:  req.Description,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		Address:      req.Address,
		Country:      req.Country,
		City:         req.City,
		UserID:       userID,
		MarkerItemID: req.MarkerItemID,
	}

	if err := database.DB.Create(&location).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to create location",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	// Load marker item for response
	if location.MarkerItemID != nil {
		database.DB.Preload("MarkerItem").First(&location, location.ID)
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(
		"Location created successfully",
		location.ToResponse(),
	))
}

// GetLocations godoc
// @Summary Get locations with pagination
// @Description Get list of locations with pagination
// @Tags Locations
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param search query string false "Search in location name and description"
// @Param country query string false "Filter by country"
// @Param city query string false "Filter by city"
// @Success 200 {object} models.PaginatedResponse{data=[]models.LocationResponse}
// @Failure 500 {object} models.APIResponse
// @Router /locations [get]
func (lc *LocationController) GetLocations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// Build query (excludes soft deleted records)
	query := database.DB.Model(&models.Location{})

	// Apply filters
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if country := c.Query("country"); country != "" {
		query = query.Where("country ILIKE ?", "%"+country+"%")
	}

	if city := c.Query("city"); city != "" {
		query = query.Where("city ILIKE ?", "%"+city+"%")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get locations
	var locations []models.Location
	if err := query.Preload("MarkerItem").Limit(limit).Offset(offset).Find(&locations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to fetch locations",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	// Convert to response format with memory count
	locationResponses := make([]models.LocationResponse, len(locations))
	for i, location := range locations {
		response := location.ToResponse()

		// Count memories for this location (only non-null location_id)
		var memoryCount int64
		database.DB.Model(&models.Memory{}).Where("location_id = ? AND location_id IS NOT NULL", location.ID).Count(&memoryCount)
		response.MemoryCount = memoryCount

		locationResponses[i] = response
	}

	// Calculate pagination
	pagination := models.CalculatePagination(page, limit, total)

	c.JSON(http.StatusOK, models.PaginatedSuccessResponse(
		"Locations retrieved successfully",
		locationResponses,
		pagination,
	))
}

// GetLocation godoc
// @Summary Get location by UUID
// @Description Get a specific location by its UUID
// @Tags Locations
// @Produce json
// @Param uuid path string true "Location UUID"
// @Success 200 {object} models.APIResponse{data=models.LocationResponse}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /locations/{uuid} [get]
func (lc *LocationController) GetLocation(c *gin.Context) {
	uuidStr := c.Param("uuid")

	// Parse UUID
	locationUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	var location models.Location
	if err := database.DB.Where("uuid = ?", locationUUID).First(&location).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"Location not found",
				"LOCATION_NOT_FOUND",
				nil,
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Database error",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	response := location.ToResponse()

	// Count memories for this location (only non-null location_id)
	var memoryCount int64
	database.DB.Model(&models.Memory{}).Where("location_id = ? AND location_id IS NOT NULL", location.ID).Count(&memoryCount)
	response.MemoryCount = memoryCount

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Location retrieved successfully",
		response,
	))
}

// UpdateLocation godoc
// @Summary Update location
// @Description Update a location's details
// @Tags Locations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Location UUID"
// @Param location body models.LocationUpdateRequest true "Location update data"
// @Success 200 {object} models.APIResponse{data=models.LocationResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /locations/{uuid} [put]
func (lc *LocationController) UpdateLocation(c *gin.Context) {
	_, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			"Authentication required",
			"UNAUTHORIZED",
			nil,
		))
		return
	}

	uuidStr := c.Param("uuid")
	locationUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	var req models.LocationUpdateRequest
	if err := utils.ValidateAndBindJSON(c, &req); err != nil {
		return
	}

	// Find location (excludes soft deleted records)
	var location models.Location
	if err := database.DB.Where("uuid = ?", locationUUID).First(&location).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"Location not found",
				"LOCATION_NOT_FOUND",
				nil,
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Database error",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Update fields
	if req.Name != "" {
		location.Name = req.Name
	}
	if req.Description != "" {
		location.Description = req.Description
	}
	if req.Latitude != 0 {
		if !utils.IsValidLatitude(req.Latitude) {
			c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
				"Invalid latitude",
				"INVALID_COORDINATES",
				nil,
			))
			return
		}
		location.Latitude = req.Latitude
	}
	if req.Longitude != 0 {
		if !utils.IsValidLongitude(req.Longitude) {
			c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
				"Invalid longitude",
				"INVALID_COORDINATES",
				nil,
			))
			return
		}
		location.Longitude = req.Longitude
	}
	if req.Address != "" {
		location.Address = req.Address
	}
	if req.Country != "" {
		location.Country = req.Country
	}
	if req.City != "" {
		location.City = req.City
	}

	// Save changes
	if err := database.DB.Save(&location).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to update location",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Location updated successfully",
		location.ToResponse(),
	))
}

// DeleteLocation godoc
// @Summary Delete location (soft delete)
// @Description Soft delete a location by setting deleted_at timestamp
// @Tags Locations
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Location UUID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /locations/{uuid} [delete]
func (lc *LocationController) DeleteLocation(c *gin.Context) {
	uuidStr := c.Param("uuid")
	locationUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	// Find location (including soft deleted ones for checking)
	var location models.Location
	if err := database.DB.Unscoped().Where("uuid = ?", locationUUID).First(&location).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"Location not found",
				"LOCATION_NOT_FOUND",
				nil,
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Database error",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Check if location is already soft deleted
	if !location.DeletedAt.Time.IsZero() {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Location is already deleted",
			"LOCATION_ALREADY_DELETED",
			nil,
		))
		return
	}

	// Use transaction to ensure data consistency
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if location has associated memories
	var memoryCount int64
	tx.Model(&models.Memory{}).Where("location_id = ? AND location_id IS NOT NULL", location.ID).Count(&memoryCount)
	if memoryCount > 0 {
		// Prevent deletion if location has memories
		tx.Rollback()
		c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
			"Cannot delete location that has associated memories",
			"LOCATION_HAS_MEMORIES",
			nil,
		))
		return
	}

	// Soft delete location using GORM's soft delete
	if err := tx.Delete(&location).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to delete location",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to commit transaction",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Location deleted successfully",
		nil,
	))
}

// SearchNearbyLocations godoc
// @Summary Search locations near coordinates
// @Description Find locations within a specified radius from given coordinates
// @Tags Locations
// @Produce json
// @Param latitude query number true "Latitude" minimum(-90) maximum(90)
// @Param longitude query number true "Longitude" minimum(-180) maximum(180)
// @Param radius query number false "Radius in kilometers (default: 10, max: 100)" minimum(0) maximum(100)
// @Param limit query int false "Maximum number of results (default: 20, max: 100)" minimum(1) maximum(100)
// @Success 200 {object} models.APIResponse{data=[]models.LocationResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /locations/nearby [get]
func (lc *LocationController) SearchNearbyLocations(c *gin.Context) {
	// Parse query parameters
	latStr := c.Query("latitude")
	lngStr := c.Query("longitude")
	radiusStr := c.DefaultQuery("radius", "10")
	limitStr := c.DefaultQuery("limit", "20")

	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Latitude and longitude are required",
			"MISSING_COORDINATES",
			nil,
		))
		return
	}

	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil || !utils.IsValidLatitude(latitude) {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid latitude",
			"INVALID_LATITUDE",
			nil,
		))
		return
	}

	longitude, err := strconv.ParseFloat(lngStr, 64)
	if err != nil || !utils.IsValidLongitude(longitude) {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid longitude",
			"INVALID_LONGITUDE",
			nil,
		))
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius < 0 || radius > 100 {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid radius (must be between 0 and 100 km)",
			"INVALID_RADIUS",
			nil,
		))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid limit (must be between 1 and 100)",
			"INVALID_LIMIT",
			nil,
		))
		return
	}

	// Use PostGIS functions for geospatial search
	// ST_DWithin uses meters, so convert km to meters
	radiusMeters := radius * 1000

	var locations []models.Location
	query := `
		SELECT * FROM mm_locations 

		AND ST_DWithin(
			ST_SetSRID(ST_MakePoint(longitude, latitude), 4326),
			ST_SetSRID(ST_MakePoint(?, ?), 4326),
			?
		)
		ORDER BY ST_Distance(
			ST_SetSRID(ST_MakePoint(longitude, latitude), 4326),
			ST_SetSRID(ST_MakePoint(?, ?), 4326)
		)
		LIMIT ?
	`

	if err := database.DB.Raw(query, longitude, latitude, radiusMeters, longitude, latitude, limit).Scan(&locations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to search nearby locations",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	// Convert to response format with memory count and distance
	locationResponses := make([]models.LocationResponse, len(locations))
	for i, location := range locations {
		response := location.ToResponse()

		// Count memories for this location (only non-null location_id)
		var memoryCount int64
		database.DB.Model(&models.Memory{}).Where("location_id = ? AND location_id IS NOT NULL", location.ID).Count(&memoryCount)
		response.MemoryCount = memoryCount

		locationResponses[i] = response
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		fmt.Sprintf("Found %d locations within %.1f km", len(locations), radius),
		locationResponses,
	))
}

// GetLocationMemories godoc
// @Summary Get memories for a location
// @Description Get all memories associated with a specific location
// @Tags Locations
// @Produce json
// @Param uuid path string true "Location UUID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param is_public query bool false "Filter by public status"
// @Success 200 {object} models.PaginatedResponse{data=[]models.MemoryResponse}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /locations/{uuid}/memories [get]
func (lc *LocationController) GetLocationMemories(c *gin.Context) {
	uuidStr := c.Param("uuid")
	locationUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	// Verify location exists (excludes soft deleted records)
	var location models.Location
	if err := database.DB.Where("uuid = ?", locationUUID).First(&location).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"Location not found",
				"LOCATION_NOT_FOUND",
				nil,
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Database error",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// Build query for memories
	query := database.DB.Preload("User").Preload("Location").Preload("Media").
		Where("location_id = ? AND location_id IS NOT NULL", location.ID)

	// Apply public filter
	if isPublicStr := c.Query("is_public"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			query = query.Where("is_public = ?", isPublic)
		}
	}

	// Get total count
	var total int64
	query.Model(&models.Memory{}).Count(&total)

	// Get memories
	var memories []models.Memory
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&memories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to fetch memories",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	// Convert to response format
	memoryResponses := make([]models.MemoryResponse, len(memories))
	for i, memory := range memories {
		memoryResponses[i] = memory.ToResponse()
	}

	// Calculate pagination
	pagination := models.CalculatePagination(page, limit, total)

	c.JSON(http.StatusOK, models.PaginatedSuccessResponse(
		"Location memories retrieved successfully",
		memoryResponses,
		pagination,
	))
}
