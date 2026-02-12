package main

import (
	"log"

	"electronic-shop/config"
	"electronic-shop/internal/handlers"
	"electronic-shop/internal/middleware"
	"electronic-shop/internal/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Connect to database
	db := config.ConnectDB()

	// Auto-migrate all models
	if err := db.AutoMigrate(
		&models.Shop{},
		&models.User{},
		&models.Product{},
		&models.Transaction{},
	); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	shopHandler := handlers.NewShopHandler(db)
	productHandler := handlers.NewProductHandler(db)
	transactionHandler := handlers.NewTransactionHandler(db)
	reportHandler := handlers.NewReportHandler(db)
	publicHandler := handlers.NewPublicHandler(db)

	// ========================
	// PUBLIC ROUTES (no auth)
	// ========================
	public := r.Group("/public")
	{
		public.GET("/:shopID/products", publicHandler.GetPublicProducts)
		public.GET("/:shopID/products/:productID/whatsapp", publicHandler.GetWhatsAppLink)
	}

	// ========================
	// AUTH ROUTES
	// ========================
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// ========================
	// PRIVATE ROUTES (JWT required)
	// ========================
	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		// Shop management (SuperAdmin only)
		shops := api.Group("/shops")
		shops.Use(middleware.CheckRole("SuperAdmin"))
		{
			shops.GET("", shopHandler.GetShop)
			shops.PUT("/whatsapp", shopHandler.UpdateWhatsApp)
		}

		// Products (SuperAdmin + Admin)
		products := api.Group("/products")
		{
			products.GET("", productHandler.GetProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.POST("", productHandler.CreateProduct)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
		}

		// Transactions (SuperAdmin + Admin)
		transactions := api.Group("/transactions")
		{
			transactions.GET("", transactionHandler.GetTransactions)
			transactions.POST("", transactionHandler.CreateTransaction)
		}

		// Users management (SuperAdmin only)
		users := api.Group("/users")
		users.Use(middleware.CheckRole("SuperAdmin"))
		{
			users.GET("", handlers.NewUserHandler(db).GetUsers)
			users.POST("", handlers.NewUserHandler(db).CreateUser)
			users.DELETE("/:id", handlers.NewUserHandler(db).DeleteUser)
		}

		// Dashboard (SuperAdmin only)
		reports := api.Group("/reports")
		reports.Use(middleware.CheckRole("SuperAdmin"))
		{
			reports.GET("/dashboard", reportHandler.GetDashboard)
		}
	}

	port := config.GetEnv("PORT", "8080")
	log.Printf("🚀 Server running on http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
