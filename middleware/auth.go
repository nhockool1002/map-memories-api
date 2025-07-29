package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"map-memories-api/models"
	"map-memories-api/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
				"Authorization header is required",
				"UNAUTHORIZED",
				nil,
			))
			c.Abort()
			return
		}

		// Handle both "Bearer <token>" and just "<token>" formats
		var token string
		var err error
		
		// Try standard Bearer format first
		token, err = utils.ExtractBearerToken(authHeader)
		if err != nil {
			// If standard format fails, try treating the entire header as token
			if strings.TrimSpace(authHeader) != "" {
				token = strings.TrimSpace(authHeader)
			} else {
				c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
					"Invalid authorization header format",
					"UNAUTHORIZED",
					nil,
				))
				c.Abort()
				return
			}
		}

		claims, err := utils.VerifyJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
				"Invalid or expired token",
				"UNAUTHORIZED",
				err.Error(),
			))
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_uuid", claims.UserUUID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)
		c.Set("claims", claims)

		c.Next()
	})
}

// OptionalAuthMiddleware validates JWT tokens if present but doesn't require them
func OptionalAuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		
		if authHeader != "" {
			token, err := utils.ExtractBearerToken(authHeader)
			if err == nil {
				claims, err := utils.VerifyJWT(token)
				if err == nil {
					// Set user information in context if token is valid
					c.Set("user_id", claims.UserID)
					c.Set("user_uuid", claims.UserUUID)
					c.Set("user_email", claims.Email)
					c.Set("user_username", claims.Username)
					c.Set("claims", claims)
					c.Set("authenticated", true)
				}
			}
		}

		c.Next()
	})
}

// GetCurrentUserID extracts the current user ID from context
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	
	id, ok := userID.(uint)
	return id, ok
}

// GetCurrentUserUUID extracts the current user UUID from context
func GetCurrentUserUUID(c *gin.Context) (string, bool) {
	userUUID, exists := c.Get("user_uuid")
	if !exists {
		return "", false
	}
	
	uuid, ok := userUUID.(string)
	return uuid, ok
}

// GetCurrentUserEmail extracts the current user email from context
func GetCurrentUserEmail(c *gin.Context) (string, bool) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		return "", false
	}
	
	email, ok := userEmail.(string)
	return email, ok
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// RequireOwnership middleware ensures the user can only access their own resources
func RequireOwnership() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
				"Authentication required",
				"UNAUTHORIZED",
				nil,
			))
			c.Abort()
			return
		}

		// Check if the requested resource belongs to the current user
		resourceUserID := c.Param("user_id")
		if resourceUserID != "" {
			// Convert to uint and compare
			resourceID, err := strconv.ParseUint(resourceUserID, 10, 32)
			if err != nil || uint(resourceID) != userID {
				c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(
					"Access denied: You can only access your own resources",
					"FORBIDDEN",
					nil,
				))
				c.Abort()
				return
			}
		}

		c.Next()
	})
}

// AdminMiddleware ensures only admin users can access the endpoint
func AdminMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// For now, we'll implement a simple admin check
		// In a real application, you'd check user roles from the database
		userEmail, exists := GetCurrentUserEmail(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseWithCode(
				"Authentication required",
				"UNAUTHORIZED",
				nil,
			))
			c.Abort()
			return
		}

		// Simple admin check - in production, use proper role-based access control
		adminEmails := []string{"admin@mapmemories.com", "administrator@mapmemories.com"}
		isAdmin := false
		for _, adminEmail := range adminEmails {
			if strings.EqualFold(userEmail, adminEmail) {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, models.ErrorResponseWithCode(
				"Admin access required",
				"FORBIDDEN",
				nil,
			))
			c.Abort()
			return
		}

		c.Next()
	})
}