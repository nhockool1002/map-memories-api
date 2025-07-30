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
	"gorm.io/gorm"
)

type ShopController struct{}

// @Summary Admin: Create shop item
// @Description Admin can create new items in the shop
// @Tags Shop
// @Accept json
// @Produce json
// @Param request body models.ShopItemCreateRequest true "Shop item creation request"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /admin/shop/items [post]
func (sc *ShopController) CreateShopItem(c *gin.Context) {
	var req models.ShopItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponseWithCode(
			"Invalid request data",
			"INVALID_REQUEST",
			map[string]interface{}{"error": err.Error()},
		))
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponseWithCode(
			"Validation failed",
			"VALIDATION_ERROR",
			map[string]interface{}{"errors": err},
		))
		return
	}

	shopItem := models.ShopItem{
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		Stock:       req.Stock,
		ItemType:    req.ItemType,
		IsActive:    true,
	}

	if err := database.DB.Create(&shopItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to create shop item",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse(
		"Shop item created successfully",
		shopItem.ToResponse(),
	))
}

// @Summary Admin: Update shop item
// @Description Admin can update existing shop items
// @Tags Shop
// @Accept json
// @Produce json
// @Param uuid path string true "Shop item UUID"
// @Param request body models.ShopItemUpdateRequest true "Shop item update request"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /admin/shop/items/{uuid} [put]
func (sc *ShopController) UpdateShopItem(c *gin.Context) {
	itemUUID := c.Param("uuid")
	if itemUUID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponseWithCode(
			"Item UUID is required",
			"MISSING_PARAMETER",
			nil,
		))
		return
	}

	var req models.ShopItemUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponseWithCode(
			"Invalid request data",
			"INVALID_REQUEST",
			map[string]interface{}{"error": err.Error()},
		))
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponseWithCode(
			"Validation failed",
			"VALIDATION_ERROR",
			map[string]interface{}{"errors": err},
		))
		return
	}

	// Find shop item
	var shopItem models.ShopItem
	if err := database.DB.Where("uuid = ?", itemUUID).First(&shopItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponseWithCode(
				"Shop item not found",
				"ITEM_NOT_FOUND",
				nil,
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
				"Failed to find shop item",
				"INTERNAL_ERROR",
				nil,
			))
		}
		return
	}

	// Update fields if provided
	if req.Name != "" {
		shopItem.Name = req.Name
	}
	if req.Description != "" {
		shopItem.Description = req.Description
	}
	if req.ImageURL != "" {
		shopItem.ImageURL = req.ImageURL
	}
	if req.Price > 0 {
		shopItem.Price = req.Price
	}
	if req.Stock >= 0 {
		shopItem.Stock = req.Stock
	}
	if req.ItemType != "" {
		shopItem.ItemType = req.ItemType
	}
	if req.IsActive != nil {
		shopItem.IsActive = *req.IsActive
	}

	if err := database.DB.Save(&shopItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to update shop item",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.APIResponse(
		"Shop item updated successfully",
		shopItem.ToResponse(),
	))
}

// @Summary Admin: Delete shop item
// @Description Admin can delete shop items
// @Tags Shop
// @Accept json
// @Produce json
// @Param uuid path string true "Shop item UUID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /admin/shop/items/{uuid} [delete]
func (sc *ShopController) DeleteShopItem(c *gin.Context) {
	itemUUID := c.Param("uuid")
	if itemUUID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponseWithCode(
			"Item UUID is required",
			"MISSING_PARAMETER",
			nil,
		))
		return
	}

	// Find shop item
	var shopItem models.ShopItem
	if err := database.DB.Where("uuid = ?", itemUUID).First(&shopItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponseWithCode(
				"Shop item not found",
				"ITEM_NOT_FOUND",
				nil,
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
				"Failed to find shop item",
				"INTERNAL_ERROR",
				nil,
			))
		}
		return
	}

	// Soft delete
	if err := database.DB.Delete(&shopItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to delete shop item",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.APIResponse(
		"Shop item deleted successfully",
		nil,
	))
}

