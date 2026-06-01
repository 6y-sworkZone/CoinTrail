package handlers

import (
	"cointrail/database"
	"cointrail/models"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ExportTransactions(c *gin.Context) {
	userID := c.GetUint("user_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := database.DB.Where("user_id = ? AND type != ?", userID, models.TransactionTypeTransfer)

	if startDate != "" {
		query = query.Where("transaction_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("transaction_date <= ?", endDate)
	}

	var transactions []models.Transaction
	if err := query.Preload("Account").Preload("Category").
		Order("transaction_date DESC").
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=transactions_%s.csv", time.Now().Format("20060102_150405")))
	c.Header("Cache-Control", "no-cache")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	writer.Write([]string{"\xEF\xBB\xBF"})

	headers := []string{"日期", "类型", "分类", "账户", "金额", "备注"}
	writer.Write(headers)

	for _, t := range transactions {
		typeStr := "收入"
		if t.Type == models.TransactionTypeExpense {
			typeStr = "支出"
		}

		categoryName := ""
		if t.Category != nil {
			categoryName = t.Category.Name
		}

		row := []string{
			t.TransactionDate.Format("2006-01-02"),
			typeStr,
			categoryName,
			t.Account.Name,
			fmt.Sprintf("%.2f", t.Amount),
			t.Note,
		}
		writer.Write(row)
	}
}

func ImportTransactions(c *gin.Context) {
	userID := c.GetUint("user_id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only CSV files are allowed"})
		return
	}

	reader := csv.NewReader(file)
	reader.Comma = ','

	_, err = reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read CSV header"})
		return
	}

	var accounts []models.Account
	database.DB.Where("user_id = ?", userID).Find(&accounts)
	accountMap := make(map[string]uint)
	for _, acc := range accounts {
		accountMap[strings.ToLower(acc.Name)] = acc.ID
	}

	var categories []models.Category
	database.DB.Where("user_id = ?", userID).Find(&categories)
	categoryMap := make(map[string]uint)
	for _, cat := range categories {
		categoryMap[strings.ToLower(cat.Name)] = cat.ID
	}

	var importResults []map[string]interface{}
	successCount := 0
	skipCount := 0
	errorCount := 0

	tx := database.DB.Begin()

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errorCount++
			continue
		}

		if len(row) < 5 {
			errorCount++
			continue
		}

		dateStr := strings.TrimSpace(row[0])
		typeStr := strings.TrimSpace(row[1])
		categoryStr := strings.TrimSpace(row[2])
		accountStr := strings.TrimSpace(row[3])
		amountStr := strings.TrimSpace(row[4])
		noteStr := ""
		if len(row) > 5 {
			noteStr = strings.TrimSpace(row[5])
		}

		var transType models.TransactionType
		if strings.Contains(typeStr, "收入") || strings.Contains(typeStr, "income") {
			transType = models.TransactionTypeIncome
		} else if strings.Contains(typeStr, "支出") || strings.Contains(typeStr, "expense") {
			transType = models.TransactionTypeExpense
		} else {
			errorCount++
			importResults = append(importResults, map[string]interface{}{
				"row":    row,
				"status": "error",
				"reason": "Invalid transaction type",
			})
			continue
		}

		transDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			transDate, err = time.Parse("2006/01/02", dateStr)
			if err != nil {
				errorCount++
				importResults = append(importResults, map[string]interface{}{
					"row":    row,
					"status": "error",
					"reason": "Invalid date format",
				})
				continue
			}
		}

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil || amount <= 0 {
			errorCount++
			importResults = append(importResults, map[string]interface{}{
				"row":    row,
				"status": "error",
				"reason": "Invalid amount",
			})
			continue
		}

		accountID, exists := accountMap[strings.ToLower(accountStr)]
		if !exists {
			if len(accounts) == 0 {
				errorCount++
				importResults = append(importResults, map[string]interface{}{
					"row":    row,
					"status": "error",
					"reason": "No accounts found. Please create an account first.",
				})
				continue
			}
			accountID = accounts[0].ID
		}

		var categoryID *uint
		if catID, exists := categoryMap[strings.ToLower(categoryStr)]; exists {
			categoryID = &catID
		}

		var existingCount int64
		tx.Model(&models.Transaction{}).
			Where("user_id = ? AND type = ? AND account_id = ? AND amount = ? AND DATE(transaction_date) = ?",
				userID, transType, accountID, amount, transDate.Format("2006-01-02")).
			Count(&existingCount)

		if existingCount > 0 {
			skipCount++
			importResults = append(importResults, map[string]interface{}{
				"row":    row,
				"status": "skipped",
				"reason": "Duplicate transaction",
			})
			continue
		}

		var account models.Account
		if err := tx.Where("id = ?", accountID).First(&account).Error; err != nil {
			errorCount++
			continue
		}

		if transType == models.TransactionTypeExpense {
			account.Balance -= amount
		} else {
			account.Balance += amount
		}
		tx.Save(&account)

		transaction := models.Transaction{
			UserID:          userID,
			AccountID:       accountID,
			CategoryID:      categoryID,
			Type:            transType,
			Amount:          amount,
			Note:            noteStr,
			TransactionDate: transDate,
		}

		if err := tx.Create(&transaction).Error; err != nil {
			errorCount++
			importResults = append(importResults, map[string]interface{}{
				"row":    row,
				"status": "error",
				"reason": "Failed to create transaction",
			})
			continue
		}

		successCount++
		importResults = append(importResults, map[string]interface{}{
			"row":    row,
			"status": "success",
		})
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Import completed",
		"success": successCount,
		"skipped": skipCount,
		"errors":  errorCount,
		"details": importResults,
	})
}
