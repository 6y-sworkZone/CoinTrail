package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cointrail/database"
	"cointrail/middleware"
	"cointrail/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTransactionTestDB(t *testing.T) (uint, string) {
	t.Helper()
	setupTestDB(t)

	user := models.User{
		Username: "txuser",
		Email:    "tx@example.com",
	}
	user.HashPassword("password123")
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	if err := database.InitDefaultCategories(user.ID); err != nil {
		t.Fatalf("Failed to init default categories: %v", err)
	}

	authToken := generateToken(user.ID)

	return user.ID, authToken
}

func createTestAccount(t *testing.T, userID uint, name string, balance float64) models.Account {
	t.Helper()
	account := models.Account{
		UserID:   userID,
		Name:     name,
		Type:     models.AccountTypeCash,
		Balance:  balance,
		Currency: "CNY",
	}
	if err := database.DB.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}
	return account
}

func getExpenseCategory(t *testing.T, userID uint) models.Category {
	t.Helper()
	var category models.Category
	if err := database.DB.Where("user_id = ? AND type = ?", userID, models.CategoryTypeExpense).First(&category).Error; err != nil {
		t.Fatalf("Failed to get expense category: %v", err)
	}
	return category
}

func authMiddlewareMock(userID uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

func setupTransactionRouter(userID uint) *gin.Engine {
	r := gin.New()
	r.Use(authMiddlewareMock(userID))
	r.POST("/api/transactions", CreateTransaction)
	r.POST("/api/accounts/transfer", Transfer)
	r.GET("/api/accounts", GetAccounts)
	return r
}

func TestCreateTransaction_Expense_BalanceDeducted(t *testing.T) {
	userID, _ := setupTransactionTestDB(t)
	defer teardownTestDB(t)

	router := setupTransactionRouter(userID)

	account := createTestAccount(t, userID, "现金", 1000.00)
	category := getExpenseCategory(t, userID)

	body := map[string]interface{}{
		"account_id":       account.ID,
		"category_id":      category.ID,
		"type":             models.TransactionTypeExpense,
		"amount":           150.50,
		"note":             "午餐",
		"transaction_date": "2024-01-15",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var transaction models.Transaction
	json.Unmarshal(w.Body.Bytes(), &transaction)
	assert.Equal(t, models.TransactionTypeExpense, transaction.Type)
	assert.Equal(t, 150.50, transaction.Amount)
	assert.Equal(t, account.ID, transaction.AccountID)
	assert.Equal(t, category.ID, *transaction.CategoryID)

	var updatedAccount models.Account
	database.DB.First(&updatedAccount, account.ID)
	assert.Equal(t, 849.50, updatedAccount.Balance, "Account balance should be 1000 - 150.50 = 849.50")
}

func TestCreateTransaction_Income_BalanceAdded(t *testing.T) {
	userID, _ := setupTransactionTestDB(t)
	defer teardownTestDB(t)

	router := setupTransactionRouter(userID)

	account := createTestAccount(t, userID, "银行卡", 2000.00)

	var incomeCategory models.Category
	database.DB.Where("user_id = ? AND type = ?", userID, models.CategoryTypeIncome).First(&incomeCategory)

	body := map[string]interface{}{
		"account_id":       account.ID,
		"category_id":      incomeCategory.ID,
		"type":             models.TransactionTypeIncome,
		"amount":           5000.00,
		"note":             "工资",
		"transaction_date": "2024-01-10",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var updatedAccount models.Account
	database.DB.First(&updatedAccount, account.ID)
	assert.Equal(t, 7000.00, updatedAccount.Balance, "Account balance should be 2000 + 5000 = 7000")
}

func TestCreateTransaction_InvalidAccount(t *testing.T) {
	userID, _ := setupTransactionTestDB(t)
	defer teardownTestDB(t)

	router := setupTransactionRouter(userID)

	category := getExpenseCategory(t, userID)

	body := map[string]interface{}{
		"account_id":       99999,
		"category_id":      category.ID,
		"type":             models.TransactionTypeExpense,
		"amount":           100.00,
		"transaction_date": "2024-01-15",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateTransaction_InvalidAmount(t *testing.T) {
	userID, _ := setupTransactionTestDB(t)
	defer teardownTestDB(t)

	router := setupTransactionRouter(userID)

	account := createTestAccount(t, userID, "现金", 1000.00)
	category := getExpenseCategory(t, userID)

	tests := []struct {
		name       string
		amount     float64
		statusCode int
	}{
		{"zero amount", 0, http.StatusBadRequest},
		{"negative amount", -50, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"account_id":       account.ID,
				"category_id":      category.ID,
				"type":             models.TransactionTypeExpense,
				"amount":           tt.amount,
				"transaction_date": "2024-01-15",
			}
			jsonBody, _ := json.Marshal(body)
			req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestTransfer_TwoRecordsCreated(t *testing.T) {
	userID, _ := setupTransactionTestDB(t)
	defer teardownTestDB(t)

	router := setupTransactionRouter(userID)

	fromAccount := createTestAccount(t, userID, "现金", 1000.00)
	toAccount := createTestAccount(t, userID, "银行卡", 2000.00)

	body := map[string]interface{}{
		"from_account_id": fromAccount.ID,
		"to_account_id":   toAccount.ID,
		"amount":          300.00,
		"note":            "转账测试",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/accounts/transfer", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var transferCount int64
	database.DB.Model(&models.Transaction{}).
		Where("type = ? AND user_id = ?", models.TransactionTypeTransfer, userID).
		Count(&transferCount)
	assert.Equal(t, int64(2), transferCount, "Should create 2 transfer records")

	var transferRecords []models.Transaction
	database.DB.Where("type = ? AND user_id = ?", models.TransactionTypeTransfer, userID).
		Order("id ASC").
		Find(&transferRecords)

	assert.Len(t, transferRecords, 2)

	transferOut := transferRecords[0]
	transferIn := transferRecords[1]

	assert.Equal(t, fromAccount.ID, transferOut.AccountID)
	assert.Equal(t, toAccount.ID, *transferOut.TargetAccountID)
	assert.Equal(t, 300.00, transferOut.Amount)
	assert.Equal(t, "转账测试", transferOut.Note)

	assert.Equal(t, toAccount.ID, transferIn.AccountID)
	assert.Equal(t, fromAccount.ID, *transferIn.TargetAccountID)
	assert.Equal(t, 300.00, transferIn.Amount)
	assert.Equal(t, "转账测试", transferIn.Note)
}

func TestTransfer_BalanceUpdated(t *testing.T) {
	userID, _ := setupTransactionTestDB(t)
	defer teardownTestDB(t)

	router := setupTransactionRouter(userID)

	fromAccount := createTestAccount(t, userID, "现金", 1000.00)
	toAccount := createTestAccount(t, userID, "银行卡", 2000.00)

	body := map[string]interface{}{
		"from_account_id": fromAccount.ID,
		"to_account_id":   toAccount.ID,
		"amount":          250.75,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/accounts/transfer", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedFromAccount models.Account
	database.DB.First(&updatedFromAccount, fromAccount.ID)
	assert.Equal(t, 749.25, updatedFromAccount.Balance, "Source account should be 1000 - 250.75 = 749.25")

	var updatedToAccount models.Account
	database.DB.First(&updatedToAccount, toAccount.ID)
	assert.Equal(t, 2250.75, updatedToAccount.Balance, "Target account should be 2000 + 250.75 = 2250.75")
}

func TestTransfer_InsufficientBalance(t *testing.T) {
	userID, _ := setupTransactionTestDB(t)
	defer teardownTestDB(t)

	router := setupTransactionRouter(userID)

	fromAccount := createTestAccount(t, userID, "现金", 100.00)
	toAccount := createTestAccount(t, userID, "银行卡", 500.00)

	body := map[string]interface{}{
		"from_account_id": fromAccount.ID,
		"to_account_id":   toAccount.ID,
		"amount":          500.00,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/accounts/transfer", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var updatedFromAccount models.Account
	database.DB.First(&updatedFromAccount, fromAccount.ID)
	assert.Equal(t, 100.00, updatedFromAccount.Balance, "Balance should not change on failed transfer")

	var updatedToAccount models.Account
	database.DB.First(&updatedToAccount, toAccount.ID)
	assert.Equal(t, 500.00, updatedToAccount.Balance, "Balance should not change on failed transfer")

	var transferCount int64
	database.DB.Model(&models.Transaction{}).
		Where("type = ? AND user_id = ?", models.TransactionTypeTransfer, userID).
		Count(&transferCount)
	assert.Equal(t, int64(0), transferCount, "Should not create transfer records on failure")
}

func TestTransfer_SameAccount(t *testing.T) {
	userID, _ := setupTransactionTestDB(t)
	defer teardownTestDB(t)

	router := setupTransactionRouter(userID)

	account := createTestAccount(t, userID, "现金", 1000.00)

	body := map[string]interface{}{
		"from_account_id": account.ID,
		"to_account_id":   account.ID,
		"amount":          100.00,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/accounts/transfer", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransfer_InvalidAccount(t *testing.T) {
	userID, _ := setupTransactionTestDB(t)
	defer teardownTestDB(t)

	router := setupTransactionRouter(userID)

	fromAccount := createTestAccount(t, userID, "现金", 1000.00)

	tests := []struct {
		name            string
		fromAccountID   uint
		toAccountID     uint
		statusCode      int
	}{
		{
			name:            "invalid source account",
			fromAccountID:   99999,
			toAccountID:     fromAccount.ID,
			statusCode:      http.StatusNotFound,
		},
		{
			name:            "invalid target account",
			fromAccountID:   fromAccount.ID,
			toAccountID:     99999,
			statusCode:      http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"from_account_id": tt.fromAccountID,
				"to_account_id":   tt.toAccountID,
				"amount":          100.00,
			}
			jsonBody, _ := json.Marshal(body)
			req, _ := http.NewRequest("POST", "/api/accounts/transfer", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	userID, authToken := setupTransactionTestDB(t)
	defer teardownTestDB(t)

	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/api/test", func(c *gin.Context) {
		uid := c.GetUint("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": uid})
	})

	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+authToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(userID), response["user_id"])
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	setupTransactionTestDB(t)
	defer teardownTestDB(t)

	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	setupTransactionTestDB(t)
	defer teardownTestDB(t)

	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/api/test", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
