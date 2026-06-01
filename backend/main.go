package main

import (
	"cointrail/config"
	"cointrail/database"
	"cointrail/handlers"
	"cointrail/middleware"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := database.Init(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.PATCH("/password", middleware.AuthMiddleware(), handlers.ChangePassword)
			auth.GET("/profile", middleware.AuthMiddleware(), handlers.GetProfile)
		}

		accounts := api.Group("/accounts")
		accounts.Use(middleware.AuthMiddleware())
		{
			accounts.GET("", handlers.GetAccounts)
			accounts.GET("/:id", handlers.GetAccount)
			accounts.POST("", handlers.CreateAccount)
			accounts.PATCH("/:id", handlers.UpdateAccount)
			accounts.DELETE("/:id", handlers.DeleteAccount)
			accounts.POST("/transfer", handlers.Transfer)
		}

		categories := api.Group("/categories")
		categories.Use(middleware.AuthMiddleware())
		{
			categories.GET("", handlers.GetCategories)
			categories.GET("/:id", handlers.GetCategory)
			categories.POST("", handlers.CreateCategory)
			categories.PATCH("/:id", handlers.UpdateCategory)
			categories.DELETE("/:id", handlers.DeleteCategory)
		}

		transactions := api.Group("/transactions")
		transactions.Use(middleware.AuthMiddleware())
		{
			transactions.GET("", handlers.GetTransactions)
			transactions.GET("/:id", handlers.GetTransaction)
			transactions.POST("", handlers.CreateTransaction)
			transactions.PATCH("/:id", handlers.UpdateTransaction)
			transactions.DELETE("/:id", handlers.DeleteTransaction)
			transactions.DELETE("/batch", handlers.BatchDeleteTransactions)
		}

		budgets := api.Group("/budgets")
		budgets.Use(middleware.AuthMiddleware())
		{
			budgets.GET("", handlers.GetBudgets)
			budgets.GET("/:id", handlers.GetBudget)
			budgets.POST("", handlers.CreateBudget)
			budgets.PATCH("/:id", handlers.UpdateBudget)
			budgets.DELETE("/:id", handlers.DeleteBudget)
		}

		stats := api.Group("/stats")
		stats.Use(middleware.AuthMiddleware())
		{
			stats.GET("/dashboard", handlers.GetDashboard)
			stats.GET("/monthly", handlers.GetMonthlySummary)
			stats.GET("/category", handlers.GetCategorySummary)
			stats.GET("/trend", handlers.GetTrendStats)
			stats.GET("/dashboard-budgets", handlers.GetDashboardBudgets)
		}

		ie := api.Group("/io")
		ie.Use(middleware.AuthMiddleware())
		{
			ie.GET("/export", handlers.ExportTransactions)
			ie.POST("/import", handlers.ImportTransactions)
		}
	}

	log.Printf("Server starting on port %s...", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
