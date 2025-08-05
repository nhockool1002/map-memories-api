package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"map-memories-api/config"

	"encoding/base64"
	"mime"

	"github.com/google/uuid"
)

// FileInfo represents uploaded file information
type FileInfo struct {
	OriginalFilename string
	Filename         string
	FilePath         string
	FileSize         int64
	MimeType         string
	MediaType        string
}

// IsAllowedFileType checks if the file type is allowed
func IsAllowedFileType(mimeType string) bool {
	allowedTypes := config.AppConfig.Upload.AllowedTypes
	for _, allowedType := range allowedTypes {
		if strings.EqualFold(mimeType, allowedType) {
			return true
		}
	}
	return false
}

// GetMediaType determines if the file is an image or video
func GetMediaType(mimeType string) string {
	if strings.HasPrefix(mimeType, "image/") {
		return "image"
	} else if strings.HasPrefix(mimeType, "video/") {
		return "video"
	}
	return "unknown"
}

// GenerateUniqueFilename generates a unique filename for uploaded files
func GenerateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Format("20060102_150405")
	uniqueID := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%s%s", timestamp, uniqueID, ext)
}

// SaveUploadedFile saves an uploaded file to the configured upload directory
func SaveUploadedFile(file *multipart.FileHeader) (*FileInfo, error) {
	// Check file size
	if file.Size > config.AppConfig.Upload.MaxFileSizeInt {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %s", config.AppConfig.Upload.MaxFileSize)
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Detect MIME type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read file for MIME type detection: %w", err)
	}

	// Reset file pointer
	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	mimeType := DetectMimeType(buffer, file.Filename)

	// Check if file type is allowed
	if !IsAllowedFileType(mimeType) {
		return nil, fmt.Errorf("file type %s is not allowed", mimeType)
	}

	// Generate unique filename
	filename := GenerateUniqueFilename(file.Filename)

	// Create upload directory if it doesn't exist
	uploadPath := config.AppConfig.Upload.Path
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Full file path
	filePath := filepath.Join(uploadPath, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return &FileInfo{
		OriginalFilename: file.Filename,
		Filename:         filename,
		FilePath:         filePath,
		FileSize:         file.Size,
		MimeType:         mimeType,
		MediaType:        GetMediaType(mimeType),
	}, nil
}

// SaveUploadedFileAsBase64 saves an uploaded file as base64 string
func SaveUploadedFileAsBase64(file *multipart.FileHeader) (*FileInfo, error) {
	// Check file size
	if file.Size > config.AppConfig.Upload.MaxFileSizeInt {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %s", config.AppConfig.Upload.MaxFileSize)
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Detect MIME type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read file for MIME type detection: %w", err)
	}

	// Reset file pointer
	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	mimeType := DetectMimeType(buffer, file.Filename)

	// Check if file type is allowed
	if !IsAllowedFileType(mimeType) {
		return nil, fmt.Errorf("file type %s is not allowed", mimeType)
	}

	// Read file content for base64 conversion
	fileContent, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Convert to base64
	base64Str := base64.StdEncoding.EncodeToString(fileContent)

	// Create data URL format
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str)

	// Generate unique filename for reference
	filename := GenerateUniqueFilename(file.Filename)

	return &FileInfo{
		OriginalFilename: file.Filename,
		Filename:         filename,
		FilePath:         dataURL, // Store base64 data URL instead of file path
		FileSize:         file.Size,
		MimeType:         mimeType,
		MediaType:        GetMediaType(mimeType),
	}, nil
}

// DetectMimeType detects MIME type from file content and extension
func DetectMimeType(buffer []byte, filename string) string {
	// Basic MIME type detection based on file signature
	if len(buffer) >= 4 {
		// JPEG
		if buffer[0] == 0xFF && buffer[1] == 0xD8 && buffer[2] == 0xFF {
			return "image/jpeg"
		}
		// PNG
		if buffer[0] == 0x89 && buffer[1] == 0x50 && buffer[2] == 0x4E && buffer[3] == 0x47 {
			return "image/png"
		}
		// GIF
		if string(buffer[0:3]) == "GIF" {
			return "image/gif"
		}
		// MP4
		if len(buffer) >= 8 && string(buffer[4:8]) == "ftyp" {
			return "video/mp4"
		}
	}

	// Fallback to extension-based detection
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/avi"
	case ".mov":
		return "video/mov"
	default:
		return "application/octet-stream"
	}
}

// DeleteFile deletes a file from the filesystem
func DeleteFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}
	return os.Remove(filePath)
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetFileSize returns the size of a file
func GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// ValidateImageFile validates if the uploaded file is a valid image
func ValidateImageFile(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > config.AppConfig.Upload.MaxFileSizeInt {
		return errors.New("file size exceeds maximum allowed size")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif"}

	isValidExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isValidExt = true
			break
		}
	}

	if !isValidExt {
		return fmt.Errorf("invalid file extension: %s", ext)
	}

	return nil
}

// ValidateVideoFile validates if the uploaded file is a valid video
func ValidateVideoFile(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > config.AppConfig.Upload.MaxFileSizeInt {
		return errors.New("file size exceeds maximum allowed size")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".mp4", ".avi", ".mov"}

	isValidExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isValidExt = true
			break
		}
	}

	if !isValidExt {
		return fmt.Errorf("invalid file extension: %s", ext)
	}

	return nil
}

// FileToBase64 converts a file to base64 string with data URL format
func FileToBase64(filePath string) (string, error) {
	// Read file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Get MIME type
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		// Default to image/png if MIME type not found
		mimeType = "image/png"
	}

	// Convert to base64
	base64Str := base64.StdEncoding.EncodeToString(content)

	// Return as data URL
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str), nil
}

// Base64ToDataURL converts base64 string to data URL format
func Base64ToDataURL(base64Str, mimeType string) string {
	if mimeType == "" {
		mimeType = "image/png"
	}
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str)
}

// ImageToBase64 converts an image file to base64 string
func ImageToBase64(filePath string) (string, error) {
	// Read the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read file content
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Get MIME type
	ext := strings.ToLower(filepath.Ext(filePath))
	var mimeType string
	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	default:
		mimeType = mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "image/png" // Default fallback
		}
	}

	// Encode to base64
	base64Str := base64.StdEncoding.EncodeToString(bytes)

	// Return with data URL format
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str), nil
}

// GetMarkerBase64 returns base64 string for marker images
func GetMarkerBase64(markerNumber int) (string, error) {
	markerPath := fmt.Sprintf("media/markers/marker%d.png", markerNumber)
	return ImageToBase64(markerPath)
}
