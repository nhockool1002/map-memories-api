package models

import (
	"time"
)

// UserSession represents user authentication sessions
type UserSession struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	TokenHash string    `json:"-" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (UserSession) TableName() string {
	return "mm_user_sessions"
}

// MemoryLike represents user likes on memories
type MemoryLike struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	UserID   uint      `json:"user_id" gorm:"not null"`
	MemoryID uint      `json:"memory_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Memory Memory `json:"memory,omitempty" gorm:"foreignKey:MemoryID"`
}

func (MemoryLike) TableName() string {
	return "mm_memory_likes"
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User        UserResponse `json:"user"`
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"` // seconds
}

// RefreshTokenRequest represents the request for refreshing tokens
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}