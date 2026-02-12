package handlers

import (
	"net/http"
	"os"
	"time"

	"electronic-shop/internal/dto"
	"electronic-shop/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Register - creates a new user and optionally a new shop
// If shop_name + whatsapp_number are provided: creates new shop + SuperAdmin
// If shop_id is provided: adds user to existing shop (SuperAdmin only can do this via /users endpoint)
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	var shopID uuid.UUID

	if req.ShopID != "" {
		// Join existing shop
		parsedID, err := uuid.Parse(req.ShopID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shop_id format"})
			return
		}
		var shop models.Shop
		if err := h.db.First(&shop, "id = ?", parsedID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Shop not found"})
			return
		}
		shopID = parsedID
	} else {
		// Create new shop (requires shop_name and whatsapp_number)
		if req.ShopName == "" || req.WhatsAppNumber == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "shop_name and whatsapp_number are required when creating a new shop",
			})
			return
		}
		if req.Role != string(models.RoleSuperAdmin) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Only SuperAdmin can create a new shop",
			})
			return
		}

		shop := models.Shop{
			Name:           req.ShopName,
			WhatsAppNumber: req.WhatsAppNumber,
			Active:         true,
		}
		if err := h.db.Create(&shop).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shop"})
			return
		}
		shopID = shop.ID
	}

	// Hash password with bcrypt cost=12
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
		ShopID:   shopID,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": dto.UserResponse{
			ID:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Role:   string(user.Role),
			ShopID: user.ShopID,
		},
	})
}

// Login - authenticates user and returns JWT token
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Return generic error to prevent email enumeration
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token: token,
		User: dto.UserResponse{
			ID:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Role:   string(user.Role),
			ShopID: user.ShopID,
		},
	})
}

// generateToken creates a signed JWT with user claims
func generateToken(user models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}

	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"shop_id": user.ShopID.String(),
		"role":    string(user.Role),
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
