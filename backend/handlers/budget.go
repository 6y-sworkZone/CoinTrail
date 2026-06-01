package handlers

import (
	"cointrail/database"
	"cointrail/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateBudgetRequest struct {
	CategoryID uint    `json:"category_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Month      string  `json:"month" binding:"required,len=7"`
}

type UpdateBudgetRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type BudgetWithUsage struct {
	models.Budget
	UsedAmount  float64 `json:"used_amount"`
	Remaining   float64 `json:"remaining"`
	Percentage  float64 `json:"percentage"`
	IsOverBudget bool   `json:"is_over_budget"`
	CategoryName string `json:"category_name"`
	CategoryIcon string `json:"category_icon"`
	CategoryColor string `json:"category_color"`
}

func GetBudgets(c *gin.Context) {
	userID := c.GetUint("user_id")
	month := c.Query("month")

	if month == "" {
		now := time.Now()
		month = now.Format("2006-01")
	}

	var budgets []models.Budget
	if err := database.DB.Where("user_id = ? AND month = ?", userID, month).
		Preload("Category").
		Find(&budgets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch budgets"})
		return
	}

	startDate := month + "-01"
	endDate := month + "-31"

	response := make([]BudgetWithUsage, len(budgets))
	for i, budget := range budgets {
		var usedAmount float64
		database.DB.Model(&models.Transaction{}).
			Where("user_id = ? AND category_id = ? AND type = ? AND transaction_date >= ? AND transaction_date <= ?",
				userID, budget.CategoryID, models.TransactionTypeExpense, startDate, endDate).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&usedAmount)

		percentage := 0.0
		if budget.Amount > 0 {
			percentage = (usedAmount / budget.Amount) * 100
		}

		response[i] = BudgetWithUsage{
			Budget:       budget,
			UsedAmount:   usedAmount,
			Remaining:    budget.Amount - usedAmount,
			Percentage:   percentage,
			IsOverBudget: usedAmount > budget.Amount,
			CategoryName: budget.Category.Name,
			CategoryIcon: budget.Category.Icon,
			CategoryColor: budget.Category.Color,
		}
	}

	c.JSON(http.StatusOK, response)
}

func GetBudget(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var budget models.Budget
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).
		Preload("Category").
		First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	startDate := budget.Month + "-01"
	endDate := budget.Month + "-31"

	var usedAmount float64
	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND category_id = ? AND type = ? AND transaction_date >= ? AND transaction_date <= ?",
			userID, budget.CategoryID, models.TransactionTypeExpense, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&usedAmount)

	percentage := 0.0
	if budget.Amount > 0 {
		percentage = (usedAmount / budget.Amount) * 100
	}

	response := BudgetWithUsage{
		Budget:       budget,
		UsedAmount:   usedAmount,
		Remaining:    budget.Amount - usedAmount,
		Percentage:   percentage,
		IsOverBudget: usedAmount > budget.Amount,
		CategoryName: budget.Category.Name,
		CategoryIcon: budget.Category.Icon,
		CategoryColor: budget.Category.Color,
	}

	c.JSON(http.StatusOK, response)
}

func CreateBudget(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var category models.Category
	if err := database.DB.Where("id = ? AND user_id = ? AND type = ?", req.CategoryID, userID, models.CategoryTypeExpense).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense category not found"})
		return
	}

	var existingBudget models.Budget
	if err := database.DB.Where("user_id = ? AND category_id = ? AND month = ?", userID, req.CategoryID, req.Month).First(&existingBudget).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Budget for this category and month already exists"})
		return
	}

	budget := models.Budget{
		UserID:     userID,
		CategoryID: req.CategoryID,
		Amount:     req.Amount,
		Month:      req.Month,
	}

	if err := database.DB.Create(&budget).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create budget"})
		return
	}

	c.JSON(http.StatusCreated, budget)
}

func UpdateBudget(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var budget models.Budget
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	var req UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	budget.Amount = req.Amount

	if err := database.DB.Save(&budget).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update budget"})
		return
	}

	c.JSON(http.StatusOK, budget)
}

func DeleteBudget(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var budget models.Budget
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	if err := database.DB.Delete(&budget).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete budget"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Budget deleted successfully"})
}

func GetDashboardBudgets(c *gin.Context) {
	userID := c.GetUint("user_id")
	now := time.Now()
	month := now.Format("2006-01")

	var budgets []models.Budget
	if err := database.DB.Where("user_id = ? AND month = ?", userID, month).
		Preload("Category").
		Find(&budgets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch budgets"})
		return
	}

	startDate := month + "-01"
	endDate := month + "-31"

	response := make([]BudgetWithUsage, len(budgets))
	for i, budget := range budgets {
		var usedAmount float64
		database.DB.Model(&models.Transaction{}).
			Where("user_id = ? AND category_id = ? AND type = ? AND transaction_date >= ? AND transaction_date <= ?",
				userID, budget.CategoryID, models.TransactionTypeExpense, startDate, endDate).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&usedAmount)

		percentage := 0.0
		if budget.Amount > 0 {
			percentage = (usedAmount / budget.Amount) * 100
		}

		response[i] = BudgetWithUsage{
			Budget:       budget,
			UsedAmount:   usedAmount,
			Remaining:    budget.Amount - usedAmount,
			Percentage:   percentage,
			IsOverBudget: usedAmount > budget.Amount,
			CategoryName: budget.Category.Name,
			CategoryIcon: budget.Category.Icon,
			CategoryColor: budget.Category.Color,
		}
	}

	c.JSON(http.StatusOK, response)
}
