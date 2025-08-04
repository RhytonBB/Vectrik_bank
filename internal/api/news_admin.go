package api

import (
    "encoding/json"
    "net/http"
    "strconv"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/RhytonBB/vectrik-webbank/internal/db"
)

type NewsRequest struct {
    Title       string   `json:"title" binding:"required"`
    Content     string   `json:"content" binding:"required"`
    Summary     string   `json:"summary"`
    Images      []string `json:"images"`
    IsPublished bool     `json:"is_published"`
    IsFeatured  bool     `json:"is_featured"`
}

// Создание новости
func CreateNews(c *gin.Context) {
    var req NewsRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Конвертируем массив изображений в JSON
    imagesJSON, err := json.Marshal(req.Images)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid images format"})
        return
    }

    news := db.News{
        Title:       req.Title,
        Content:     req.Content,
        Summary:     req.Summary,
        Images:      string(imagesJSON),
        IsPublished: req.IsPublished,
        IsFeatured:  req.IsFeatured,
    }

    if req.IsPublished {
        now := time.Now()
        news.PublishedAt = &now
    }

    if err := db.DB.Create(&news).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create news"})
        return
    }

    c.JSON(http.StatusCreated, news)
}

// Получение списка новостей (админ)
func GetNewsList(c *gin.Context) {
    var news []db.News
    if err := db.DB.Order("created_at desc").Find(&news).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch news"})
        return
    }

    c.JSON(http.StatusOK, news)
}

// Получение конкретной новости (админ)
func GetNewsById(c *gin.Context) {
    id := c.Param("id")
    var news db.News
    
    if err := db.DB.First(&news, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "News not found"})
        return
    }

    c.JSON(http.StatusOK, news)
}

// Обновление новости
func UpdateNews(c *gin.Context) {
    id := c.Param("id")
    var req NewsRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var news db.News
    if err := db.DB.First(&news, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "News not found"})
        return
    }

    // Конвертируем массив изображений в JSON
    imagesJSON, err := json.Marshal(req.Images)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid images format"})
        return
    }

    news.Title = req.Title
    news.Content = req.Content
    news.Summary = req.Summary
    news.Images = string(imagesJSON)
    news.IsFeatured = req.IsFeatured

    // Обработка публикации
    if req.IsPublished && !news.IsPublished {
        now := time.Now()
        news.PublishedAt = &now
    } else if !req.IsPublished {
        news.PublishedAt = nil
    }
    news.IsPublished = req.IsPublished

    if err := db.DB.Save(&news).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update news"})
        return
    }

    c.JSON(http.StatusOK, news)
}

// Публикация/снятие с публикации
func ToggleNewsPublish(c *gin.Context) {
    id := c.Param("id")
    var news db.News
    
    if err := db.DB.First(&news, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "News not found"})
        return
    }

    news.IsPublished = !news.IsPublished
    if news.IsPublished {
        now := time.Now()
        news.PublishedAt = &now
    } else {
        news.PublishedAt = nil
    }

    if err := db.DB.Save(&news).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update news"})
        return
    }

    c.JSON(http.StatusOK, news)
}

// Удаление новости
func DeleteNews(c *gin.Context) {
    id := c.Param("id")
    if err := db.DB.Delete(&db.News{}, id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete news"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "News deleted successfully"})
} 