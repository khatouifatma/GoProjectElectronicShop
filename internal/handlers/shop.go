package handlers

import (
	"net/http"

	"electronic-shop/internal/dto"
	"electronic-shop/internal/middleware"
	"electronic-shop/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ShopHandler struct {
	db *gorm.DB
}

func NewShopHandler(db *gorm.DB) *ShopHandler {
	return &ShopHandler{db: db}
}

// GetShop - returns the current user's shop info
func (h *ShopHandler) GetShop(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var shop models.Shop
	if err := h.db.First(&shop, "id = ?", shopID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shop not found"})
		return
	}

	c.JSON(http.StatusOK, shop)
}

// UpdateWhatsApp - updates the shop's WhatsApp number (SuperAdmin only)
func (h *ShopHandler) UpdateWhatsApp(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.UpdateWhatsAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.db.Model(&models.Shop{}).
		Where("id = ?", shopID).
		Update("whats_app_number", req.WhatsAppNumber)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update WhatsApp number"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shop not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "WhatsApp number updated successfully",
		"whatsapp_number":  req.WhatsAppNumber,
	})
}
