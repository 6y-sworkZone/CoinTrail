package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"cointrail/database"
	"cointrail/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func setupTestDB(t *testing.T) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "cointrail_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	if err := database.Init(tmpFile.Name()); err != nil {
		t.Fatalf("Failed to init test database: %v", err)
	}

	testDB = database.DB

	if err := testDB.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.Category{},
		&models.Transaction{},
		&models.Budget{},
	); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	gin.SetMode(gin.TestMode)
}

func teardownTestDB(t *testing.T) {
	t.Helper()
	sqlDB, err := testDB.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.POST("/api/auth/register", Register)
	r.POST("/api/auth/login", Login)
	return r
}

func TestRegister_EmailValidation(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	tests := []struct {
		name       string
		body       map[string]interface{}
		statusCode int
	}{
		{
			name: "invalid email format",
			body: map[string]interface{}{
				"username": "testuser",
				"email":    "invalid-email",
				"password": "password123",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "empty email",
			body: map[string]interface{}{
				"username": "testuser",
				"email":    "",
				"password": "password123",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "valid registration",
			body: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
			},
			statusCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestRegister_PasswordValidation(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	tests := []struct {
		name       string
		password   string
		statusCode int
	}{
		{
			name:       "password too short (5 chars)",
			password:   "12345",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "password too short (empty)",
			password:   "",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "password min length (6 chars)",
			password:   "123456",
			statusCode: http.StatusCreated,
		},
		{
			name:       "password longer than 6 chars",
			password:   "password123",
			statusCode: http.StatusCreated,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"username": "testuser" + fmt.Sprintf("%d", i),
				"email":    "test" + fmt.Sprintf("%d", i) + "@example.com",
				"password": tt.password,
			}
			jsonBody, _ := json.Marshal(body)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	body := map[string]interface{}{
		"username": "testuser1",
		"email":    "duplicate@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req1, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	req2, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusConflict, w2.Code)

	var response map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "already exists")
}

func TestLogin_ReturnsJWT(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	registerBody := map[string]interface{}{
		"username": "loginuser",
		"email":    "login@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var registerResponse AuthResponse
	json.Unmarshal(w.Body.Bytes(), &registerResponse)
	assert.NotEmpty(t, registerResponse.Token)

	loginBody := map[string]interface{}{
		"email":    "login@example.com",
		"password": "password123",
	}
	jsonLoginBody, _ := json.Marshal(loginBody)
	loginReq, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonLoginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	assert.Equal(t, http.StatusOK, loginW.Code)

	var loginResponse AuthResponse
	json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	assert.NotEmpty(t, loginResponse.Token)
	assert.NotEqual(t, registerResponse.Token, "Login should return new token")
	assert.NotNil(t, loginResponse.User)
	assert.Equal(t, "loginuser", loginResponse.User.Username)
	assert.Equal(t, "login@example.com", loginResponse.User.Email)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	registerBody := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	tests := []struct {
		name       string
		email      string
		password   string
		statusCode int
	}{
		{
			name:       "wrong password",
			email:      "test@example.com",
			password:   "wrongpassword",
			statusCode: http.StatusUnauthorized,
		},
		{
			name:       "wrong email",
			email:      "wrong@example.com",
			password:   "password123",
			statusCode: http.StatusUnauthorized,
		},
		{
			name:       "empty email",
			email:      "",
			password:   "password123",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "empty password",
			email:      "test@example.com",
			password:   "",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginBody := map[string]interface{}{
				"email":    tt.email,
				"password": tt.password,
			}
			jsonLoginBody, _ := json.Marshal(loginBody)
			loginReq, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonLoginBody))
			loginReq.Header.Set("Content-Type", "application/json")
			loginW := httptest.NewRecorder()
			router.ServeHTTP(loginW, loginReq)

			assert.Equal(t, tt.statusCode, loginW.Code)
		})
	}
}
