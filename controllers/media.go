package controllers

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"map-memories-api/database"
	"map-memories-api/middleware"
	"map-memories-api/models"
	"map-memories-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MediaController struct{}

// UploadMedia godoc
// @Summary Upload media file
// @Description Upload an image or video file for a memory. The file will be automatically converted to base64 and stored in the database.
// @Tags Media
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param memory_id formData int true "Memory ID"
// @Param display_order formData int false "Display order (default: 0)"
// @Param file formData file true "Media file"
// @Success 201 {object} models.APIResponse{data=models.MediaResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 413 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /media/upload [post]
func (mc *MediaController) UploadMedia(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			"Authentication required",
			"UNAUTHORIZED",
			nil,
		))
		return
	}

	// Parse form data
	memoryIDStr := c.PostForm("memory_id")
	displayOrderStr := c.DefaultPostForm("display_order", "0")

	memoryID, err := strconv.Atoi(memoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid memory ID",
			"INVALID_MEMORY_ID",
			nil,
		))
		return
	}

	displayOrder, err := strconv.Atoi(displayOrderStr)
	if err != nil {
		displayOrder = 0
	}

	// Verify memory exists and belongs to user
	var memory models.Memory
	if err := database.DB.Where("id = ? AND user_id = ?", memoryID, userID).First(&memory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(
				"Memory not found or access denied",
				"MEMORY_ACCESS_DENIED",
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

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"No file uploaded",
			"NO_FILE_UPLOADED",
			nil,
		))
		return
	}

	// Save file as base64
	fileInfo, err := utils.SaveUploadedFileAsBase64(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Failed to convert file to base64: "+err.Error(),
			"FILE_CONVERSION_ERROR",
			nil,
		))
		return
	}

	// Create media record
	media := models.Media{
		MemoryID:         uint(memoryID),
		Filename:         fileInfo.Filename,
		OriginalFilename: fileInfo.OriginalFilename,
		FilePath:         fileInfo.FilePath,
		FileSize:         fileInfo.FileSize,
		MimeType:         fileInfo.MimeType,
		MediaType:        fileInfo.MediaType,
		DisplayOrder:     displayOrder,
	}

	if err := database.DB.Create(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to save media record",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(
		"Media uploaded successfully",
		media.ToResponse(),
	))
}

// GetMedia godoc
// @Summary Get media list
// @Description Get list of media files with optional filters
// @Tags Media
// @Produce json
// @Param memory_id query int false "Filter by memory ID"
// @Param media_type query string false "Filter by media type (image, video)" Enums(image, video)
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} models.PaginatedResponse{data=[]models.MediaResponse}
// @Failure 500 {object} models.APIResponse
// @Router /media [get]
func (mc *MediaController) GetMedia(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// Build query
	query := database.DB.Preload("Memory").Model(&models.Media{})

	// Apply filters
	if memoryIDStr := c.Query("memory_id"); memoryIDStr != "" {
		if memoryID, err := strconv.Atoi(memoryIDStr); err == nil {
			query = query.Where("memory_id = ?", memoryID)
		}
	}

	if mediaType := c.Query("media_type"); mediaType != "" {
		if mediaType == "image" || mediaType == "video" {
			query = query.Where("media_type = ?", mediaType)
		}
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get media
	var mediaList []models.Media
	if err := query.Order("display_order ASC, created_at DESC").
		Limit(limit).Offset(offset).Find(&mediaList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to fetch media",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	// Convert to response format
	mediaResponses := make([]models.MediaResponse, len(mediaList))
	for i, media := range mediaList {
		mediaResponses[i] = media.ToResponse()
	}

	// Calculate pagination
	pagination := models.CalculatePagination(page, limit, total)

	c.JSON(http.StatusOK, models.PaginatedSuccessResponse(
		"Media retrieved successfully",
		mediaResponses,
		pagination,
	))
}

// GetMediaFile godoc
// @Summary Get media file by UUID
// @Description Get a specific media file by its UUID
// @Tags Media
// @Produce json
// @Param uuid path string true "Media UUID"
// @Success 200 {object} models.APIResponse{data=models.MediaResponse}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /media/{uuid} [get]
func (mc *MediaController) GetMediaFile(c *gin.Context) {
	uuidStr := c.Param("uuid")

	// Parse UUID
	mediaUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	var media models.Media
	if err := database.DB.Preload("Memory").Where("uuid = ?", mediaUUID).First(&media).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"Media not found",
				"MEDIA_NOT_FOUND",
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

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Media retrieved successfully",
		media.ToResponse(),
	))
}

