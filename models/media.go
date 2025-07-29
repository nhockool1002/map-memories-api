package models

import (
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UUID             uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	MemoryID         uint      `json:"memory_id" gorm:"not null"`
	Filename         string    `json:"filename" gorm:"not null"`
	OriginalFilename string    `json:"original_filename" gorm:"not null"`
	FilePath         string    `json:"file_path" gorm:"type:text;not null"`
	FileSize         int64     `json:"file_size" gorm:"not null"`
	MimeType         string    `json:"mime_type" gorm:"not null"`
	MediaType        string    `json:"media_type" gorm:"not null;check:media_type IN ('image', 'video')"`
	DisplayOrder     int       `json:"display_order" gorm:"default:0"`
	CreatedAt        time.Time `json:"created_at"`

	// Relationships
	Memory Memory `json:"memory,omitempty" gorm:"foreignKey:MemoryID"`
}

func (Media) TableName() string {
	return "mm_media"
}

// MediaResponse represents the media response
type MediaResponse struct {
	ID               uint      `json:"id"`
	UUID             uuid.UUID `json:"uuid"`
	Filename         string    `json:"filename"`
	OriginalFilename string    `json:"original_filename"`
	FilePath         string    `json:"file_path"`
	FileSize         int64     `json:"file_size"`
	MimeType         string    `json:"mime_type"`
	MediaType        string    `json:"media_type"`
	DisplayOrder     int       `json:"display_order"`
	URL              string    `json:"url"` // Full URL for accessing the file
	ThumbnailURL     string    `json:"thumbnail_url,omitempty"` // For images/videos
	CreatedAt        time.Time `json:"created_at"`
}

// ToResponse converts Media to MediaResponse
func (m *Media) ToResponse() MediaResponse {
	return MediaResponse{
		ID:               m.ID,
		UUID:             m.UUID,
		Filename:         m.Filename,
		OriginalFilename: m.OriginalFilename,
		FilePath:         m.FilePath,
		FileSize:         m.FileSize,
		MimeType:         m.MimeType,
		MediaType:        m.MediaType,
		DisplayOrder:     m.DisplayOrder,
		URL:              "/api/v1/media/" + m.UUID.String(),
		CreatedAt:        m.CreatedAt,
	}
}

// MediaUploadRequest represents the request for uploading media
type MediaUploadRequest struct {
	MemoryID     uint `json:"memory_id" validate:"required"`
	DisplayOrder int  `json:"display_order"`
}

// MediaUpdateRequest represents the request for updating media
type MediaUpdateRequest struct {
	DisplayOrder int `json:"display_order"`
}

// MediaListRequest represents the request for listing media
type MediaListRequest struct {
	MemoryID  uint   `json:"memory_id"`
	MediaType string `json:"media_type" validate:"oneof=image video"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Offset    int    `json:"offset" validate:"min=0"`
}