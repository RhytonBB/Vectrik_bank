package api

import (
    "encoding/json"
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/RhytonBB/vectrik-webbank/internal/db"
)

type NewsResponse struct {
    ID          uint     `json:"id"`
    Title       string   `json:"title"`
    Content     string   `json:"content"`
    Summary     string   `json:"summary"`
    Images      []string `json:"images"`
    IsFeatured  bool     `json:"is_featured"`
    Views       int      `json:"views"`
    PublishedAt *string  `json:"published_at"`
    CreatedAt   string   `json:"created_at"`
}

// Получение опубликованных новостей
func GetPublishedNews(c *gin.Context) {
    var news []db.News
    if err := db.DB.Where("is_published = ?", true).Order("is_featured desc, published_at desc").Find(&news).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch news"})
        return
    }

    var response []NewsResponse
    for _, item := range news {
        var images []string
        if item.Images != "" {
            json.Unmarshal([]byte(item.Images), &images)
        }

        publishedAt := ""
        if item.PublishedAt != nil {
            publishedAt = item.PublishedAt.Format("2006-01-02T15:04:05Z")
        }

        response = append(response, NewsResponse{
            ID:          item.ID,
            Title:       item.Title,
            Content:     item.Content,
            Summary:     item.Summary,
            Images:      images,
            IsFeatured:  item.IsFeatured,
            Views:       item.Views,
            PublishedAt: &publishedAt,
            CreatedAt:   item.CreatedAt.Format("2006-01-02T15:04:05Z"),
        })
    }

    c.JSON(http.StatusOK, response)
}

// Получение конкретной новости
func GetNewsById(c *gin.Context) {
    id := c.Param("id")
    var news db.News
    
    if err := db.DB.Where("id = ? AND is_published = ?", id, true).First(&news).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "News not found"})
        return
    }

    // Увеличиваем счетчик просмотров
    db.DB.Model(&news).Update("views", news.Views+1)

    var images []string
    if news.Images != "" {
        json.Unmarshal([]byte(news.Images), &images)
    }

    publishedAt := ""
    if news.PublishedAt != nil {
        publishedAt = news.PublishedAt.Format("2006-01-02T15:04:05Z")
    }

    response := NewsResponse{
        ID:          news.ID,
        Title:       news.Title,
        Content:     news.Content,
        Summary:     news.Summary,
        Images:      images,
        IsFeatured:  news.IsFeatured,
        Views:       news.Views + 1, // Показываем обновленное значение
        PublishedAt: &publishedAt,
        CreatedAt:   news.CreatedAt.Format("2006-01-02T15:04:05Z"),
    }

    c.JSON(http.StatusOK, response)
}

// Получение последних новостей (для превью)
func GetLatestNews(c *gin.Context) {
    limit := 3
    if limitStr := c.Query("limit"); limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 10 {
            limit = l
        }
    }

    var news []db.News
    if err := db.DB.Where("is_published = ?", true).Order("is_featured desc, published_at desc").Limit(limit).Find(&news).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch news"})
        return
    }

    var response []NewsResponse
    for _, item := range news {
        var images []string
        if item.Images != "" {
            json.Unmarshal([]byte(item.Images), &images)
        }

        publishedAt := ""
        if item.PublishedAt != nil {
            publishedAt = item.PublishedAt.Format("2006-01-02T15:04:05Z")
        }

        response = append(response, NewsResponse{
            ID:          item.ID,
            Title:       item.Title,
            Content:     item.Content,
            Summary:     item.Summary,
            Images:      images,
            IsFeatured:  item.IsFeatured,
            Views:       item.Views,
            PublishedAt: &publishedAt,
            CreatedAt:   item.CreatedAt.Format("2006-01-02T15:04:05Z"),
        })
    }

    c.JSON(http.StatusOK, response)
} 