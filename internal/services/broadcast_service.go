package services

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"fmt"
	"bytes"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
)

var botToken string

func InitBotToken(token string) {
	botToken = token
}

func CreateBroadcastMessage(adminID uint, text string, mediaURL, mediaType *string, targetType string, targetUserIDs []uint) (*db.BroadcastMessage, error) {
	if text == "" || targetType == "" {
		return nil, errors.New("Текст и тип аудитории обязательны")
	}
	if targetType != "all" && targetType != "registered" && targetType != "manual" {
		return nil, errors.New("Некорректный тип целевой аудитории")
	}
	if targetType == "manual" && len(targetUserIDs) == 0 {
		return nil, errors.New("Выберите хотя бы одного пользователя")
	}
	targetUserIDsJSON, err := json.Marshal(targetUserIDs)
	if err != nil {
		return nil, errors.New("Ошибка обработки списка пользователей")
	}
	broadcast := db.BroadcastMessage{
		AdminID:       adminID,
		Text:          text,
		MediaURL:      mediaURL,
		MediaType:     mediaType,
		TargetType:    targetType,
		TargetUserIDs: string(targetUserIDsJSON),
		Status:        "pending",
		CreatedAt:     time.Now(),
	}
	if err := db.DB.Create(&broadcast).Error; err != nil {
		return nil, err
	}
	go sendBroadcastMessage(broadcast.ID)
	return &broadcast, nil
}

func sendBroadcastMessage(broadcastID uint) {
	var broadcast db.BroadcastMessage
	if err := db.DB.Preload("Admin").First(&broadcast, broadcastID).Error; err != nil {
		return
	}
	db.DB.Model(&broadcast).Update("status", "sending")
	var users []db.User
	query := db.DB.Where("telegram_id IS NOT NULL")
	switch broadcast.TargetType {
	case "all":
		// все пользователи с telegram_id
	case "registered":
		query = query.Where("is_registered = ?", true)
	case "manual":
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
	db.DB.Model(&broadcast).Updates(map[string]interface{}{
		"sent_count":    sentCount,
		"failed_count":  failedCount,
		"status":        "completed",
		"completed_at":  time.Now(),
	})
}

func sendTelegramMessage(botToken string, chatID int64, text string, mediaURL *string, mediaType *string) bool {
	if mediaURL != nil && mediaType != nil {
		filePath := strings.TrimPrefix(*mediaURL, "/")
		if !strings.HasPrefix(filePath, "internal/media/") {
			filePath = "internal/media/" + filePath
		}
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return false
		}
		file, err := os.Open(filePath)
		if err != nil {
			return false
		}
		defer file.Close()
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
			if text != "" {
				return sendTextOnly(botToken, chatID, text)
			}
			return true
		}
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		writer.WriteField("chat_id", fmt.Sprintf("%d", chatID))
		if text != "" {
			writer.WriteField("caption", text)
			writer.WriteField("parse_mode", "HTML")
		}
		part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
		if err != nil {
			return false
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return false
		}
		writer.Close()
		url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, method)
		resp, err := http.Post(url, writer.FormDataContentType(), &buf)
		if err != nil || resp.StatusCode != 200 {
			return false
		}
		resp.Body.Close()
		return true
	}
	if text != "" {
		return sendTextOnly(botToken, chatID, text)
	}
	return true
}

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

func GetBroadcastByID(broadcastID uint) (*db.BroadcastMessage, error) {
	var broadcast db.BroadcastMessage
	if err := db.DB.First(&broadcast, broadcastID).Error; err != nil {
		return nil, err
	}
	return &broadcast, nil
}

func GetAllBroadcasts() ([]db.BroadcastMessage, error) {
	var broadcasts []db.BroadcastMessage
	if err := db.DB.Order("created_at desc").Find(&broadcasts).Error; err != nil {
		return nil, err
	}
	return broadcasts, nil
}

func GetBroadcastHistory(broadcastID uint) ([]db.BroadcastHistory, error) {
	var history []db.BroadcastHistory
	if err := db.DB.Preload("User").Where("broadcast_id = ?", broadcastID).Order("sent_at desc").Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

func GetBroadcastUsers(targetType string) ([]db.User, error) {
	var users []db.User
	query := db.DB.Where("telegram_id IS NOT NULL")
	switch targetType {
	case "registered":
		query = query.Where("is_registered = ?", true)
	case "all":
		// все с telegram_id
	default:
		return nil, errors.New("Некорректный тип аудитории")
	}
	if err := query.Order("first_name").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func UpdateBroadcastStatus(broadcastID uint, status string) error {
	return db.DB.Model(&db.BroadcastMessage{}).Where("id = ?", broadcastID).Update("status", status).Error
}

func ParseTargetUserIDs(jsonStr string) ([]uint, error) {
	var ids []uint
	if jsonStr == "" {
		return ids, nil
	}
	if err := json.Unmarshal([]byte(jsonStr), &ids); err != nil {
		return nil, err
	}
	return ids, nil
} 