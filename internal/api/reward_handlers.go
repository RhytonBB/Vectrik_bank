package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
	"gorm.io/gorm"
)

// Удалить переменную rewardTemplate, все вызовы template.ParseFiles, template.Execute и т.д.
// Все функции handleRewardGet, handleRewardPost, renderRewardWithError, renderRewardWithMessage должны возвращать только JSON-ответы или http.Error.

func AdminRewardHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Метод не поддерживается. Используйте /api/admin/reward", http.StatusMethodNotAllowed)
}

func handleRewardGet(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	users, err := services.RewardSearchUsers(query)
	if err != nil {
		log.Printf("❌ Ошибка выборки пользователей: %v", err)
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handleRewardPost(w http.ResponseWriter, r *http.Request, adminID uint) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Неверные данные формы", http.StatusBadRequest)
		return
	}
	userIDsStr := r.Form["user_ids"]
	if len(userIDsStr) == 0 {
		http.Error(w, "Выберите хотя бы одного пользователя", http.StatusBadRequest)
		return
	}
	var userIDs []uint
	for _, s := range userIDsStr {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
			return
		}
		userIDs = append(userIDs, uint(id))
	}
	amountStr := r.FormValue("amount")
	reason := r.FormValue("reason")
	context := r.FormValue("context")
	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		http.Error(w, "Некорректная сумма", http.StatusBadRequest)
		return
	}
	err = services.RewardUsers(userIDs, amount, reason, context, adminID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Вектрики успешно начислены"}`))
}

// API: Поиск пользователей для reward в JSON
func AdminRewardUsersAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	scope := r.URL.Query().Get("scope")
	var users []db.User
	if scope == "group" {
		var adminID uint
		switch v := claims["user_id"].(type) {
		case float64:
			adminID = uint(v)
		case int:
			adminID = uint(v)
		case uint:
			adminID = v
		}
		users, _ = services.GetGroupUsersByAdmin(adminID)
	} else {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		users, _ = services.RewardSearchUsers(query)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// API: Начисление вектриков через JSON
func AdminRewardAPIHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("REWARD API: received request")
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
	var req struct {
		UserIDs []uint `json:"user_ids"`
		Amount  int    `json:"amount"`
		Reason  string `json:"reason"`
		Context string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Printf("REWARD API: payload: %+v\n", req)
	if len(req.UserIDs) == 0 || req.Amount <= 0 || req.Reason == "" {
		http.Error(w, "Все поля обязательны", http.StatusBadRequest)
		return
	}
	validReasons := services.GetValidRewardReasons()
	if !validReasons[req.Reason] {
		http.Error(w, "Некорректная причина", http.StatusBadRequest)
		return
	}
	tx := db.DB.Begin()
	for _, userID := range req.UserIDs {
		transaction := db.Transaction{
			UserID:    userID,
			Amount:    req.Amount,
			Reason:    req.Reason,
			Source:    "reward",
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&transaction).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Ошибка при создании транзакции", http.StatusInternalServerError)
			return
		}
		if err := tx.Model(&db.User{}).
			Where("id = ?", userID).
			UpdateColumn("balance", gorm.Expr("balance + ?", req.Amount)).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Ошибка при обновлении баланса", http.StatusInternalServerError)
			return
		}
		adminLog := db.AdminLog{
			AdminID:      adminID,
			TargetUserID: userID,
			Action:       "reward",
			Amount:       &req.Amount,
			CreatedAt:    time.Now(),
		}
		if req.Context != "" {
			adminLog.Context = &req.Context
		}
		if err := tx.Create(&adminLog).Error; err != nil {
			tx.Rollback()
			http.Error(w, "Ошибка при записи лога", http.StatusInternalServerError)
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Ошибка при сохранении данных", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}
