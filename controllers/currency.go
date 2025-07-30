package controllers

import (
	"net/http"
	"strconv"

	"map-memories-api/database"
	"map-memories-api/middleware"
	"map-memories-api/models"
	"map-memories-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CurrencyController struct{}

// @Summary Admin: Add currency to user
// @Description Admin can add currency to any user account
// @Tags Currency
// @Accept json
// @Produce json
// @Param request body models.CurrencyUpdateRequest true "Currency update request"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /admin/currency/add [post]
func (cc *CurrencyController) AdminAddCurrency(c *gin.Context) {
	var req models.CurrencyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid request data",
			"INVALID_REQUEST",
			map[string]interface{}{"error": err.Error()},
		))
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Validation failed",
			"VALIDATION_ERROR",
			map[string]interface{}{"errors": err},
		))
		return
	}

	adminID, _ := middleware.GetCurrentUserID(c)

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find target user
	var user models.User
	if err := tx.First(&user, req.UserID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"User not found",
				"USER_NOT_FOUND",
				nil,
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				"Failed to find user",
				"INTERNAL_ERROR",
				nil,
			))
		}
		return
	}

	// Update user currency
	user.Currency += req.Amount
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to update currency",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Create transaction log
	transactionLog := models.TransactionLog{
		UserID:      req.UserID,
		AdminID:     &adminID,
		Type:        "admin_add",
		Amount:      req.Amount,
		Description: req.Description,
	}

	if err := tx.Create(&transactionLog).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to create transaction log",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Currency added successfully",
		map[string]interface{}{
			"user_id":      user.ID,
			"new_balance":  user.Currency,
			"amount_added": req.Amount,
		},
	))
}

// @Summary Admin: Subtract currency from user
// @Description Admin can subtract currency from any user account
// @Tags Currency
// @Accept json
// @Produce json
// @Param request body models.CurrencyUpdateRequest true "Currency update request"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /admin/currency/subtract [post]
func (cc *CurrencyController) AdminSubtractCurrency(c *gin.Context) {
	var req models.CurrencyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid request data",
			"INVALID_REQUEST",
			map[string]interface{}{"error": err.Error()},
		))
		return
	}

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Validation failed",
			"VALIDATION_ERROR",
			map[string]interface{}{"errors": err},
		))
		return
	}

	adminID, _ := middleware.GetCurrentUserID(c)

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find target user
	var user models.User
	if err := tx.First(&user, req.UserID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"User not found",
				"USER_NOT_FOUND",
				nil,
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				"Failed to find user",
				"INTERNAL_ERROR",
				nil,
			))
		}
		return
	}

	// Check if user has sufficient balance
	if user.Currency < req.Amount {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Insufficient balance",
			"INSUFFICIENT_BALANCE",
			map[string]interface{}{
				"current_balance": user.Currency,
				"required_amount": req.Amount,
			},
		))
		return
	}

	// Update user currency
	user.Currency -= req.Amount
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to update currency",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Create transaction log
	transactionLog := models.TransactionLog{
		UserID:      req.UserID,
		AdminID:     &adminID,
		Type:        "admin_subtract",
		Amount:      -req.Amount, // Negative for subtract
		Description: req.Description,
	}

	if err := tx.Create(&transactionLog).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to create transaction log",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Currency subtracted successfully",
		map[string]interface{}{
			"user_id":           user.ID,
			"new_balance":       user.Currency,
			"amount_subtracted": req.Amount,
		},
	))
}

// @Summary Get transaction history
// @Description Get transaction history for a specific user
// @Tags Currency
// @Accept json
// @Produce json
// @Param user_id query uint true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /admin/currency/history [get]
func (cc *CurrencyController) GetTransactionHistory(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"user_id parameter is required",
			"MISSING_PARAMETER",
			nil,
		))
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCode(
			"Invalid user_id format",
			"INVALID_PARAMETER",
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

	// Get transactions
	var transactions []models.TransactionLog
	var total int64

	query := database.DB.Where("user_id = ?", userID).
		Preload("User").
		Preload("Admin").
		Order("created_at DESC")

	// Count total
	if err := query.Model(&models.TransactionLog{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to count transactions",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to fetch transaction history",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Convert to response
	var transactionResponses []models.TransactionLogResponse
	for _, transaction := range transactions {
		transactionResponses = append(transactionResponses, transaction.ToResponse())
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Transaction history retrieved successfully",
		map[string]interface{}{
			"transactions": transactionResponses,
			"pagination": map[string]interface{}{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	))
}

// @Summary Get user balance
// @Description Get current balance of authenticated user
// @Tags Currency
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /currency/balance [get]
func (cc *CurrencyController) GetBalance(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			"Authentication required",
			"UNAUTHORIZED",
			nil,
		))
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"User not found",
				"USER_NOT_FOUND",
				nil,
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
				"Failed to get user data",
				"INTERNAL_ERROR",
				nil,
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Balance retrieved successfully",
		map[string]interface{}{
			"user_id": user.ID,
			"balance": user.Currency,
		},
	))
}

// @Summary Get my transaction history
// @Description Get transaction history for authenticated user
// @Tags Currency
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /currency/history [get]
func (cc *CurrencyController) GetMyTransactionHistory(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
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

	// Get transactions
	var transactions []models.TransactionLog
	var total int64

	query := database.DB.Where("user_id = ?", userID).
		Preload("User").
		Preload("Admin").
		Order("created_at DESC")

	// Count total
	if err := query.Model(&models.TransactionLog{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to count transactions",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to fetch transaction history",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Convert to response
	var transactionResponses []models.TransactionLogResponse
	for _, transaction := range transactions {
		transactionResponses = append(transactionResponses, transaction.ToResponse())
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Transaction history retrieved successfully",
		map[string]interface{}{
			"transactions": transactionResponses,
			"pagination": map[string]interface{}{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	))
}