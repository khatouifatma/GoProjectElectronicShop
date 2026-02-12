package handlers

import (
	"net/http"

	"electronic-shop/internal/dto"
	"electronic-shop/internal/middleware"
	"electronic-shop/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReportHandler struct {
	db *gorm.DB
}

func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

// GetDashboard - returns financial summary for SuperAdmin
func (h *ReportHandler) GetDashboard(c *gin.Context) {
	shopID, ok := middleware.GetShopIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Total sales (sum of all Sale transactions)
	var totalSales float64
	h.db.Model(&models.Transaction{}).
		Where("shop_id = ? AND type = ?", shopID, models.TransactionSale).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalSales)

	// Total expenses (Expense + Withdrawal)
	var totalExpenses float64
	h.db.Model(&models.Transaction{}).
		Where("shop_id = ? AND type IN ?", shopID, []string{
			string(models.TransactionExpense),
			string(models.TransactionWithdrawal),
		}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpenses)

	// Low stock products (stock < 5)
	var lowStockProducts []models.Product
	h.db.Where("shop_id = ? AND stock < 5", shopID).
		Find(&lowStockProducts)

	var lowStockItems []dto.LowStockItem
	for _, p := range lowStockProducts {
		lowStockItems = append(lowStockItems, dto.LowStockItem{
			ID:       p.ID,
			Name:     p.Name,
			Stock:    p.Stock,
			Category: p.Category,
		})
	}
	if lowStockItems == nil {
		lowStockItems = []dto.LowStockItem{}
	}

	// Count totals
	var totalProducts int64
	h.db.Model(&models.Product{}).Where("shop_id = ?", shopID).Count(&totalProducts)

	var totalTransactions int64
	h.db.Model(&models.Transaction{}).Where("shop_id = ?", shopID).Count(&totalTransactions)

	netProfit := totalSales - totalExpenses

	c.JSON(http.StatusOK, dto.DashboardResponse{
		TotalSales:        totalSales,
		TotalExpenses:     totalExpenses,
		NetProfit:         netProfit,
		LowStockProducts:  lowStockItems,
		TotalProducts:     totalProducts,
		TotalTransactions: totalTransactions,
	})
}
