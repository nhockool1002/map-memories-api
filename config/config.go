package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Environment
	Environment string

	// Database
	Database DatabaseConfig

	// JWT
	JWT JWTConfig

	// Server
	Server ServerConfig

	// Upload
	Upload UploadConfig

	// Redis
	Redis RedisConfig

	// CORS
	CORS CORSConfig

	// Rate Limiting
	RateLimit RateLimitConfig

	// Pagination
	Pagination PaginationConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type ServerConfig struct {
	Port string
	Host string
}

type UploadConfig struct {
	Path           string
	MaxFileSize    string
	AllowedTypes   []string
	MaxFileSizeInt int64
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type RateLimitConfig struct {
	PerMinute int
	Burst     int
}

type PaginationConfig struct {
	DefaultPageSize int
	MaxPageSize     int
}

var AppConfig *Config

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		Environment: getEnv("ENV", "development"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "mm_user"),
			Password: getEnv("DB_PASSWORD", "mm_password"),
			Name:     getEnv("DB_NAME", "map_memories"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			Expiry: getEnvAsDuration("JWT_EXPIRY", 24*time.Hour),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		Upload: UploadConfig{
			Path:         getEnv("UPLOAD_PATH", "./uploads"),
			MaxFileSize:  getEnv("MAX_FILE_SIZE", "50MB"),
			AllowedTypes: getEnvAsSlice("ALLOWED_FILE_TYPES", []string{"image/jpeg", "image/png", "image/gif", "video/mp4"}),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods: getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}),
		},
		RateLimit: RateLimitConfig{
			PerMinute: getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
			Burst:     getEnvAsInt("RATE_LIMIT_BURST", 10),
		},
		Pagination: PaginationConfig{
			DefaultPageSize: getEnvAsInt("DEFAULT_PAGE_SIZE", 20),
			MaxPageSize:     getEnvAsInt("MAX_PAGE_SIZE", 100),
		},
	}

	// Parse max file size
	config.Upload.MaxFileSizeInt = parseFileSize(config.Upload.MaxFileSize)

	AppConfig = config
	return config
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		c.Database.Host,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.Port,
		c.Database.SSLMode,
	)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}

func parseFileSize(sizeStr string) int64 {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	
	var multiplier int64 = 1
	if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	}

	if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
		return size * multiplier
	}

	return 50 * 1024 * 1024 // Default 50MB
}