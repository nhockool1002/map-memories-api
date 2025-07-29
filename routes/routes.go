package routes

import (
	"net/http"

	"map-memories-api/controllers"
	"map-memories-api/database"
	"map-memories-api/middleware"
	"map-memories-api/models"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine) {
	// Initialize controllers
	authController := &controllers.AuthController{}
	memoryController := &controllers.MemoryController{}
	locationController := &controllers.LocationController{}
	mediaController := &controllers.MediaController{}

	// CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		// Check database connection
		if err := database.Health(); err != nil {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponseWithCode(
				"Service unhealthy",
				"SERVICE_UNAVAILABLE",
				map[string]interface{}{"database": "unhealthy"},
			))
			return
		}

		c.JSON(http.StatusOK, models.SuccessResponse(
			"Service is healthy",
			map[string]interface{}{
				"status":   "healthy",
				"database": "connected",
			},
		))
	})

	// API version 1
	v1 := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := v1.Group("/")
		{
			// Authentication routes
			auth := public.Group("auth")
			{
				auth.POST("/register", authController.Register)
				auth.POST("/login", authController.Login)
			}

			// Public locations (read-only)
			locations := public.Group("locations")
			{
				locations.GET("", locationController.GetLocations)
				locations.GET("/:uuid", locationController.GetLocation)
				locations.GET("/nearby", locationController.SearchNearbyLocations)
				locations.GET("/:uuid/memories", locationController.GetLocationMemories)
			}

			// Public memories (read-only, public memories only)
			memories := public.Group("memories")
			{
				memories.GET("", memoryController.GetMemories) // Will filter public memories
				memories.GET("/:uuid", memoryController.GetMemory)
			}

			// Public media (serve files)
			media := public.Group("media")
			{
				media.GET("/:uuid/file", mediaController.ServeMediaFile)
			}
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Authentication profile routes
			auth := protected.Group("auth")
			{
				auth.GET("/profile", authController.GetProfile)
				auth.PUT("/profile", authController.UpdateProfile)
				auth.POST("/logout", authController.Logout)
			}

			// Memory management
			memories := protected.Group("memories")
			{
				memories.POST("", memoryController.CreateMemory)
				memories.PUT("/:uuid", memoryController.UpdateMemory)
				memories.DELETE("/:uuid", memoryController.DeleteMemory)
				memories.GET("/:memory_uuid/media", mediaController.GetMemoryMedia)
			}

			// Location management
			locations := protected.Group("locations")
			{
				locations.POST("", locationController.CreateLocation)
				locations.PUT("/:uuid", locationController.UpdateLocation)
				// Delete is admin only - will be added below
			}

			// Media management
			media := protected.Group("media")
			{
				media.POST("/upload", mediaController.UploadMedia)
				media.GET("", mediaController.GetMedia)
				media.GET("/:uuid", mediaController.GetMediaFile)
				media.PUT("/:uuid", mediaController.UpdateMedia)
				media.DELETE("/:uuid", mediaController.DeleteMedia)
			}
		}

		// Admin routes (admin access required)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminMiddleware())
		{
			// Admin location management
			locations := admin.Group("locations")
			{
				locations.DELETE("/:uuid", locationController.DeleteLocation)
			}

			// Admin can access all memories
			memories := admin.Group("memories")
			{
				memories.GET("", memoryController.GetMemories) // All memories
			}

			// Admin media management
			media := admin.Group("media")
			{
				media.GET("", mediaController.GetMedia) // All media
			}
		}
	}

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Root redirect to swagger
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/swagger/index.html")
	})

	// Handle 404
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, models.ErrorResponseWithCode(
			"Endpoint not found",
			"NOT_FOUND",
			nil,
		))
	})

	// Handle 405 (Method Not Allowed)
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, models.ErrorResponseWithCode(
			"Method not allowed",
			"METHOD_NOT_ALLOWED",
			nil,
		))
	})
}