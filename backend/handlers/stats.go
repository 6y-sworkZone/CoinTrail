package handlers

import (
	"cointrail/database"
	"cointrail/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type MonthlySummary struct {
	Month   string  `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Balance float64 `json:"balance"`
}

type CategorySummary struct {
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	CategoryIcon string  `json:"category_icon"`
	CategoryColor string `json:"category_color"`
	Amount       float64 `json:"amount"`
	Percentage   float64 `json:"percentage"`
}

type DashboardStats struct {
	TotalBalance    float64 `json:"total_balance"`
	MonthlyIncome   float64 `json:"monthly_income"`
	MonthlyExpense  float64 `json:"monthly_expense"`
	MonthlyBalance  float64 `json:"monthly_balance"`
	TodayIncome     float64 `json:"today_income"`
	TodayExpense    float64 `json:"today_expense"`
}

func GetDashboard(c *gin.Context) {
	userID := c.GetUint("user_id")
	now := time.Now()
	currentMonth := now.Format("2006-01")
	today := now.Format("2006-01-02")

	var totalBalance float64
	database.DB.Model(&models.Account{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&totalBalance)

	var monthlyIncome float64
	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND strftime('%Y-%m', transaction_date) = ?",
			userID, models.TransactionTypeIncome, currentMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&monthlyIncome)

	var monthlyExpense float64
	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND strftime('%Y-%m', transaction_date) = ?",
			userID, models.TransactionTypeExpense, currentMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&monthlyExpense)

	var todayIncome float64
	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND DATE(transaction_date) = ?",
			userID, models.TransactionTypeIncome, today).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&todayIncome)

	var todayExpense float64
	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND DATE(transaction_date) = ?",
			userID, models.TransactionTypeExpense, today).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&todayExpense)

	stats := DashboardStats{
		TotalBalance:   totalBalance,
		MonthlyIncome:  monthlyIncome,
		MonthlyExpense: monthlyExpense,
		MonthlyBalance: monthlyIncome - monthlyExpense,
		TodayIncome:    todayIncome,
		TodayExpense:   todayExpense,
	}

	c.JSON(http.StatusOK, stats)
}

func GetMonthlySummary(c *gin.Context) {
	userID := c.GetUint("user_id")
	startMonth := c.Query("start_month")
	endMonth := c.Query("end_month")

	if startMonth == "" || endMonth == "" {
		now := time.Now()
		endMonth = now.Format("2006-01")
		startMonth = now.AddDate(0, -11, 0).Format("2006-01")
	}

	var results []struct {
		Month   string
		Type    string
		Amount  float64
	}

	query := database.DB.Model(&models.Transaction{}).
		Select("strftime('%Y-%m', transaction_date) as month, type, COALESCE(SUM(amount), 0) as amount").
		Where("user_id = ? AND type != ? AND strftime('%Y-%m', transaction_date) >= ? AND strftime('%Y-%m', transaction_date) <= ?",
			userID, models.TransactionTypeTransfer, startMonth, endMonth).
		Group("month, type").
		Order("month ASC")

	query.Scan(&results)

	monthlyMap := make(map[string]*MonthlySummary)
	for _, r := range results {
		if _, exists := monthlyMap[r.Month]; !exists {
			monthlyMap[r.Month] = &MonthlySummary{Month: r.Month}
		}
		if r.Type == string(models.TransactionTypeIncome) {
			monthlyMap[r.Month].Income = r.Amount
		} else {
			monthlyMap[r.Month].Expense = r.Amount
		}
	}

	var summary []MonthlySummary
	current, _ := time.Parse("2006-01", startMonth)
	end, _ := time.Parse("2006-01", endMonth)

	for !current.After(end) {
		monthStr := current.Format("2006-01")
		if s, exists := monthlyMap[monthStr]; exists {
			s.Balance = s.Income - s.Expense
			summary = append(summary, *s)
		} else {
			summary = append(summary, MonthlySummary{Month: monthStr})
		}
		current = current.AddDate(0, 1, 0)
	}

	c.JSON(http.StatusOK, summary)
}

func GetCategorySummary(c *gin.Context) {
	userID := c.GetUint("user_id")
	transactionType := c.Query("type")
	month := c.Query("month")

	if transactionType == "" {
		transactionType = string(models.TransactionTypeExpense)
	}
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	var totalAmount float64
	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND type = ? AND strftime('%Y-%m', transaction_date) = ?",
			userID, transactionType, month).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalAmount)

	var results []struct {
		CategoryID   uint
		CategoryName string
		CategoryIcon string
		CategoryColor string
		Amount       float64
	}

	database.DB.Model(&models.Transaction{}).
		Select("transactions.category_id, categories.name as category_name, categories.icon as category_icon, categories.color as category_color, COALESCE(SUM(transactions.amount), 0) as amount").
		Joins("LEFT JOIN categories ON transactions.category_id = categories.id").
		Where("transactions.user_id = ? AND transactions.type = ? AND strftime('%Y-%m', transactions.transaction_date) = ?",
			userID, transactionType, month).
		Group("transactions.category_id").
		Order("amount DESC").
		Scan(&results)

	summary := make([]CategorySummary, len(results))
	for i, r := range results {
		percentage := 0.0
		if totalAmount > 0 {
			percentage = (r.Amount / totalAmount) * 100
		}
		summary[i] = CategorySummary{
			CategoryID:    r.CategoryID,
			CategoryName:  r.CategoryName,
			CategoryIcon:  r.CategoryIcon,
			CategoryColor: r.CategoryColor,
			Amount:        r.Amount,
			Percentage:    percentage,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   totalAmount,
		"details": summary,
	})
}

func GetTrendStats(c *gin.Context) {
	userID := c.GetUint("user_id")
	months := 12

	var results []struct {
		Month   string
		Type    string
		Amount  float64
	}

	startDate := time.Now().AddDate(0, -months+1, 0).Format("2006-01") + "-01"

	database.DB.Model(&models.Transaction{}).
		Select("strftime('%Y-%m', transaction_date) as month, type, COALESCE(SUM(amount), 0) as amount").
		Where("user_id = ? AND type != ? AND transaction_date >= ?",
			userID, models.TransactionTypeTransfer, startDate).
		Group("month, type").
		Order("month ASC").
		Scan(&results)

	monthlyMap := make(map[string]*MonthlySummary)
	for _, r := range results {
		if _, exists := monthlyMap[r.Month]; !exists {
			monthlyMap[r.Month] = &MonthlySummary{Month: r.Month}
		}
		if r.Type == string(models.TransactionTypeIncome) {
			monthlyMap[r.Month].Income = r.Amount
		} else {
			monthlyMap[r.Month].Expense = r.Amount
		}
	}

	var trend []MonthlySummary
	now := time.Now()
	for i := months - 1; i >= 0; i-- {
		month := now.AddDate(0, -i, 0).Format("2006-01")
		if s, exists := monthlyMap[month]; exists {
			s.Balance = s.Income - s.Expense
			trend = append(trend, *s)
		} else {
			trend = append(trend, MonthlySummary{Month: month})
		}
	}

	c.JSON(http.StatusOK, trend)
}
