package handlers

import (
	"cointrail/database"
	"cointrail/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateAccountRequest struct {
	Name     string             `json:"name" binding:"required,max=50"`
	Type     models.AccountType `json:"type" binding:"required,oneof=cash bank wechat alipay credit invest other"`
	Balance  float64            `json:"balance"`
	Currency string             `json:"currency"`
	Icon     string             `json:"icon"`
	Note     string             `json:"note"`
}

type UpdateAccountRequest struct {
	Name     string             `json:"name" binding:"omitempty,max=50"`
	Type     models.AccountType `json:"type" binding:"omitempty,oneof=cash bank wechat alipay credit invest other"`
	Balance  float64            `json:"balance"`
	Currency string             `json:"currency"`
	Icon     string             `json:"icon"`
	Note     string             `json:"note"`
}

type TransferRequest struct {
	FromAccountID uint    `json:"from_account_id" binding:"required"`
	ToAccountID   uint    `json:"to_account_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Note          string  `json:"note"`
}

func GetAccounts(c *gin.Context) {
	userID := c.GetUint("user_id")

	var accounts []models.Account
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

func GetAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var account models.Account
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, account)
}

func CreateAccount(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account := models.Account{
		UserID:   userID,
		Name:     req.Name,
		Type:     req.Type,
		Balance:  req.Balance,
		Currency: req.Currency,
		Icon:     req.Icon,
		Note:     req.Note,
	}

	if account.Currency == "" {
		account.Currency = "CNY"
	}

	if err := database.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	c.JSON(http.StatusCreated, account)
}

func UpdateAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var account models.Account
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		account.Name = req.Name
	}
	if req.Type != "" {
		account.Type = req.Type
	}
	account.Balance = req.Balance
	if req.Currency != "" {
		account.Currency = req.Currency
	}
	if req.Icon != "" {
		account.Icon = req.Icon
	}
	if req.Note != "" {
		account.Note = req.Note
	}

	if err := database.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	c.JSON(http.StatusOK, account)
}

func DeleteAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var account models.Account
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&account).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	tx := database.DB.Begin()

	if err := tx.Where("account_id = ? OR target_account_id = ?", id, id).Delete(&models.Transaction{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete related transactions"})
		return
	}

	if err := tx.Delete(&account).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

func Transfer(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.FromAccountID == req.ToAccountID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot transfer to the same account"})
		return
	}

	tx := database.DB.Begin()

	var fromAccount models.Account
	if err := tx.Where("id = ? AND user_id = ?", req.FromAccountID, userID).First(&fromAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Source account not found"})
		return
	}

	var toAccount models.Account
	if err := tx.Where("id = ? AND user_id = ?", req.ToAccountID, userID).First(&toAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Target account not found"})
		return
	}

	if fromAccount.Balance < req.Amount {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	fromAccount.Balance -= req.Amount
	if err := tx.Save(&fromAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update source account"})
		return
	}

	toAccount.Balance += req.Amount
	if err := tx.Save(&toAccount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update target account"})
		return
	}

	transferOut := models.Transaction{
		UserID:          userID,
		AccountID:       req.FromAccountID,
		TargetAccountID: &req.ToAccountID,
		Type:            models.TransactionTypeTransfer,
		Amount:          req.Amount,
		Note:            req.Note,
		TransactionDate: time.Now(),
	}
	if err := tx.Create(&transferOut).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transfer record"})
		return
	}

	transferIn := models.Transaction{
		UserID:          userID,
		AccountID:       req.ToAccountID,
		TargetAccountID: &req.FromAccountID,
		Type:            models.TransactionTypeTransfer,
		Amount:          req.Amount,
		Note:            req.Note,
		TransactionDate: time.Now(),
	}
	if err := tx.Create(&transferIn).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transfer record"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":       "Transfer successful",
		"from_account":  fromAccount,
		"to_account":    toAccount,
	})
}
