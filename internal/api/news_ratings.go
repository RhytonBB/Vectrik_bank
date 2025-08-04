package api

import (
    "encoding/json"
    "net/http"
    "strconv"
    "github.com/gorilla/mux"
    "github.com/RhytonBB/vectrik-webbank/internal/db"
    "errors"
    "github.com/RhytonBB/vectrik-webbank/internal/auth"
)

type RatingRequest struct {
    Rating string `json:"rating"` // thumbs_up, thumbs_down, fire
}

// Оценить новость
func RateNews(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    newsID, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        http.Error(w, "Invalid news ID", http.StatusBadRequest)
        return
    }

    // Получаем пользователя из сессии
    user, err := getCurrentUser(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var req RatingRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Проверяем валидность рейтинга
    if req.Rating != "thumbs_up" && req.Rating != "thumbs_down" && req.Rating != "fire" {
        http.Error(w, "Invalid rating type", http.StatusBadRequest)
        return
    }

    // Проверяем, что новость существует и опубликована
    var news db.News
    if err := db.DB.Where("id = ? AND is_published = ?", newsID, true).First(&news).Error; err != nil {
        http.Error(w, "News not found", http.StatusNotFound)
        return
    }

    // Проверяем, не оценивал ли уже пользователь эту новость
    var existingRating db.NewsRating
    if err := db.DB.Where("news_id = ? AND user_id = ?", newsID, user.ID).First(&existingRating).Error; err == nil {
        // Пользователь уже оценивал - обновляем рейтинг
        if existingRating.Rating == req.Rating {
            // Убираем рейтинг (двойной клик)
            db.DB.Delete(&existingRating)
            updateNewsRatingCounts(uint(newsID), existingRating.Rating, -1)
        } else {
            // Меняем рейтинг
            oldRating := existingRating.Rating
            existingRating.Rating = req.Rating
            db.DB.Save(&existingRating)
            updateNewsRatingCounts(uint(newsID), oldRating, -1)
            updateNewsRatingCounts(uint(newsID), req.Rating, 1)
        }
    } else {
        // Создаем новый рейтинг
        rating := db.NewsRating{
            NewsID: uint(newsID),
            UserID: user.ID,
            Rating: req.Rating,
        }
        db.DB.Create(&rating)
        updateNewsRatingCounts(uint(newsID), req.Rating, 1)
    }

    // Возвращаем обновленные данные новости
    var updatedNews db.News
    db.DB.First(&updatedNews, newsID)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Rating updated successfully",
        "news": map[string]interface{}{
            "id":           updatedNews.ID,
            "thumbs_up":    updatedNews.ThumbsUp,
            "thumbs_down":  updatedNews.ThumbsDown,
            "fire_count":   updatedNews.FireCount,
        },
    })
}

// Получить рейтинг пользователя для новости
func GetUserRating(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    newsID, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        http.Error(w, "Invalid news ID", http.StatusBadRequest)
        return
    }

    user, err := getCurrentUser(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var rating db.NewsRating
    if err := db.DB.Where("news_id = ? AND user_id = ?", newsID, user.ID).First(&rating).Error; err != nil {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "rating": nil,
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "rating": rating.Rating,
    })
}

// Обновляет счетчики рейтингов новости
func updateNewsRatingCounts(newsID uint, ratingType string, delta int) {
    var field string
    switch ratingType {
    case "thumbs_up":
        field = "thumbs_up"
    case "thumbs_down":
        field = "thumbs_down"
    case "fire":
        field = "fire_count"
    default:
        return
    }

    db.DB.Exec("UPDATE news SET "+field+" = "+field+" + ? WHERE id = ?", delta, newsID)
}

// Вспомогательная функция для получения текущего пользователя
func getCurrentUser(r *http.Request) (*db.User, error) {
    cookie, err := r.Cookie("token")
    if err != nil {
        return nil, err
    }
    
    claims, err := auth.ParseJWT(cookie.Value)
    if err != nil {
        return nil, err
    }
    
    var userID uint
    switch v := claims["user_id"].(type) {
    case float64:
        userID = uint(v)
    case int:
        userID = uint(v)
    case uint:
        userID = v
    default:
        return nil, errors.New("invalid user_id type")
    }
    
    var user db.User
    if err := db.DB.First(&user, userID).Error; err != nil {
        return nil, err
    }
    
    return &user, nil
} 