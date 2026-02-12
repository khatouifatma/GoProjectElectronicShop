package handlers

import (
	"net/http"

	"electronic-shop/internal/dto"
	"electronic-shop/internal/middleware"
	"electronic-shop/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// GetUsers - returns all users in the authenticated user's shop
func (h *UserHandler) GetUsers(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var users []models.User
	if err := h.db.Where("shop_id = ?", shopID).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	var responses []dto.UserResponse
	for _, u := range users {
		responses = append(responses, dto.UserResponse{
			ID:     u.ID,
			Name:   u.Name,
			Email:  u.Email,
			Role:   string(u.Role),
			ShopID: u.ShopID,
		})
	}

	if responses == nil {
		responses = []dto.UserResponse{}
	}

	c.JSON(http.StatusOK, gin.H{"users": responses, "total": len(responses)})
}

// CreateUser - SuperAdmin creates a new user in their shop
func (h *UserHandler) CreateUser(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check email uniqueness
	var existing models.User
	if err := h.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     models.UserRole(req.Role),
		ShopID:   shopID, // Always assign to current shop
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, dto.UserResponse{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Role:   string(user.Role),
		ShopID: user.ShopID,
	})
}

// DeleteUser - SuperAdmin deletes a user from their shop
func (h *UserHandler) DeleteUser(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Prevent self-deletion
	currentUserID, _ := middleware.GetUserIDFromContext(c)
	if userID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// Only delete users in the same shop (multi-tenant isolation)
	result := h.db.Where("id = ? AND shop_id = ?", userID, shopID).Delete(&models.User{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in your shop"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
