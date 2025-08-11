package api

import (
    "encoding/json"
    "net/http"
    "time"
    "github.com/gorilla/mux"
    "github.com/RhytonBB/vectrik-webbank/internal/db"
)

type NewsRequest struct {
    Title       string        `json:"title"`
    Content     string        `json:"content"`
    Summary     string        `json:"summary"`
    Images      []db.NewsImage `json:"images"`
    IsPublished bool          `json:"is_published"`
    IsFeatured  bool          `json:"is_featured"`
}

// Создание новости
func CreateNews(w http.ResponseWriter, r *http.Request) {
    var req NewsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Валидация
    if req.Title == "" || req.Content == "" {
        http.Error(w, "Title and content are required", http.StatusBadRequest)
        return
    }

    // Конвертируем массив изображений в JSON
    imagesJSON, err := json.Marshal(req.Images)
    if err != nil {
        http.Error(w, "Invalid images format", http.StatusBadRequest)
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
        http.Error(w, "Failed to create news", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(news)
}

// Получение списка новостей (админ)
func GetNewsList(w http.ResponseWriter, r *http.Request) {
    var news []db.News
    if err := db.DB.Order("created_at desc").Find(&news).Error; err != nil {
        http.Error(w, "Failed to fetch news", http.StatusInternalServerError)
        return
    }

    // Парсим изображения для каждой новости
    for i := range news {
        if news[i].Images != "" {
            var images []db.NewsImage
            json.Unmarshal([]byte(news[i].Images), &images)
            // Создаем временную структуру с парсированными изображениями
            imagesJSON, err := json.Marshal(images)
            if err != nil {
                continue // Пропускаем если ошибка парсинга
            }
            news[i].Images = string(imagesJSON)
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(news)
}

// Получение конкретной новости (админ)
func GetNewsById(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    
    var news db.News
    if err := db.DB.First(&news, id).Error; err != nil {
        http.Error(w, "News not found", http.StatusNotFound)
        return
    }

    // Парсим изображения
    if news.Images != "" {
        var images []db.NewsImage
        json.Unmarshal([]byte(news.Images), &images)
        // Создаем временную структуру с парсированными изображениями
        imagesJSON, err := json.Marshal(images)
        if err != nil {
            // Если ошибка парсинга, оставляем оригинальную строку
        } else {
            news.Images = string(imagesJSON)
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(news)
}

// Обновление новости
func UpdateNews(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    
    var req NewsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    var news db.News
    if err := db.DB.First(&news, id).Error; err != nil {
        http.Error(w, "News not found", http.StatusNotFound)
        return
    }

    // Конвертируем массив изображений в JSON
    imagesJSON, err := json.Marshal(req.Images)
    if err != nil {
        http.Error(w, "Invalid images format", http.StatusBadRequest)
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
        http.Error(w, "Failed to update news", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(news)
}

// Публикация/снятие с публикации
func ToggleNewsPublish(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    
    var news db.News
    if err := db.DB.First(&news, id).Error; err != nil {
        http.Error(w, "News not found", http.StatusNotFound)
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
        http.Error(w, "Failed to update news", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(news)
}

// Удаление новости
func DeleteNews(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    
    if err := db.DB.Delete(&db.News{}, id).Error; err != nil {
        http.Error(w, "Failed to delete news", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "News deleted successfully"})
} 