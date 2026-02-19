package handlers

import (
	"errors"
	"net/http"
	"time"

	"electronic-shop/internal/dto"
	"electronic-shop/internal/middleware"
	"electronic-shop/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransactionHandler struct {
	db *gorm.DB
}

func NewTransactionHandler(db *gorm.DB) *TransactionHandler {
	return &TransactionHandler{db: db}
}

// GetTransactions - returns all transactions for the authenticated user's shop
func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	transactionType := c.Query("type")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	query := h.db.Where("shop_id = ?", shopID).Preload("Product")

	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}

	// date_from : début de journée (00:00:00)
	if dateFrom != "" {
		t, err := time.Parse("2006-01-02", dateFrom)
		if err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}

	// date_to : fin de journée (23:59:59)
	if dateTo != "" {
		t, err := time.Parse("2006-01-02", dateTo)
		if err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour-time.Second))
		}
	}

	var transactions []models.Transaction
	if err := query.Order("created_at DESC").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"total":        len(transactions),
	})
}

// CreateTransaction - creates a transaction with stock management
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use a DB transaction for atomicity
	var transaction models.Transaction

	err := h.db.Transaction(func(tx *gorm.DB) error {
		// If it's a Sale, validate product stock
		if req.Type == string(models.TransactionSale) {
			if req.ProductID == nil {
				return errors.New("product_id is required for Sale transactions")
			}
			if req.Quantity <= 0 {
				return errors.New("quantity must be greater than 0 for Sales")
			}

			// Fetch product - MUST belong to same shop
			var product models.Product
			if err := tx.Where("id = ? AND shop_id = ?", *req.ProductID, shopID).First(&product).Error; err != nil {
				return errors.New("product not found")
			}

			// CRITICAL: Prevent negative stock
			if product.Stock < req.Quantity {
				return errors.New("insufficient stock: available " + string(rune('0'+product.Stock)))
			}

			// Deduct stock atomically
			if err := tx.Model(&product).Update("stock", product.Stock-req.Quantity).Error; err != nil {
				return errors.New("failed to update stock")
			}
		}

		// Create transaction record
		transaction = models.Transaction{
			Type:      models.TransactionType(req.Type),
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
			Amount:    req.Amount,
			Comment:   req.Comment,
			ShopID:    shopID, // Always from JWT
		}

		return tx.Create(&transaction).Error
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Reload with product info
	h.db.Preload("Product").First(&transaction, "id = ?", transaction.ID)

	c.JSON(http.StatusCreated, transaction)
}
