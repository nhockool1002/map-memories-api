package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ShopItem represents items available in the shop
type ShopItem struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UUID        uuid.UUID      `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	Name        string         `json:"name" gorm:"not null" validate:"required,max=255"`
	Description string         `json:"description" gorm:"type:text"`
	ImageBase64 string         `json:"image_base64" gorm:"type:text;not null"`
	ImageURL    string         `json:"image_url" gorm:"-"` // Virtual field for backward compatibility
	Price       int64          `json:"price" gorm:"not null" validate:"required,min=0"`
	Stock       int            `json:"stock" gorm:"not null;default:0" validate:"min=0"`
	ItemType    string         `json:"item_type" gorm:"size:50;not null;default:'marker'" validate:"required"`
	IsActive    bool           `json:"is_active" gorm:"default:true;not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	UserItems []UserItem `json:"user_items,omitempty" gorm:"foreignKey:ShopItemID"`
}

func (ShopItem) TableName() string {
	return "mm_shop_items"
}

// UserItem represents items owned by users
type UserItem struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UUID       uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	UserID     uint      `json:"user_id" gorm:"not null"`
	ShopItemID uint      `json:"shop_item_id" gorm:"not null"`
	Quantity   int       `json:"quantity" gorm:"not null;default:1" validate:"min=1"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	User     User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ShopItem ShopItem `json:"shop_item,omitempty" gorm:"foreignKey:ShopItemID"`
}

func (UserItem) TableName() string {
	return "mm_user_items"
}

// TransactionLog represents all financial transactions
type TransactionLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UUID        uuid.UUID `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	AdminID     *uint     `json:"admin_id"` // Admin who performed the action (null for user actions)
	Type        string    `json:"type" gorm:"size:50;not null"` // "purchase", "admin_add", "admin_subtract", "admin_transfer"
	Amount      int64     `json:"amount" gorm:"not null"`       // Positive for add, negative for subtract
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Admin *User `json:"admin,omitempty" gorm:"foreignKey:AdminID"`
}

func (TransactionLog) TableName() string {
	return "mm_transaction_logs"
}

// ShopItemCreateRequest represents the request for creating a shop item
type ShopItemCreateRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	Description string `json:"description"`
	ImageBase64 string `json:"image_base64" validate:"required"`
	Price       int64  `json:"price" validate:"required,min=0"`
	Stock       int    `json:"stock" validate:"min=0"`
	ItemType    string `json:"item_type" validate:"required"`
}

// ShopItemUpdateRequest represents the request for updating a shop item
type ShopItemUpdateRequest struct {
	Name        string `json:"name" validate:"max=255"`
	Description string `json:"description"`
	ImageBase64 string `json:"image_base64"`
	Price       int64  `json:"price" validate:"min=0"`
	Stock       int    `json:"stock" validate:"min=0"`
	ItemType    string `json:"item_type"`
	IsActive    *bool  `json:"is_active"`
}

// ShopItemResponse represents the shop item response
type ShopItemResponse struct {
	ID          uint      `json:"id"`
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageBase64 string    `json:"image_base64"`
	Price       int64     `json:"price"`
	Stock       int       `json:"stock"`
	ItemType    string    `json:"item_type"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts ShopItem to ShopItemResponse
func (s *ShopItem) ToResponse() ShopItemResponse {
	// Set ImageURL for backward compatibility
	s.ImageURL = s.ImageBase64
	
	return ShopItemResponse{
		ID:          s.ID,
		UUID:        s.UUID,
		Name:        s.Name,
		Description: s.Description,
		ImageBase64: s.ImageBase64,
		Price:       s.Price,
		Stock:       s.Stock,
		ItemType:    s.ItemType,
		IsActive:    s.IsActive,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

// UserItemResponse represents the user item response
type UserItemResponse struct {
	ID       uint             `json:"id"`
	UUID     uuid.UUID        `json:"uuid"`
	Quantity int              `json:"quantity"`
	ShopItem ShopItemResponse `json:"shop_item"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// ToResponse converts UserItem to UserItemResponse
func (u *UserItem) ToResponse() UserItemResponse {
	return UserItemResponse{
		ID:        u.ID,
		UUID:      u.UUID,
		Quantity:  u.Quantity,
		ShopItem:  u.ShopItem.ToResponse(),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// TransactionLogResponse represents the transaction log response
type TransactionLogResponse struct {
	ID          uint      `json:"id"`
	UUID        uuid.UUID `json:"uuid"`
	UserID      uint      `json:"user_id"`
	AdminID     *uint     `json:"admin_id"`
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	User        UserResponse `json:"user,omitempty"`
	Admin       *UserResponse `json:"admin,omitempty"`
}

// ToResponse converts TransactionLog to TransactionLogResponse
func (t *TransactionLog) ToResponse() TransactionLogResponse {
	var admin *UserResponse
	if t.Admin != nil {
		adminResp := t.Admin.ToResponse()
		admin = &adminResp
	}

	return TransactionLogResponse{
		ID:          t.ID,
		UUID:        t.UUID,
		UserID:      t.UserID,
		AdminID:     t.AdminID,
		Type:        t.Type,
		Amount:      t.Amount,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		User:        t.User.ToResponse(),
		Admin:       admin,
	}
}

// CurrencyUpdateRequest represents the request for admin currency operations
type CurrencyUpdateRequest struct {
	UserID      uint   `json:"user_id" validate:"required"`
	Amount      int64  `json:"amount" validate:"required"`
	Description string `json:"description"`
}

// PurchaseItemRequest represents the request for purchasing an item
type PurchaseItemRequest struct {
	ShopItemID uint `json:"shop_item_id" validate:"required"`
	Quantity   int  `json:"quantity" validate:"required,min=1"`
}