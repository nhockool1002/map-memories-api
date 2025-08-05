package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Location struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UUID         uuid.UUID      `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	Name         string         `json:"name" gorm:"not null" validate:"required,max=255"`
	Description  string         `json:"description" gorm:"type:text"`
	Latitude     float64        `json:"latitude" gorm:"type:decimal(10,8);not null" validate:"required,min=-90,max=90"`
	Longitude    float64        `json:"longitude" gorm:"type:decimal(11,8);not null" validate:"required,min=-180,max=180"`
	Address      string         `json:"address" gorm:"type:text"`
	Country      string         `json:"country" gorm:"size:100"`
	City         string         `json:"city" gorm:"size:100"`
	UserID       uint           `json:"user_id" gorm:"not null"`
	MarkerItemID *uint          `json:"marker_item_id"` // NULL for default marker, ShopItem ID for custom marker
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Memories   []Memory  `json:"memories,omitempty" gorm:"foreignKey:LocationID"`
	User       User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	MarkerItem *ShopItem `json:"marker_item,omitempty" gorm:"foreignKey:MarkerItemID"`
}

func (Location) TableName() string {
	return "mm_locations"
}

// LocationCreateRequest represents the request for creating a location
type LocationCreateRequest struct {
	Name         string  `json:"name" validate:"required,max=255"`
	Description  string  `json:"description"`
	Latitude     float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude    float64 `json:"longitude" validate:"required,min=-180,max=180"`
	Address      string  `json:"address"`
	Country      string  `json:"country"`
	City         string  `json:"city"`
	MarkerItemID *uint   `json:"marker_item_id"` // Optional custom marker
}

// LocationUpdateRequest represents the request for updating a location
type LocationUpdateRequest struct {
	Name         string  `json:"name" validate:"max=255"`
	Description  string  `json:"description"`
	Latitude     float64 `json:"latitude" validate:"min=-90,max=90"`
	Longitude    float64 `json:"longitude" validate:"min=-180,max=180"`
	Address      string  `json:"address"`
	Country      string  `json:"country"`
	City         string  `json:"city"`
	MarkerItemID *uint   `json:"marker_item_id"`
}

// LocationResponse represents the location response with memory count
type LocationResponse struct {
	ID           uint              `json:"id"`
	UUID         uuid.UUID         `json:"uuid"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Latitude     float64           `json:"latitude"`
	Longitude    float64           `json:"longitude"`
	Address      string            `json:"address"`
	Country      string            `json:"country"`
	City         string            `json:"city"`
	UserID       uint              `json:"user_id"`
	MarkerItemID *uint             `json:"marker_item_id"`
	MarkerItem   *ShopItemResponse `json:"marker_item,omitempty"`
	MemoryCount  int64             `json:"memory_count"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// ToResponse converts Location to LocationResponse
func (l *Location) ToResponse() LocationResponse {
	var markerItem *ShopItemResponse
	if l.MarkerItem != nil {
		markerResp := l.MarkerItem.ToResponse()
		markerItem = &markerResp
	}

	return LocationResponse{
		ID:           l.ID,
		UUID:         l.UUID,
		Name:         l.Name,
		Description:  l.Description,
		Latitude:     l.Latitude,
		Longitude:    l.Longitude,
		Address:      l.Address,
		Country:      l.Country,
		City:         l.City,
		UserID:       l.UserID,
		MarkerItemID: l.MarkerItemID,
		MarkerItem:   markerItem,
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,
	}
}

// LocationSearchRequest represents the request for searching locations near a point
type LocationSearchRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	Radius    float64 `json:"radius" validate:"min=0,max=100"` // in kilometers
	Limit     int     `json:"limit" validate:"min=1,max=100"`
}
