package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthRequired validates JWT and injects user context
// SECURITY: shopID is ALWAYS extracted from the JWT token - never from URL params
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Format: Authorization: Bearer <token>",
			})
			return
		}

		tokenString := parts[1]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "default-secret-change-in-production"
		}

		// Parse with MapClaims (handles string UUIDs in JWT)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		userIDStr, _ := claims["user_id"].(string)
		shopIDStr, _ := claims["shop_id"].(string)
		role, _ := claims["role"].(string)
		email, _ := claims["email"].(string)

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			return
		}

		shopID, err := uuid.Parse(shopIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid shop ID in token"})
			return
		}

		// CRITICAL: Set context from JWT only
		c.Set("userID", userID)
		c.Set("shopID", shopID)
		c.Set("role", role)
		c.Set("email", email)

		c.Next()
	}
}

// CheckRole verifies the user has one of the required roles
func CheckRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Role not found"})
			return
		}

		userRole := role.(string)
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Access denied. Required: " + strings.Join(roles, " or "),
		})
	}
}

// GetShopIDFromContext safely extracts shopID from Gin context
func GetShopIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("shopID")
	if !exists {
		return uuid.UUID{}, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetUserIDFromContext safely extracts userID from Gin context
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return uuid.UUID{}, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetRoleFromContext safely extracts role string from Gin context
func GetRoleFromContext(c *gin.Context) string {
	val, _ := c.Get("role")
	if r, ok := val.(string); ok {
		return r
	}
	return ""
}
