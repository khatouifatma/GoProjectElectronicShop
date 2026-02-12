package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"electronic-shop/internal/dto"
	"electronic-shop/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PublicHandler struct {
	db *gorm.DB
}

func NewPublicHandler(db *gorm.DB) *PublicHandler {
	return &PublicHandler{db: db}
}

// buildWhatsAppLink generates the formatted WhatsApp redirect URL
func buildWhatsAppLink(whatsAppNumber, productName string) string {
	message := fmt.Sprintf("Bonjour je veux plus d'information sur %s", productName)
	encodedMessage := url.QueryEscape(message)
	return fmt.Sprintf("https://wa.me/%s?text=%s", whatsAppNumber, encodedMessage)
}

// GetPublicProducts - returns products for a shop (no auth required)
// SECURITY: Never exposes PurchasePrice
func (h *PublicHandler) GetPublicProducts(c *gin.Context) {
	shopIDStr := c.Param("shopID")
	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shop ID format"})
		return
	}

	// Verify shop exists and is active
	var shop models.Shop
	if err := h.db.Where("id = ? AND active = true", shopID).First(&shop).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shop not found or inactive"})
		return
	}

	// Optional: filter by category
	category := c.Query("category")
	hideOutOfStock := c.Query("in_stock_only") == "true"

	query := h.db.Where("shop_id = ?", shopID)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if hideOutOfStock {
		query = query.Where("stock > 0")
	}

	var products []models.Product
	if err := query.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	// Build public response - NEVER include PurchasePrice
	var responses []dto.PublicProductResponse
	for _, p := range products {
		stockStatus := "En stock"
		if p.Stock == 0 {
			stockStatus = "Rupture de stock"
		} else if p.Stock < 5 {
			stockStatus = "Stock limité"
		}

		responses = append(responses, dto.PublicProductResponse{
			ID:           p.ID,
			Name:         p.Name,
			Description:  p.Description,
			Category:     p.Category,
			SellingPrice: p.SellingPrice,
			Stock:        p.Stock,
			StockStatus:  stockStatus,
			ImageURL:     p.ImageURL,
			WhatsAppLink: buildWhatsAppLink(shop.WhatsAppNumber, p.Name),
		})
	}

	if responses == nil {
		responses = []dto.PublicProductResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"shop": gin.H{
			"id":   shop.ID,
			"name": shop.Name,
		},
		"products": responses,
		"total":    len(responses),
	})
}

// GetWhatsAppLink - returns the WhatsApp redirect link for a specific product
func (h *PublicHandler) GetWhatsAppLink(c *gin.Context) {
	shopIDStr := c.Param("shopID")
	productIDStr := c.Param("productID")

	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shop ID"})
		return
	}

	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Get shop (for WhatsApp number)
	var shop models.Shop
	if err := h.db.Where("id = ? AND active = true", shopID).First(&shop).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shop not found"})
		return
	}

	// Get product - must belong to this shop (multi-tenant)
	var product models.Product
	if err := h.db.Where("id = ? AND shop_id = ?", productID, shopID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	whatsappLink := buildWhatsAppLink(shop.WhatsAppNumber, product.Name)

	c.JSON(http.StatusOK, gin.H{
		"product_id":    product.ID,
		"product_name":  product.Name,
		"whatsapp_link": whatsappLink,
		"shop_name":     shop.Name,
	})
}
