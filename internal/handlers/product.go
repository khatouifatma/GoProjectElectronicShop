package handlers

import (
	"net/http"

	"electronic-shop/internal/dto"
	"electronic-shop/internal/middleware"
	"electronic-shop/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

// toPrivateResponse converts a product to a response DTO
// SuperAdmin sees PurchasePrice, Admin does not
func toPrivateResponse(p models.Product, role string) dto.PrivateProductResponse {
	resp := dto.PrivateProductResponse{
		ID:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		Category:     p.Category,
		SellingPrice: p.SellingPrice,
		Stock:        p.Stock,
		ImageURL:     p.ImageURL,
		ShopID:       p.ShopID,
	}
	// Only SuperAdmin can see purchase price
	if role == string(models.RoleSuperAdmin) {
		resp.PurchasePrice = p.PurchasePrice
	}
	return resp
}

// GetProducts - returns all products for the authenticated user's shop
func (h *ProductHandler) GetProducts(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	role := middleware.GetRoleFromContext(c)
	category := c.Query("category")
	search := c.Query("search")

	query := h.db.Where("shop_id = ?", shopID)

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	var products []models.Product
	if err := query.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	var responses []dto.PrivateProductResponse
	for _, p := range products {
		responses = append(responses, toPrivateResponse(p, role))
	}

	if responses == nil {
		responses = []dto.PrivateProductResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"products": responses,
		"total":    len(responses),
	})
}

// GetProduct - returns a single product by ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	// CRITICAL: Always filter by shopID from JWT to ensure isolation
	if err := h.db.Where("id = ? AND shop_id = ?", productID, shopID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	role := middleware.GetRoleFromContext(c)
	c.JSON(http.StatusOK, toPrivateResponse(product, role))
}

// CreateProduct - creates a new product in the authenticated user's shop
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		Name:          req.Name,
		Description:   req.Description,
		Category:      req.Category,
		PurchasePrice: req.PurchasePrice,
		SellingPrice:  req.SellingPrice,
		Stock:         req.Stock,
		ImageURL:      req.ImageURL,
		ShopID:        shopID, // Always use shopID from JWT
	}

	if err := h.db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	role := middleware.GetRoleFromContext(c)
	c.JSON(http.StatusCreated, toPrivateResponse(product, role))
}

// UpdateProduct - updates a product (must belong to user's shop)
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Verify product belongs to this shop (multi-tenant isolation)
	var product models.Product
	if err := h.db.Where("id = ? AND shop_id = ?", productID, shopID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update map (only update provided fields)
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.PurchasePrice > 0 {
		updates["purchase_price"] = req.PurchasePrice
	}
	if req.SellingPrice > 0 {
		updates["selling_price"] = req.SellingPrice
	}
	if req.Stock >= 0 {
		updates["stock"] = req.Stock
	}
	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	}

	if err := h.db.Model(&product).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	// Reload updated product
	h.db.First(&product, "id = ?", productID)
	role := middleware.GetRoleFromContext(c)
	c.JSON(http.StatusOK, toPrivateResponse(product, role))
}

// DeleteProduct - soft deletes a product (must belong to user's shop)
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// CRITICAL: Always include shopID in delete query
	result := h.db.Where("id = ? AND shop_id = ?", productID, shopID).Delete(&models.Product{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
