package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"strconv"
	"time"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/gorilla/mux"
)

// API: Создать broadcast сообщение
func AdminBroadcastAPIHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	isAdmin, ok := claims["is_admin"].(bool)
	if !ok || !isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var adminID uint
	switch v := claims["user_id"].(type) {
	case float64:
		adminID = uint(v)
	case int:
		adminID = uint(v)
	case uint:
		adminID = v
	default:
		http.Error(w, "Invalid admin ID", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Text          string   `json:"text"`
			MediaURL      *string  `json:"media_url"`
			MediaType     *string  `json:"media_type"`
			TargetType    string   `json:"target_type"`
			TargetUserIDs []uint   `json:"target_user_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Text == "" {
			http.Error(w, "Текст сообщения обязателен", http.StatusBadRequest)
			return
		}

		if req.TargetType != "all" && req.TargetType != "registered" && req.TargetType != "manual" {
			http.Error(w, "Некорректный тип целевой аудитории", http.StatusBadRequest)
			return
		}

		if req.TargetType == "manual" && len(req.TargetUserIDs) == 0 {
			http.Error(w, "Выберите хотя бы одного пользователя", http.StatusBadRequest)
			return
		}

		// Сериализуем массив ID в JSON строку
		targetUserIDsJSON, err := json.Marshal(req.TargetUserIDs)
		if err != nil {
			http.Error(w, "Ошибка обработки списка пользователей", http.StatusInternalServerError)
			return
		}

		broadcast := db.BroadcastMessage{
			AdminID:       adminID,
			Text:          req.Text,
			MediaURL:      req.MediaURL,
			MediaType:     req.MediaType,
			TargetType:    req.TargetType,
			TargetUserIDs: string(targetUserIDsJSON),
			Status:        "pending",
			CreatedAt:     time.Now(),
		}

		if err := db.DB.Create(&broadcast).Error; err != nil {
			http.Error(w, "Ошибка создания сообщения: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Запускаем отправку в фоне
		go sendBroadcastMessage(broadcast.ID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(broadcast)
		return
	}

	if r.Method == http.MethodGet {
		var broadcasts []db.BroadcastMessage
		if err := db.DB.Preload("Admin").Order("created_at desc").Find(&broadcasts).Error; err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(broadcasts)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// API: Получить историю отправки конкретного сообщения
func AdminBroadcastHistoryAPIHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	isAdmin, ok := claims["is_admin"].(bool)
	if !ok || !isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	broadcastIDStr := vars["id"]
	broadcastID, err := strconv.Atoi(broadcastIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	var history []db.BroadcastHistory
	if err := db.DB.Preload("User").Where("broadcast_id = ?", broadcastID).Order("sent_at desc").Find(&history).Error; err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// API: Получить список пользователей для выбора
func AdminBroadcastUsersAPIHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	isAdmin, ok := claims["is_admin"].(bool)
	if !ok || !isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	targetType := r.URL.Query().Get("type")
	var users []db.User

	query := db.DB.Where("telegram_id IS NOT NULL")
	
	switch targetType {
	case "registered":
		query = query.Where("is_registered = ?", true)
	case "all":
		// Все пользователи с telegram_id
		break
	default:
		http.Error(w, "Некорректный тип", http.StatusBadRequest)
		return
	}

	if err := query.Order("first_name").Find(&users).Error; err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Функция отправки сообщения через Telegram-бота
func sendBroadcastMessage(broadcastID uint) {
	var broadcast db.BroadcastMessage
	if err := db.DB.Preload("Admin").First(&broadcast, broadcastID).Error; err != nil {
		return
	}

	// Обновляем статус на "sending"
	db.DB.Model(&broadcast).Update("status", "sending")

	// Получаем список пользователей для отправки
	var users []db.User
	query := db.DB.Where("telegram_id IS NOT NULL")
	
	switch broadcast.TargetType {
	case "all":
		// Все пользователи с telegram_id
		break
	case "registered":
		query = query.Where("is_registered = ?", true)
	case "manual":
		// Десериализуем JSON строку обратно в массив
		var targetUserIDs []uint
		if broadcast.TargetUserIDs != "" {
			if err := json.Unmarshal([]byte(broadcast.TargetUserIDs), &targetUserIDs); err != nil {
				db.DB.Model(&broadcast).Updates(map[string]interface{}{
					"status":        "failed",
					"completed_at":  time.Now(),
				})
				return
			}
		}
		
		if len(targetUserIDs) > 0 {
			query = query.Where("id IN ?", targetUserIDs)
		} else {
			db.DB.Model(&broadcast).Updates(map[string]interface{}{
				"status":        "completed",
				"completed_at":  time.Now(),
			})
			return
		}
	}

	if err := query.Find(&users).Error; err != nil {
		db.DB.Model(&broadcast).Updates(map[string]interface{}{
			"status":        "failed",
			"completed_at":  time.Now(),
		})
		return
	}

	// Отправляем сообщения через Telegram-бота
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		db.DB.Model(&broadcast).Updates(map[string]interface{}{
			"status":        "failed",
			"completed_at":  time.Now(),
		})
		return
	}

	sentCount := 0
	failedCount := 0

	for _, user := range users {
		if user.TelegramID == nil {
			continue
		}

		success := sendTelegramMessage(botToken, *user.TelegramID, broadcast.Text, broadcast.MediaURL, broadcast.MediaType)
		
		history := db.BroadcastHistory{
			BroadcastID: broadcastID,
			UserID:      user.ID,
			TelegramID:  *user.TelegramID,
			Status:      "sent",
			SentAt:      time.Now(),
		}

		if !success {
			history.Status = "failed"
			errorMsg := "Ошибка отправки"
			history.ErrorMessage = &errorMsg
			failedCount++
		} else {
			sentCount++
		}

		db.DB.Create(&history)
	}

	// Обновляем статистику
	db.DB.Model(&broadcast).Updates(map[string]interface{}{
		"sent_count":    sentCount,
		"failed_count":  failedCount,
		"status":        "completed",
		"completed_at":  time.Now(),
	})
}

// Функция отправки сообщения в Telegram
func sendTelegramMessage(botToken string, chatID int64, text string, mediaURL *string, mediaType *string) bool {
	// Если есть медиа, отправляем его с текстом в caption
	if mediaURL != nil && mediaType != nil {
		// Получаем путь к файлу на диске
		filePath := strings.TrimPrefix(*mediaURL, "/")
		if !strings.HasPrefix(filePath, "internal/media/") {
			filePath = "internal/media/" + filePath
		}
		
		// Проверяем существование файла
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return false
		}
		
		// Открываем файл
		file, err := os.Open(filePath)
		if err != nil {
			return false
		}
		defer file.Close()
		
		// Определяем метод API в зависимости от типа медиа
		var method string
		var fieldName string
		switch *mediaType {
		case "image":
			method = "sendPhoto"
			fieldName = "photo"
		case "video":
			method = "sendVideo"
			fieldName = "video"
		case "document":
			method = "sendDocument"
			fieldName = "document"
		default:
			// Если неизвестный тип, отправляем только текст
			if text != "" {
				return sendTextOnly(botToken, chatID, text)
			}
			return true
		}
		
		// Создаем multipart запрос
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		
		// Добавляем chat_id
		writer.WriteField("chat_id", fmt.Sprintf("%d", chatID))
		
		// Добавляем caption (текст) если есть
		if text != "" {
			writer.WriteField("caption", text)
			writer.WriteField("parse_mode", "HTML")
		}
		
		// Добавляем файл
		part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
		if err != nil {
			return false
		}
		
		_, err = io.Copy(part, file)
		if err != nil {
			return false
		}
		
		writer.Close()
		
		// Отправляем запрос
		url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, method)
		resp, err := http.Post(url, writer.FormDataContentType(), &buf)
		if err != nil || resp.StatusCode != 200 {
			return false
		}
		resp.Body.Close()
		
		return true
	}
	
	// Если медиа нет, отправляем только текст
	if text != "" {
		return sendTextOnly(botToken, chatID, text)
	}
	
	return true
}

// Вспомогательная функция для отправки только текста
func sendTextOnly(botToken string, chatID int64, text string) bool {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	data := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
		"parse_mode": "HTML",
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false
	}
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	resp.Body.Close()
	
	return true
} 