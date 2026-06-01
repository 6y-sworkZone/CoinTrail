package handlers

import (
	"cointrail/database"
	"cointrail/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateCategoryRequest struct {
	Name  string               `json:"name" binding:"required,max=50"`
	Type  models.CategoryType  `json:"type" binding:"required,oneof=expense income"`
	Icon  string               `json:"icon"`
	Color string               `json:"color"`
	Sort  int                  `json:"sort"`
}

type UpdateCategoryRequest struct {
	Name  string               `json:"name" binding:"omitempty,max=50"`
	Type  models.CategoryType  `json:"type" binding:"omitempty,oneof=expense income"`
	Icon  string               `json:"icon"`
	Color string               `json:"color"`
	Sort  int                  `json:"sort"`
}

func GetCategories(c *gin.Context) {
	userID := c.GetUint("user_id")
	categoryType := c.Query("type")

	var categories []models.Category
	query := database.DB.Where("user_id = ?", userID)

	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}

	if err := query.Order("sort ASC, created_at DESC").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func GetCategory(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var category models.Category
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func CreateCategory(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := models.Category{
		UserID:    userID,
		Name:      req.Name,
		Type:      req.Type,
		Icon:      req.Icon,
		Color:     req.Color,
		Sort:      req.Sort,
		IsDefault: false,
	}

	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func UpdateCategory(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var category models.Category
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Type != "" {
		category.Type = req.Type
	}
	if req.Icon != "" {
		category.Icon = req.Icon
	}
	if req.Color != "" {
		category.Color = req.Color
	}
	category.Sort = req.Sort

	if err := database.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func DeleteCategory(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	var category models.Category
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if category.IsDefault {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete default category"})
		return
	}

	tx := database.DB.Begin()

	if err := tx.Model(&models.Transaction{}).Where("category_id = ?", id).Update("category_id", nil).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlink transactions"})
		return
	}

	if err := tx.Where("category_id = ?", id).Delete(&models.Budget{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete related budgets"})
		return
	}

	if err := tx.Delete(&category).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
