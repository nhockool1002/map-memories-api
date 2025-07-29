package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// DateOnly represents a date in YYYY-MM-DD format
type DateOnly struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	// Remove quotes
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	
	// Parse date in YYYY-MM-DD format
	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return err
	}
	
	d.Time = t
	return nil
}

// MarshalJSON implements json.Marshaler
func (d DateOnly) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Time.Format("2006-01-02") + `"`), nil
}

type Memory struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UUID       uuid.UUID      `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	UserID     uint           `json:"user_id" gorm:"not null"`
	LocationID uint           `json:"location_id" gorm:"not null"`
	Title      string         `json:"title" gorm:"not null" validate:"required,max=255"`
	Content    string         `json:"content" gorm:"type:text;not null" validate:"required"`
	VisitDate  *time.Time     `json:"visit_date"`
	IsPublic   bool           `json:"is_public" gorm:"default:false"`
	Tags       pq.StringArray `json:"tags" gorm:"type:text[]"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	User     User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Location Location `json:"location,omitempty" gorm:"foreignKey:LocationID"`
	Media    []Media `json:"media,omitempty" gorm:"foreignKey:MemoryID"`
	Likes    []MemoryLike `json:"likes,omitempty" gorm:"foreignKey:MemoryID"`
}

func (Memory) TableName() string {
	return "mm_memories"
}

// MemoryCreateRequest represents the request for creating a memory
type MemoryCreateRequest struct {
	LocationID uint      `json:"location_id" validate:"required"`
	Title      string    `json:"title" validate:"required,max=255"`
	Content    string    `json:"content" validate:"required"`
	VisitDate  *DateOnly `json:"visit_date"`
	IsPublic   bool      `json:"is_public"`
	Tags       []string  `json:"tags"`
}

// MemoryUpdateRequest represents the request for updating a memory
type MemoryUpdateRequest struct {
	Title     string     `json:"title" validate:"max=255"`
	Content   string     `json:"content"`
	VisitDate *DateOnly  `json:"visit_date"`
	IsPublic  bool       `json:"is_public"`
	Tags      []string   `json:"tags"`
}

// MemoryResponse represents the memory response with additional data
type MemoryResponse struct {
	ID         uint              `json:"id"`
	UUID       uuid.UUID         `json:"uuid"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	VisitDate  *time.Time        `json:"visit_date"`
	IsPublic   bool              `json:"is_public"`
	Tags       []string          `json:"tags"`
	LikeCount  int64             `json:"like_count"`
	MediaCount int64             `json:"media_count"`
	IsLiked    bool              `json:"is_liked,omitempty"` // for current user
	User       UserResponse      `json:"user"`
	Location   LocationResponse  `json:"location"`
	Media      []MediaResponse   `json:"media,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// ToResponse converts Memory to MemoryResponse
func (m *Memory) ToResponse() MemoryResponse {
	tags := make([]string, 0)
	if m.Tags != nil {
		tags = []string(m.Tags)
	}

	response := MemoryResponse{
		ID:        m.ID,
		UUID:      m.UUID,
		Title:     m.Title,
		Content:   m.Content,
		VisitDate: m.VisitDate,
		IsPublic:  m.IsPublic,
		Tags:      tags,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.User.ID != 0 {
		response.User = m.User.ToResponse()
	}

	if m.Location.ID != 0 {
		response.Location = m.Location.ToResponse()
	}

	if m.Media != nil {
		media := make([]MediaResponse, len(m.Media))
		for i, mediaItem := range m.Media {
			media[i] = mediaItem.ToResponse()
		}
		response.Media = media
	}

	return response
}

// MemoryListRequest represents the request for listing memories
type MemoryListRequest struct {
	UserID     *uint   `json:"user_id"`
	LocationID *uint   `json:"location_id"`
	IsPublic   *bool   `json:"is_public"`
	Tags       []string `json:"tags"`
	Search     string  `json:"search"`
	Limit      int     `json:"limit" validate:"min=1,max=100"`
	Offset     int     `json:"offset" validate:"min=0"`
	SortBy     string  `json:"sort_by" validate:"oneof=created_at visit_date title"`
	SortOrder  string  `json:"sort_order" validate:"oneof=asc desc"`
}

// MemoryNearbyRequest represents the request for finding memories near a location
type MemoryNearbyRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	Radius    float64 `json:"radius" validate:"min=0,max=100"` // in kilometers
	IsPublic  *bool   `json:"is_public"`
	Limit     int     `json:"limit" validate:"min=1,max=100"`
	Offset    int     `json:"offset" validate:"min=0"`
}