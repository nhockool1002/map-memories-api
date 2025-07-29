package models

import (
	"time"

	"github.com/google/uuid"
)

type Location struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UUID        uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	Name        string    `json:"name" gorm:"not null" validate:"required,max=255"`
	Description string    `json:"description" gorm:"type:text"`
	Latitude    float64   `json:"latitude" gorm:"type:decimal(10,8);not null" validate:"required,min=-90,max=90"`
	Longitude   float64   `json:"longitude" gorm:"type:decimal(11,8);not null" validate:"required,min=-180,max=180"`
	Address     string    `json:"address" gorm:"type:text"`
	Country     string    `json:"country" gorm:"size:100"`
	City        string    `json:"city" gorm:"size:100"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Memories []Memory `json:"memories,omitempty" gorm:"foreignKey:LocationID"`
}

func (Location) TableName() string {
	return "mm_locations"
}

// LocationCreateRequest represents the request for creating a location
type LocationCreateRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude   float64 `json:"longitude" validate:"required,min=-180,max=180"`
	Address     string  `json:"address"`
	Country     string  `json:"country"`
	City        string  `json:"city"`
}

// LocationUpdateRequest represents the request for updating a location
type LocationUpdateRequest struct {
	Name        string  `json:"name" validate:"max=255"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude" validate:"min=-90,max=90"`
	Longitude   float64 `json:"longitude" validate:"min=-180,max=180"`
	Address     string  `json:"address"`
	Country     string  `json:"country"`
	City        string  `json:"city"`
}

// LocationResponse represents the location response with memory count
type LocationResponse struct {
	ID          uint      `json:"id"`
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Address     string    `json:"address"`
	Country     string    `json:"country"`
	City        string    `json:"city"`
	MemoryCount int64     `json:"memory_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts Location to LocationResponse
func (l *Location) ToResponse() LocationResponse {
	return LocationResponse{
		ID:          l.ID,
		UUID:        l.UUID,
		Name:        l.Name,
		Description: l.Description,
		Latitude:    l.Latitude,
		Longitude:   l.Longitude,
		Address:     l.Address,
		Country:     l.Country,
		City:        l.City,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}

// LocationSearchRequest represents the request for searching locations near a point
type LocationSearchRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	Radius    float64 `json:"radius" validate:"min=0,max=100"` // in kilometers
	Limit     int     `json:"limit" validate:"min=1,max=100"`
}