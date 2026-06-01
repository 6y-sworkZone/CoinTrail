package handlers

import (
	"cointrail/database"
	"cointrail/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateTransactionRequest struct {
	AccountID       uint                   `json:"account_id" binding:"required"`
	CategoryID      *uint                  `json:"category_id"`
	Type            models.TransactionType `json:"type" binding:"required,oneof=expense income"`
	Amount          float64                `json:"amount" binding:"required,gt=0"`
	Note            string                 `json:"note"`
	TransactionDate string                 `json:"transaction_date" binding:"required"`
}

type UpdateTransactionRequest struct {
	AccountID       uint                   `json:"account_id"`
	CategoryID      *uint                  `json:"category_id"`
	Type            models.TransactionType `json:"type" binding:"omitempty,oneof=expense income"`
	Amount          float64                `json:"amount"`
	Note            string                 `json:"note"`
	TransactionDate string                 `json:"transaction_date"`
}

type BatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

type TransactionResponse struct {
	models.Transaction
	AccountName  string `json:"account_name"`
	CategoryName string `json:"category_name"`
	CategoryIcon string `json:"category_icon"`
	CategoryColor string `json:"category_color"`
}

func GetTransactions(c *gin.Context) {
	userID := c.GetUint("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	transactionType := c.Query("type")
	categoryID := c.Query("category_id")
	accountID := c.Query("account_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	offset := (page - 1) * pageSize

	var transactions []models.Transaction
	query := database.DB.Where("user_id = ? AND type != ?", userID, models.TransactionTypeTransfer)

	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if startDate != "" {
		query = query.Where("transaction_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("transaction_date <= ?", endDate)
	}

	var total int64
	query.Model(&models.Transaction{}).Count(&total)

	if err := query.Preload("Account").Preload("Category").
		Order("transaction_date DESC, created_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	response := make([]TransactionResponse, len(transactions))
	for i, t := range transactions {
		response[i] = TransactionResponse{
			Transaction:   t,
			AccountName:   t.Account.Name,
		}
		if t.Category != nil {
			response[i].CategoryName = t.Category.Name
			response[i].CategoryIcon = t.Category.Icon
			response[i].CategoryColor = t.Category.Color
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       response,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func GetTransaction(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var transaction models.Transaction
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).
		Preload("Account").Preload("Category").
		First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	response := TransactionResponse{
		Transaction:   transaction,
		AccountName:   transaction.Account.Name,
	}
	if transaction.Category != nil {
		response.CategoryName = transaction.Category.Name
		response.CategoryIcon = transaction.Category.Icon
		response.CategoryColor = transaction.Category.Color
	}

	c.JSON(http.StatusOK, response)
}

func CreateTransaction(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transDate, err := time.Parse("2006-01-02", req.TransactionDate)
	if err != nil {
		transDate, err = time.Parse(time.RFC3339, req.TransactionDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
	}

	var account models.Account
	if err := database.DB.Where("id = ? AND user_id = ?", req.AccountID, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	if req.CategoryID != nil {
		var category models.Category
		if err := database.DB.Where("id = ? AND user_id = ?", *req.CategoryID, userID).First(&category).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
	}

	tx := database.DB.Begin()

	transaction := models.Transaction{
		UserID:          userID,
		AccountID:       req.AccountID,
		CategoryID:      req.CategoryID,
		Type:            req.Type,
		Amount:          req.Amount,
		Note:            req.Note,
		TransactionDate: transDate,
	}

	if req.Type == models.TransactionTypeExpense {
		account.Balance -= req.Amount
	} else {
		account.Balance += req.Amount
	}

	if err := tx.Save(&account).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
		return
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, transaction)
}

func UpdateTransaction(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var transaction models.Transaction
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if transaction.Type == models.TransactionTypeTransfer {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot update transfer transaction"})
		return
	}

	var req UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	var oldAccount models.Account
	if err := tx.Where("id = ?", transaction.AccountID).First(&oldAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Original account not found"})
		return
	}

	if transaction.Type == models.TransactionTypeExpense {
		oldAccount.Balance += transaction.Amount
	} else {
		oldAccount.Balance -= transaction.Amount
	}
	if err := tx.Save(&oldAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revert account balance"})
		return
	}

	newAccountID := req.AccountID
	if newAccountID == 0 {
		newAccountID = transaction.AccountID
	}

	var newAccount models.Account
	if err := tx.Where("id = ? AND user_id = ?", newAccountID, userID).First(&newAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "New account not found"})
		return
	}

	newAmount := req.Amount
	if newAmount == 0 {
		newAmount = transaction.Amount
	}

	newType := req.Type
	if newType == "" {
		newType = transaction.Type
	}

	if newType == models.TransactionTypeExpense {
		newAccount.Balance -= newAmount
	} else {
		newAccount.Balance += newAmount
	}
	if err := tx.Save(&newAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update new account balance"})
		return
	}

	transaction.AccountID = newAccountID
	transaction.Amount = newAmount
	transaction.Type = newType

	if req.CategoryID != nil {
		transaction.CategoryID = req.CategoryID
	}
	if req.Note != "" {
		transaction.Note = req.Note
	}
	if req.TransactionDate != "" {
		transDate, err := time.Parse("2006-01-02", req.TransactionDate)
		if err != nil {
			transDate, err = time.Parse(time.RFC3339, req.TransactionDate)
			if err == nil {
				transaction.TransactionDate = transDate
			}
		} else {
			transaction.TransactionDate = transDate
		}
	}

	if err := tx.Save(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, transaction)
}

func DeleteTransaction(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var transaction models.Transaction
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if transaction.Type == models.TransactionTypeTransfer {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete transfer transaction directly"})
		return
	}

	tx := database.DB.Begin()

	var account models.Account
	if err := tx.Where("id = ?", transaction.AccountID).First(&account).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	if transaction.Type == models.TransactionTypeExpense {
		account.Balance += transaction.Amount
	} else {
		account.Balance -= transaction.Amount
	}
	if err := tx.Save(&account).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
		return
	}

	if err := tx.Delete(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted successfully"})
}

func BatchDeleteTransactions(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	var transactions []models.Transaction
	if err := tx.Where("id IN ? AND user_id = ? AND type != ?", req.IDs, userID, models.TransactionTypeTransfer).
		Find(&transactions).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	accountBalances := make(map[uint]float64)

	for _, t := range transactions {
		if _, exists := accountBalances[t.AccountID]; !exists {
			var account models.Account
			if err := tx.Where("id = ?", t.AccountID).First(&account).Error; err != nil {
				continue
			}
			accountBalances[t.AccountID] = account.Balance
		}

		if t.Type == models.TransactionTypeExpense {
			accountBalances[t.AccountID] += t.Amount
		} else {
			accountBalances[t.AccountID] -= t.Amount
		}
	}

	for accountID, balance := range accountBalances {
		if err := tx.Model(&models.Account{}).Where("id = ?", accountID).Update("balance", balance).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balances"})
			return
		}
	}

	if err := tx.Where("id IN ? AND user_id = ?", req.IDs, userID).Delete(&models.Transaction{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transactions"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Transactions deleted successfully", "count": len(transactions)})
}