// ServeMediaFile godoc
// @Summary Serve media file
// @Description Serve the actual media file for viewing/download
// @Tags Media
// @Produce application/octet-stream
// @Param uuid path string true "Media UUID"
// @Success 200 {file} file "Media file"
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /media/{uuid}/file [get]
func (mc *MediaController) ServeMediaFile(c *gin.Context) {
	uuidStr := c.Param("uuid")

	// Parse UUID
	mediaUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	var media models.Media
	if err := database.DB.Where("uuid = ?", mediaUUID).First(&media).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"Media not found",
				"MEDIA_NOT_FOUND",
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

	// Check if file_path contains base64 data URL
	if !strings.HasPrefix(media.FilePath, "data:") {
		c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
			"Media file format not supported",
			"FILE_FORMAT_ERROR",
			nil,
		))
		return
	}

	// Extract base64 data from data URL
	parts := strings.Split(media.FilePath, ",")
	if len(parts) != 2 {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Invalid base64 data format",
			"INVALID_BASE64_FORMAT",
			nil,
		))
		return
	}

	// Decode base64 data
	fileData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to decode base64 data",
			"BASE64_DECODE_ERROR",
			err.Error(),
		))
		return
	}

	// Set appropriate headers
	c.Header("Content-Type", media.MimeType)
	c.Header("Content-Disposition", "inline; filename=\""+media.OriginalFilename+"\"")
	c.Header("Cache-Control", "public, max-age=31536000") // 1 year cache

	// Serve base64 decoded data
	c.Data(http.StatusOK, media.MimeType, fileData)
}

// UpdateMedia godoc
// @Summary Update media
// @Description Update media display order (only by owner)
// @Tags Media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Media UUID"
// @Param media body models.MediaUpdateRequest true "Media update data"
// @Success 200 {object} models.APIResponse{data=models.MediaResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /media/{uuid} [put]
func (mc *MediaController) UpdateMedia(c *gin.Context) {
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
	mediaUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	var req models.MediaUpdateRequest
	if err := utils.ValidateAndBindJSON(c, &req); err != nil {
		return
	}

	// Find media with memory to check ownership
	var media models.Media
	if err := database.DB.Preload("Memory").Where("uuid = ?", mediaUUID).First(&media).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"Media not found",
				"MEDIA_NOT_FOUND",
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

	// Check ownership through memory
	if media.Memory.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(
			"Access denied: You can only update your own media",
			"FORBIDDEN",
			nil,
		))
		return
	}

	// Update fields
	media.DisplayOrder = req.DisplayOrder

	// Save changes
	if err := database.DB.Save(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to update media",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Media updated successfully",
		media.ToResponse(),
	))
}

// DeleteMedia godoc
// @Summary Delete media
// @Description Delete a media file (only by owner)
// @Tags Media
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Media UUID"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /media/{uuid} [delete]
func (mc *MediaController) DeleteMedia(c *gin.Context) {
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
	mediaUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	// Find media with memory to check ownership
	var media models.Media
	if err := database.DB.Preload("Memory").Where("uuid = ?", mediaUUID).First(&media).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"Media not found",
				"MEDIA_NOT_FOUND",
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

	// Check ownership through memory
	if media.Memory.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(
			"Access denied: You can only delete your own media",
			"FORBIDDEN",
			nil,
		))
		return
	}

	// No need to delete file from filesystem since it's stored as base64 in database

	// Delete media record
	if err := database.DB.Delete(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to delete media record",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Media deleted successfully",
		nil,
	))
}

// GetMemoryMedia godoc
// @Summary Get media for a memory
// @Description Get all media files associated with a specific memory
// @Tags Media
// @Produce json
// @Param uuid path string true "Memory UUID"
// @Success 200 {object} models.APIResponse{data=[]models.MediaResponse}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /memories/{uuid}/media [get]
func (mc *MediaController) GetMemoryMedia(c *gin.Context) {
	memoryUuidStr := c.Param("uuid")

	// Parse UUID
	memoryUUID, err := uuid.Parse(memoryUuidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid UUID format",
			"INVALID_UUID",
			nil,
		))
		return
	}

	// Verify memory exists
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

	// Get media for this memory
	var mediaList []models.Media
	if err := database.DB.Where("memory_id = ?", memory.ID).
		Order("display_order ASC, created_at ASC").Find(&mediaList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to fetch media",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	// Convert to response format
	mediaResponses := make([]models.MediaResponse, len(mediaList))
	for i, media := range mediaList {
		mediaResponses[i] = media.ToResponse()
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Memory media retrieved successfully",
		mediaResponses,
	))
}
