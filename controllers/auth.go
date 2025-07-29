package controllers

import (
	"net/http"
	"time"

	"map-memories-api/database"
	"map-memories-api/middleware"
	"map-memories-api/models"
	"map-memories-api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct{}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with username, email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body models.UserRegistrationRequest true "User registration data"
// @Success 201 {object} models.APIResponse{data=models.AuthResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var req models.UserRegistrationRequest
	
	if err := utils.ValidateAndBindJSON(c, &req); err != nil {
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponseWithCode(
			"User already exists with this email or username",
			"USER_EXISTS",
			nil,
		))
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to process password",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Create new user
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to create user",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to generate authentication token",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Create response
	authResponse := models.AuthResponse{
		User:        user.ToResponse(),
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Hour * 24 / time.Second), // 24 hours in seconds
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(
		"User registered successfully",
		authResponse,
	))
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body models.UserLoginRequest true "User login credentials"
// @Success 200 {object} models.APIResponse{data=models.AuthResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var req models.UserLoginRequest
	
	if err := utils.ValidateAndBindJSON(c, &req); err != nil {
		return
	}

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
				"Invalid email or password",
				"INVALID_CREDENTIALS",
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

	// Verify password
	if err := utils.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			"Invalid email or password",
			"INVALID_CREDENTIALS",
			nil,
		))
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to generate authentication token",
			"INTERNAL_ERROR",
			nil,
		))
		return
	}

	// Create response
	authResponse := models.AuthResponse{
		User:        user.ToResponse(),
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Hour * 24 / time.Second), // 24 hours in seconds
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Login successful",
		authResponse,
	))
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current authenticated user's profile
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.UserResponse}
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/profile [get]
func (ac *AuthController) GetProfile(c *gin.Context) {
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
		"Profile retrieved successfully",
		user.ToResponse(),
	))
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update current authenticated user's profile
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body models.UserUpdateRequest true "User profile update data"
// @Success 200 {object} models.APIResponse{data=models.UserResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/profile [put]
func (ac *AuthController) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
			"Authentication required",
			"UNAUTHORIZED",
			nil,
		))
		return
	}

	var req models.UserUpdateRequest
	if err := utils.ValidateAndBindJSON(c, &req); err != nil {
		return
	}

	// Find user
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
				"User not found",
				"USER_NOT_FOUND",
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

	// Update user fields
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	// Save changes
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseWithCode(
			"Failed to update profile",
			"INTERNAL_ERROR",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Profile updated successfully",
		user.ToResponse(),
	))
}

// Logout godoc
// @Summary Logout user
// @Description Logout current authenticated user
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	// In a real application, you would invalidate the token
	// For now, we'll just return a success response
	// You could add token blacklisting here
	
	c.JSON(http.StatusOK, models.SuccessResponse(
		"Logout successful",
		nil,
	))
}