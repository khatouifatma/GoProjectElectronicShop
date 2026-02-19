package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ========================
// SHOP MODEL
// ========================

type Shop struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name           string    `gorm:"not null" json:"name"`
	Active         bool      `gorm:"default:true" json:"active"`
	WhatsAppNumber string    `gorm:"not null" json:"whatsapp_number"`
	CreatedAt      time.Time `json:"created_at"`
	Users          []User    `gorm:"foreignKey:ShopID" json:"-"`
	Products       []Product `gorm:"foreignKey:ShopID" json:"-"`
}

func (s *Shop) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()
	return nil
}

// ========================
// USER MODEL
// ========================

type UserRole string

const (
	RoleSuperAdmin UserRole = "SuperAdmin"
	RoleAdmin      UserRole = "Admin"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"` // Never expose password
	Role      UserRole  `gorm:"type:varchar(20);not null" json:"role"`
	ShopID    uuid.UUID `gorm:"type:uuid;not null" json:"shop_id"`
	Shop      Shop      `gorm:"foreignKey:ShopID" json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}

// ========================
// PRODUCT MODEL
// ========================

type Product struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name          string         `gorm:"not null" json:"name"`
	Description   string         `json:"description"`
	Category      string         `json:"category"`
	PurchasePrice float64        `gorm:"not null" json:"purchase_price,omitempty"` // Hidden in public routes via DTO
	SellingPrice  float64        `gorm:"not null" json:"selling_price"`
	Stock         int            `gorm:"default:0" json:"stock"`
	ImageURL      string         `json:"image_url"`
	ShopID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"shop_id"`
	Shop          Shop           `gorm:"foreignKey:ShopID" json:"-"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}

// ========================
// TRANSACTION MODEL
// ========================

type TransactionType string

const (
	TransactionSale       TransactionType = "Sale"
	TransactionExpense    TransactionType = "Expense"
	TransactionWithdrawal TransactionType = "Withdrawal"
)

type Transaction struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Type      TransactionType `gorm:"type:varchar(20);not null" json:"type"`
	ProductID *uuid.UUID      `gorm:"type:uuid" json:"product_id,omitempty"`
	Product   *Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int             `json:"quantity"`
	Amount    float64         `gorm:"not null" json:"amount"`
	Comment   string          `gorm:"type:text" json:"comment,omitempty"`
	ShopID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"shop_id"`
	CreatedAt time.Time       `json:"created_at"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	t.ID = uuid.New()
	return nil
}
