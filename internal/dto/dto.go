package dto

import "github.com/google/uuid"

// ========================
// AUTH DTOs
// ========================

type RegisterRequest struct {
	Name           string `json:"name" binding:"required,min=2"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	Role           string `json:"role" binding:"required,oneof=SuperAdmin Admin"`
	ShopName       string `json:"shop_name"`        // Required only if first SuperAdmin
	WhatsAppNumber string `json:"whatsapp_number"`  // Required only if creating new shop
	ShopID         string `json:"shop_id"`          // Provide existing ShopID to join a shop
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	ShopID uuid.UUID `json:"shop_id"`
}

// ========================
// SHOP DTOs
// ========================

type UpdateWhatsAppRequest struct {
	WhatsAppNumber string `json:"whatsapp_number" binding:"required"`
}

// ========================
// PRODUCT DTOs
// ========================

type CreateProductRequest struct {
	Name          string  `json:"name" binding:"required,min=1"`
	Description   string  `json:"description"`
	Category      string  `json:"category"`
	PurchasePrice float64 `json:"purchase_price" binding:"required,gt=0"`
	SellingPrice  float64 `json:"selling_price" binding:"required,gt=0"`
	Stock         int     `json:"stock" binding:"min=0"`
	ImageURL      string  `json:"image_url"`
}

type UpdateProductRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Category      string  `json:"category"`
	PurchasePrice float64 `json:"purchase_price"`
	SellingPrice  float64 `json:"selling_price"`
	Stock         int     `json:"stock"`
	ImageURL      string  `json:"image_url"`
}

// PrivateProductResponse - for authenticated users
type PrivateProductResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	PurchasePrice float64   `json:"purchase_price"` // Only SuperAdmin sees this (filtered in handler)
	SellingPrice  float64   `json:"selling_price"`
	Stock         int       `json:"stock"`
	ImageURL      string    `json:"image_url"`
	ShopID        uuid.UUID `json:"shop_id"`
}

// PublicProductResponse - NEVER exposes PurchasePrice
type PublicProductResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Category     string    `json:"category"`
	SellingPrice float64   `json:"selling_price"`
	Stock        int       `json:"stock"`
	StockStatus  string    `json:"stock_status"`
	ImageURL     string    `json:"image_url"`
	WhatsAppLink string    `json:"whatsapp_link"`
}

// ========================
// TRANSACTION DTOs
// ========================

type CreateTransactionRequest struct {
	Type      string     `json:"type" binding:"required,oneof=Sale Expense Withdrawal"`
	ProductID *uuid.UUID `json:"product_id"`
	Quantity  int        `json:"quantity" binding:"min=0"`
	Amount    float64    `json:"amount" binding:"required,gt=0"`
}

// ========================
// DASHBOARD DTOs
// ========================

type DashboardResponse struct {
	TotalSales      float64          `json:"total_sales"`
	TotalExpenses   float64          `json:"total_expenses"`
	NetProfit       float64          `json:"net_profit"`
	LowStockProducts []LowStockItem  `json:"low_stock_products"`
	TotalProducts   int64            `json:"total_products"`
	TotalTransactions int64          `json:"total_transactions"`
}

type LowStockItem struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Stock    int       `json:"stock"`
	Category string    `json:"category"`
}

// ========================
// USER MANAGEMENT DTOs
// ========================

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=SuperAdmin Admin"`
}
