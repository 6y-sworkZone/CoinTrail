package database

import (
	"cointrail/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(dbURL string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbURL), &gorm.Config{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.Category{},
		&models.Transaction{},
		&models.Budget{},
	)
	if err != nil {
		return err
	}

	return nil
}

func InitDefaultCategories(userID uint) error {
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := InitDefaultCategoriesTx(tx, userID); err != nil {
		return err
	}
	return tx.Commit().Error
}

func InitDefaultCategoriesTx(tx *gorm.DB, userID uint) error {
	var count int64
	tx.Model(&models.Category{}).Where("user_id = ?", userID).Count(&count)
	if count > 0 {
		return nil
	}

	expenseCategories := models.GetDefaultExpenseCategories()
	for i := range expenseCategories {
		expenseCategories[i].UserID = userID
		expenseCategories[i].Type = models.CategoryTypeExpense
		expenseCategories[i].IsDefault = true
		if err := tx.Create(&expenseCategories[i]).Error; err != nil {
			return err
		}
	}

	incomeCategories := models.GetDefaultIncomeCategories()
	for i := range incomeCategories {
		incomeCategories[i].UserID = userID
		incomeCategories[i].Type = models.CategoryTypeIncome
		incomeCategories[i].IsDefault = true
		if err := tx.Create(&incomeCategories[i]).Error; err != nil {
			return err
		}
	}

	return nil
}
