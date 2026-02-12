package config

import (
	"fmt"
	"log"
	"os"

	"electronic-shop/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// LoadEnv loads .env file if present
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
}

// GetEnv returns env variable or fallback default
func GetEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// ConnectDB connects to PostgreSQL and returns a *gorm.DB
func ConnectDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		GetEnv("DB_HOST", "localhost"),
		GetEnv("DB_USER", "postgres"),
		GetEnv("DB_PASSWORD", "postgres"),
		GetEnv("DB_NAME", "electronic_shop"),
		GetEnv("DB_PORT", "5432"),
		GetEnv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Database connected successfully")

	// Create indexes after migration
	db.Exec("CREATE INDEX IF NOT EXISTS idx_products_shop_id ON products(shop_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_transactions_shop_id ON transactions(shop_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_shop_id ON users(shop_id)")

	// Seed default SuperAdmin shop if none exist
	seedDefaultShop(db)

	return db
}

func seedDefaultShop(db *gorm.DB) {
	var count int64
	db.Model(&models.Shop{}).Count(&count)
	if count == 0 {
		log.Println("📦 No shops found. Use POST /auth/register to create the first SuperAdmin.")
	}
}
