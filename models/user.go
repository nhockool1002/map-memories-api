package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UUID         uuid.UUID      `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	Username     string         `json:"username" gorm:"uniqueIndex;not null" validate:"required,min=3,max=50"`
	Email        string         `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	PasswordHash string         `json:"-" gorm:"not null"`
	FullName     string         `json:"full_name" gorm:"size:255"`
	AvatarURL    string         `json:"avatar_url" gorm:"type:text"`
	IsAdmin      bool           `json:"is_admin" gorm:"default:false;not null"`
	Currency     int64          `json:"currency" gorm:"default:0;not null"` // Xu currency
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Memories  []Memory    `json:"memories,omitempty" gorm:"foreignKey:UserID"`
	Sessions  []UserSession `json:"-" gorm:"foreignKey:UserID"`
	Likes     []MemoryLike `json:"-" gorm:"foreignKey:UserID"`
	UserItems []UserItem   `json:"user_items,omitempty" gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "mm_users"
}

// UserRegistrationRequest represents the request for user registration
type UserRegistrationRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name"`
}

// UserLoginRequest represents the request for user login
type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserResponse represents the user response (without sensitive data)
type UserResponse struct {
	ID        uint         `json:"id"`
	UUID      uuid.UUID    `json:"uuid"`
	Username  string       `json:"username"`
	Email     string       `json:"email"`
	FullName  string       `json:"full_name"`
	AvatarURL string       `json:"avatar_url"`
	IsAdmin   bool         `json:"is_admin"`
	Currency  int64        `json:"currency"`
	UserItems []UserItemResponse `json:"user_items,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	var userItems []UserItemResponse
	for _, item := range u.UserItems {
		userItems = append(userItems, item.ToResponse())
	}
	
	return UserResponse{
		ID:        u.ID,
		UUID:      u.UUID,
		Username:  u.Username,
		Email:     u.Email,
		FullName:  u.FullName,
		AvatarURL: u.AvatarURL,
		IsAdmin:   u.IsAdmin,
		Currency:  u.Currency,
		UserItems: userItems,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// UserUpdateRequest represents the request for updating user profile
type UserUpdateRequest struct {
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
}