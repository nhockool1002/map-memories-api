package middleware

import (
	"strings"

	"map-memories-api/config"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		
		// Check if origin is allowed
		if isOriginAllowed(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(config.AppConfig.CORS.AllowedOrigins) == 1 && config.AppConfig.CORS.AllowedOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(config.AppConfig.CORS.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AppConfig.CORS.AllowedHeaders, ", "))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// isOriginAllowed checks if the origin is in the allowed origins list
func isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	allowedOrigins := config.AppConfig.CORS.AllowedOrigins
	
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		
		// Support wildcard subdomains
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}
	
	return false
}