// @Summary Get shop items
// @Description Get all available shop items
// @Tags Shop
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param item_type query string false "Filter by item type"
// @Param active_only query bool false "Show only active items" default(true)
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /shop/items [get]
func (sc *ShopController) GetShopItems(c *gin.Context) {
	// Pagination
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	// Filters
	itemType := c.Query("item_type")
	activeOnly := c.Query("active_only") != "false" // Default to true

	query := database.DB.Model(&models.ShopItem{})

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if itemType != "" {
		query = query.Where("item_type = ?", itemType)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to count shop items",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Get items
	var shopItems []models.ShopItem
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&shopItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to fetch shop items",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Convert to response
	var itemResponses []models.ShopItemResponse
	for _, item := range shopItems {
		itemResponses = append(itemResponses, item.ToResponse())
	}

	c.JSON(http.StatusOK, models.APIResponse(
		"Shop items retrieved successfully",
		map[string]interface{}{
			"items": itemResponses,
			"pagination": map[string]interface{}{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	))
}

// @Summary Get shop item by UUID
// @Description Get details of a specific shop item
// @Tags Shop
// @Accept json
// @Produce json
// @Param uuid path string true "Shop item UUID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /shop/items/{uuid} [get]
func (sc *ShopController) GetShopItem(c *gin.Context) {
	itemUUID := c.Param("uuid")
	if itemUUID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponseWithCode(
			"Item UUID is required",
			"MISSING_PARAMETER",
			nil,
		))
		return
	}

	var shopItem models.ShopItem
	if err := database.DB.Where("uuid = ?", itemUUID).First(&shopItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponseWithCode(
				"Shop item not found",
				"ITEM_NOT_FOUND",
				nil,
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
				"Failed to find shop item",
				"INTERNAL_ERROR",
				nil,
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.APIResponse(
		"Shop item retrieved successfully",
		shopItem.ToResponse(),
	))
}

// @Summary Purchase shop item
// @Description User can purchase items from the shop
// @Tags Shop
// @Accept json
// @Produce json
// @Param request body models.PurchaseItemRequest true "Purchase request"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /shop/purchase [post]
func (sc *ShopController) PurchaseItem(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponseWithCode(
			"Authentication required",
			"UNAUTHORIZED",
			nil,
		))
		return
	}

	var req models.PurchaseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponseWithCode(
			"Invalid request data",
			"INVALID_REQUEST",
			map[string]interface{}{"error": err.Error()},
		))
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponseWithCode(
			"Validation failed",
			"VALIDATION_ERROR",
			map[string]interface{}{"errors": err},
		))
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find shop item
	var shopItem models.ShopItem
	if err := tx.Where("id = ? AND is_active = ?", req.ShopItemID, true).First(&shopItem).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponseWithCode(
				"Shop item not found or inactive",
				"ITEM_NOT_FOUND",
				nil,
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
				"Failed to find shop item",
				"INTERNAL_ERROR",
				nil,
			))
		}
		return
	}

	// Check stock
	if shopItem.Stock < req.Quantity {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, models.APIResponseWithCode(
			"Insufficient stock",
			"INSUFFICIENT_STOCK",
			map[string]interface{}{
				"available_stock": shopItem.Stock,
				"requested_quantity": req.Quantity,
			},
		))
		return
	}

	// Calculate total cost
	totalCost := shopItem.Price * int64(req.Quantity)

	// Find user
	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to find user",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Check user balance
	if user.Currency < totalCost {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, models.APIResponseWithCode(
			"Insufficient balance",
			"INSUFFICIENT_BALANCE",
			map[string]interface{}{
				"current_balance": user.Currency,
				"required_amount": totalCost,
			},
		))
		return
	}

	// Update user balance
	user.Currency -= totalCost
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to update user balance",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Update shop item stock
	shopItem.Stock -= req.Quantity
	if err := tx.Save(&shopItem).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to update item stock",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Check if user already owns this item
	var existingUserItem models.UserItem
	err := tx.Where("user_id = ? AND shop_item_id = ?", userID, req.ShopItemID).First(&existingUserItem).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new user item
		userItem := models.UserItem{
			UserID:     userID,
			ShopItemID: req.ShopItemID,
			Quantity:   req.Quantity,
		}
		if err := tx.Create(&userItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
				"Failed to create user item",
				"INTERNAL_ERROR",
				nil,
			))
			return
		}
	} else if err == nil {
		// Update existing user item quantity
		existingUserItem.Quantity += req.Quantity
		if err := tx.Save(&existingUserItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
				"Failed to update user item",
				"INTERNAL_ERROR",
				nil,
			))
			return
		}
	} else {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to check user items",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Create transaction log
	transactionLog := models.TransactionLog{
		UserID:      userID,
		Type:        "purchase",
		Amount:      -totalCost, // Negative for purchase
		Description: "Purchased " + strconv.Itoa(req.Quantity) + "x " + shopItem.Name,
	}

	if err := tx.Create(&transactionLog).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to create transaction log",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, models.APIResponse(
		"Item purchased successfully",
		map[string]interface{}{
			"item_name":      shopItem.Name,
			"quantity":       req.Quantity,
			"total_cost":     totalCost,
			"new_balance":    user.Currency,
			"remaining_stock": shopItem.Stock,
		},
	))
}

// @Summary Get user's owned items
// @Description Get all items owned by the authenticated user
// @Tags Shop
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param item_type query string false "Filter by item type"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /shop/my-items [get]
func (sc *ShopController) GetUserItems(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponseWithCode(
			"Authentication required",
			"UNAUTHORIZED",
			nil,
		))
		return
	}

	// Pagination
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	// Filters
	itemType := c.Query("item_type")

	query := database.DB.Where("user_id = ?", userID).
		Preload("ShopItem")

	if itemType != "" {
		query = query.Joins("JOIN mm_shop_items ON mm_user_items.shop_item_id = mm_shop_items.id").
			Where("mm_shop_items.item_type = ?", itemType)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.UserItem{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to count user items",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Get items
	var userItems []models.UserItem
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&userItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponseWithCode(
			"Failed to fetch user items",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Convert to response
	var itemResponses []models.UserItemResponse
	for _, item := range userItems {
		itemResponses = append(itemResponses, item.ToResponse())
	}

	c.JSON(http.StatusOK, models.APIResponse(
		"User items retrieved successfully",
		map[string]interface{}{
			"items": itemResponses,
			"pagination": map[string]interface{}{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	))
}