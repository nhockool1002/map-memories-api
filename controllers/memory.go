package controllers

import (
	"net/http"
	"strconv"

	"map-memories-api/database"
	"map-memories-api/middleware"
	"map-memories-api/models"
	"map-memories-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type MemoryController struct{}

// CreateMemory godoc
// @Summary Create a new memory
// @Description Create a new memory with title, content, and location
// @Tags Memories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param memory body models.MemoryCreateRequest true "Memory creation data"
// @Success 201 {object} models.APIResponse{data=models.MemoryResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /memories [post]
func (mc *MemoryController) CreateMemory(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			"Authentication required",
			"UNAUTHORIZED",
			nil,
		))
		return
	}

	var req models.MemoryCreateRequest
	if err := utils.ValidateAndBindJSON(c, &req); err != nil {
		return
	}

	// Verify location exists
	var location models.Location
	if err := database.DB.First(&location, req.LocationID).Error; err != nil {
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

	// Create memory
	memory := models.Memory{
		UserID:     userID,
		LocationID: req.LocationID,
		Title:      req.Title,
		Content:    req.Content,
		VisitDate:  req.VisitDate,
		IsPublic:   req.IsPublic,
		Tags:       pq.StringArray(req.Tags),
	}

	if err := database.DB.Create(&memory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to create memory",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	// Load relationships
	database.DB.Preload("User").Preload("Location").First(&memory, memory.ID)

	c.JSON(http.StatusCreated, models.SuccessResponse(
		"Memory created successfully",
		memory.ToResponse(),
	))
}

// GetMemories godoc
// @Summary Get memories with pagination and filters
// @Description Get list of memories with optional filters and pagination
// @Tags Memories
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param user_id query int false "Filter by user ID"
// @Param location_id query int false "Filter by location ID"
// @Param is_public query bool false "Filter by public status"
// @Param search query string false "Search in title and content"
// @Param tags query string false "Filter by tags (comma-separated)"
// @Param sort_by query string false "Sort field (created_at, visit_date, title)" Enums(created_at, visit_date, title)
// @Param sort_order query string false "Sort order (asc, desc)" Enums(asc, desc)
// @Success 200 {object} models.PaginatedResponse{data=[]models.MemoryResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /memories [get]
func (mc *MemoryController) GetMemories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if limit > 100 {
		limit = 100
	}
	
	offset := (page - 1) * limit

	// Build query
	query := database.DB.Preload("User").Preload("Location").Preload("Media")

	// Apply filters
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.Atoi(userIDStr); err == nil {
			query = query.Where("user_id = ?", userID)
		}
	}

	if locationIDStr := c.Query("location_id"); locationIDStr != "" {
		if locationID, err := strconv.Atoi(locationIDStr); err == nil {
			query = query.Where("location_id = ?", locationID)
		}
	}

	if isPublicStr := c.Query("is_public"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			query = query.Where("is_public = ?", isPublic)
		}
	}

	if search := c.Query("search"); search != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if tagsStr := c.Query("tags"); tagsStr != "" {
		// Handle tag filtering - search for any of the provided tags
		query = query.Where("tags && ?", pq.StringArray{tagsStr})
	}

	// Apply sorting
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	
	if sortBy == "visit_date" || sortBy == "title" || sortBy == "created_at" {
		query = query.Order(sortBy + " " + sortOrder)
	}

	// Get total count
	var total int64
	countQuery := query
	countQuery.Model(&models.Memory{}).Count(&total)

	// Get memories
	var memories []models.Memory
	if err := query.Limit(limit).Offset(offset).Find(&memories).Error; err != nil {
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
		"Memories retrieved successfully",
		memoryResponses,
		pagination,
	))
}

// GetMemory godoc
// @Summary Get memory by UUID
// @Description Get a specific memory by its UUID
// @Tags Memories
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Memory UUID"
// @Success 200 {object} models.APIResponse{data=models.MemoryResponse}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /memories/{uuid} [get]
func (mc *MemoryController) GetMemory(c *gin.Context) {
	uuidStr := c.Param("uuid")
	
	// Parse UUID
	memoryUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	var memory models.Memory
	if err := database.DB.Preload("User").Preload("Location").Preload("Media").
		Where("uuid = ?", memoryUUID).First(&memory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"Memory not found",
				"MEMORY_NOT_FOUND",
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

	// Check if memory is private and user is not the owner
	currentUserID, authenticated := middleware.GetCurrentUserID(c)
	if !memory.IsPublic && (!authenticated || memory.UserID != currentUserID) {
		c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(
			"Access denied",
			"FORBIDDEN",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Memory retrieved successfully",
		memory.ToResponse(),
	))
}

// UpdateMemory godoc
// @Summary Update memory
// @Description Update a memory (only by owner)
// @Tags Memories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Memory UUID"
// @Param memory body models.MemoryUpdateRequest true "Memory update data"
// @Success 200 {object} models.APIResponse{data=models.MemoryResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /memories/{uuid} [put]
func (mc *MemoryController) UpdateMemory(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			"Authentication required",
			"UNAUTHORIZED",
			nil,
		))
		return
	}

	uuidStr := c.Param("uuid")
	memoryUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	var req models.MemoryUpdateRequest
	if err := utils.ValidateAndBindJSON(c, &req); err != nil {
		return
	}

	// Find memory
	var memory models.Memory
	if err := database.DB.Where("uuid = ?", memoryUUID).First(&memory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"Memory not found",
				"MEMORY_NOT_FOUND",
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

	// Check ownership
	if memory.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(
			"Access denied: You can only update your own memories",
			"FORBIDDEN",
			nil,
		))
		return
	}

	// Update fields
	if req.Title != "" {
		memory.Title = req.Title
	}
	if req.Content != "" {
		memory.Content = req.Content
	}
	if req.VisitDate != nil {
		memory.VisitDate = req.VisitDate
	}
	memory.IsPublic = req.IsPublic
	if req.Tags != nil {
		memory.Tags = pq.StringArray(req.Tags)
	}

	// Save changes
	if err := database.DB.Save(&memory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to update memory",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	// Load relationships
	database.DB.Preload("User").Preload("Location").Preload("Media").First(&memory, memory.ID)

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Memory updated successfully",
		memory.ToResponse(),
	))
}

// DeleteMemory godoc
// @Summary Delete memory
// @Description Delete a memory (only by owner)
// @Tags Memories
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Memory UUID"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /memories/{uuid} [delete]
func (mc *MemoryController) DeleteMemory(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			"Authentication required",
			"UNAUTHORIZED",
			nil,
		))
		return
	}

	uuidStr := c.Param("uuid")
	memoryUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	// Find memory
	var memory models.Memory
	if err := database.DB.Where("uuid = ?", memoryUUID).First(&memory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"Memory not found",
				"MEMORY_NOT_FOUND",
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

	// Check ownership
	if memory.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(
			"Access denied: You can only delete your own memories",
			"FORBIDDEN",
			nil,
		))
		return
	}

	// Delete memory (soft delete)
	if err := database.DB.Delete(&memory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to delete memory",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Memory deleted successfully",
		nil,
	))
}