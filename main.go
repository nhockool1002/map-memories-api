package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"map-memories-api/config"
	"map-memories-api/database"
	"map-memories-api/routes"
	_ "map-memories-api/docs"

	"github.com/gin-gonic/gin"
)

// @title Map Memories API
// @version 1.0
// @description API để quản lý kỷ niệm và địa điểm trên bản đồ
// @description 
// @description Ứng dụng Map Memories cho phép người dùng:
// @description - Đăng ký và xác thực tài khoản
// @description - Tạo và quản lý địa điểm trên bản đồ
// @description - Viết và chia sẻ kỷ niệm tại các địa điểm
// @description - Upload hình ảnh và video cho kỷ niệm
// @description - Tìm kiếm địa điểm và kỷ niệm gần đó
// @description 
// @description ## Xác thực
// @description API sử dụng JWT Bearer token để xác thực. Để truy cập các endpoint được bảo vệ:
// @description 1. Đăng ký tài khoản hoặc đăng nhập để nhận token
// @description 2. Thêm token vào header: `Authorization: Bearer <your-token>`
// @description 
// @description ## Rate Limiting
// @description API có giới hạn 60 requests/phút cho mỗi IP
// @description 
// @description ## File Upload
// @description - Hỗ trợ upload hình ảnh: JPEG, PNG, GIF
// @description - Hỗ trợ upload video: MP4, AVI, MOV
// @description - Kích thước file tối đa: 50MB
// @description 
// @description ## Geospatial Features
// @description - Tìm kiếm địa điểm trong bán kính từ tọa độ
// @description - Sử dụng PostGIS để tính toán khoảng cách chính xác
// @description - Hỗ trợ tọa độ WGS84 (EPSG:4326)

// @contact.name Map Memories API Support
// @contact.email support@mapmemories.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8222
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name Authentication
// @tag.description Xác thực và quản lý tài khoản người dùng

// @tag.name Memories
// @tag.description Quản lý kỷ niệm và bài viết

// @tag.name Locations
// @tag.description Quản lý địa điểm và tìm kiếm geospatial

// @tag.name Media
// @tag.description Upload và quản lý hình ảnh, video

func main() {
	// Load configuration
	config.LoadConfig()

	// Set Gin mode
	if config.AppConfig.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	database.Connect()
	defer database.Close()

	// Run database migrations
	database.AutoMigrate()

	// Run database seeding
	database.SeedData()

	// Create Gin router
	r := gin.New()

	// Gin middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Trust proxy (important for getting real client IP behind load balancer)
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// Limit request size
	r.MaxMultipartMemory = config.AppConfig.Upload.MaxFileSizeInt

	// Setup routes
	routes.SetupRoutes(r)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting Map Memories API server on %s", server.Addr)
		log.Printf("Environment: %s", config.AppConfig.Environment)
		log.Printf("Swagger documentation available at: http://%s/swagger/index.html", server.Addr)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server exited gracefully")
	}
}