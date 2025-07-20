package api

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"gorm.io/gorm"
	"github.com/RhytonBB/vectrik-webbank/pkg/hash"
)

var (
	userListTemplate     = template.Must(template.ParseFiles("web/templates/admin_users.html"))
	userDetailTemplate   = template.Must(template.ParseFiles("web/templates/admin_user_detail.html"))
	userRegisterTemplate = template.Must(template.ParseFiles("web/templates/admin_user_register.html"))
)

func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[AdminUsersHandler] called, method:", r.Method)
	cookie, err := r.Cookie("token")
	if err != nil {
		log.Println("[AdminUsersHandler] cookie error:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		log.Println("[AdminUsersHandler] ParseJWT error:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	isAdmin, ok := claims["is_admin"].(bool)
	if !ok || !isAdmin {
		log.Println("[AdminUsersHandler] not admin or no is_admin claim")
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var users []db.User
		if err := db.DB.Order("first_name").Find(&users).Error; err != nil {
			http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
			return
		}
		userListTemplate.Execute(w, struct{ Users []db.User }{users})
	case http.MethodPost:
		handleAdjustment(w, r, claims)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func AdminUserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		userRegisterTemplate.Execute(w, nil)
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			userRegisterTemplate.Execute(w, struct{ Error string }{"Ошибка обработки формы"})
			return
		}

		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		tgUsername := strings.TrimSpace(r.FormValue("tg_username"))

		if firstName == "" || lastName == "" || tgUsername == "" {
			userRegisterTemplate.Execute(w, struct{ Error string }{"Все поля обязательны"})
			return
		}

		user := db.User{
			FirstName:        firstName,
			LastName:         lastName,
			TelegramUsername: tgUsername,
		}
		if err := db.DB.Create(&user).Error; err != nil {
			userRegisterTemplate.Execute(w, struct{ Error string }{"Ошибка: " + err.Error()})
			return
		}

		userRegisterTemplate.Execute(w, struct{ Message string }{"Пользователь успешно создан"})
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func AdminUserDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}
	userID := uint(userID64)

	var user db.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	var purchases []db.Purchase
	var transactions []db.Transaction

	if err := db.DB.Preload("Product").Where("user_id = ?", userID).Find(&purchases).Error; err != nil {
		log.Printf("Ошибка получения покупок: %v", err)
	}
	if err := db.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&transactions).Error; err != nil {
		log.Printf("Ошибка получения транзакций: %v", err)
	}

	userDetailTemplate.Execute(w, struct {
		User         db.User
		Purchases    []db.Purchase
		Transactions []db.Transaction
	}{
		User:         user,
		Purchases:    purchases,
		Transactions: transactions,
	})
}

func handleAdjustment(w http.ResponseWriter, r *http.Request, claims map[string]interface{}) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	userIDStr := r.FormValue("user_id")
	amountStr := r.FormValue("amount")
	context := r.FormValue("context")

	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount == 0 {
		http.Error(w, "Неверная сумма", http.StatusBadRequest)
		return
	}

	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный ID пользователя", http.StatusBadRequest)
		return
	}
	userID := uint(userID64)

	var adminID uint
	switch v := claims["user_id"].(type) {
	case float64:
		adminID = uint(v)
	case int:
		adminID = uint(v)
	case uint:
		adminID = v
	default:
		http.Error(w, "Некорректный ID администратора", http.StatusInternalServerError)
		return
	}


	tx := db.DB.Begin()

	transaction := db.Transaction{
		UserID:    userID,
		Amount:    amount,
		Reason:    "adjustment",
		Source:    "adjustment",
			CreatedAt: time.Now(),
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Ошибка транзакции", http.StatusInternalServerError)
		return
	}

	if err := tx.Model(&db.User{}).
		Where("id = ?", userID).
		UpdateColumn("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Ошибка обновления баланса", http.StatusInternalServerError)
		return
	}

	adminLog := db.AdminLog{
		AdminID:      adminID,
		TargetUserID: userID,
		Action:       "adjustment",
		Amount:       &amount,
		Context:      &context,
		CreatedAt:    time.Now(),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Ошибка записи лога", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Ошибка при сохранении", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// API: Получить список пользователей в JSON
func AdminUsersAPIHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[AdminUsersAPIHandler] called, method:", r.Method)
	if r.Method == http.MethodPost {
		cookie, err := r.Cookie("token")
		if err != nil {
			log.Println("[AdminUsersAPIHandler] cookie error:", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		claims, err := auth.ParseJWT(cookie.Value)
		if err != nil {
			log.Println("[AdminUsersAPIHandler] ParseJWT error:", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		isAdmin, ok := claims["is_admin"].(bool)
		if !ok || !isAdmin {
			log.Println("[AdminUsersAPIHandler] not admin or no is_admin claim")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		var req struct {
			FirstName string `json:"first_name"`
			LastName string `json:"last_name"`
			TelegramUsername string `json:"tg_username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println("[AdminUsersAPIHandler] decode error:", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if req.FirstName == "" || req.LastName == "" || req.TelegramUsername == "" {
			log.Println("[AdminUsersAPIHandler] missing fields")
			http.Error(w, "Все поля обязательны", http.StatusBadRequest)
			return
		}
		user := db.User{
			FirstName: req.FirstName,
			LastName: req.LastName,
			TelegramUsername: req.TelegramUsername,
		}
		if err := db.DB.Create(&user).Error; err != nil {
			log.Println("[AdminUsersAPIHandler] DB create error:", err)
			http.Error(w, "Ошибка: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}
	cookie, err := r.Cookie("token")
	if err != nil {
		log.Println("[AdminUsersAPIHandler] cookie error:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		log.Println("[AdminUsersAPIHandler] ParseJWT error:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	isAdmin, ok := claims["is_admin"].(bool)
	if !ok || !isAdmin {
		log.Println("[AdminUsersAPIHandler] not admin or no is_admin claim")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var users []db.User
	q := r.URL.Query().Get("q")
    minBalanceStr := r.URL.Query().Get("min_balance")
    maxBalanceStr := r.URL.Query().Get("max_balance")
	tx := db.DB.Order("first_name")
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("first_name ILIKE ? OR last_name ILIKE ? OR telegram_username ILIKE ?", like, like, like)
	}
    if minBalanceStr != "" {
        if min, err := strconv.Atoi(minBalanceStr); err == nil {
            tx = tx.Where("balance >= ?", min)
        }
    }
    if maxBalanceStr != "" {
        if max, err := strconv.Atoi(maxBalanceStr); err == nil {
            tx = tx.Where("balance <= ?", max)
        }
    }
	if err := tx.Find(&users).Error; err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// API: Детали пользователя
func AdminUserDetailAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	userIDStr := vars["id"]
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}
	userID := uint(userID64)
	var user db.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// API: Активация/деактивация пользователя
func AdminUserActivateAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	userIDStr := vars["id"]
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}
	userID := uint(userID64)
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := db.DB.Model(&db.User{}).Where("id = ?", userID).Update("is_active", req.IsActive).Error; err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// API: Назначение/снятие админа
func AdminUserAdminAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	userIDStr := vars["id"]
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}
	userID := uint(userID64)
	var req struct {
		IsAdmin bool `json:"is_admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := db.DB.Model(&db.User{}).Where("id = ?", userID).Update("is_admin", req.IsAdmin).Error; err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// API: Бан пользователя
func AdminUserBanAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	userIDStr := vars["id"]
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}
	userID := uint(userID64)
	if err := db.DB.Model(&db.User{}).Where("id = ?", userID).Update("is_banned", true).Error; err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// API: Разбан пользователя
func AdminUserUnbanAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	userIDStr := vars["id"]
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}
	userID := uint(userID64)
	if err := db.DB.Model(&db.User{}).Where("id = ?", userID).Update("is_banned", false).Error; err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// API: Редактирование и удаление пользователя
func AdminUserEditDeleteAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	userIDStr := vars["id"]
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}
	userID := uint(userID64)

	switch r.Method {
	case http.MethodPut:
		var req struct {
			FirstName string  `json:"first_name"`
			LastName  string  `json:"last_name"`
			TelegramUsername string `json:"telegram_username"`
			Login     *string `json:"login"`
			Password  *string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		updates := map[string]interface{}{
			"first_name": req.FirstName,
			"last_name": req.LastName,
			"telegram_username": req.TelegramUsername,
		}
		if req.Login != nil {
			updates["login"] = req.Login
		}
		if req.Password != nil && *req.Password != "" {
			hashStr, err := hash.HashPassword(*req.Password)
			if err != nil {
				http.Error(w, "Ошибка хеширования пароля", http.StatusInternalServerError)
				return
			}
			updates["password_hash"] = hashStr
		}
		if err := db.DB.Model(&db.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			http.Error(w, "Ошибка обновления пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	case http.MethodDelete:
		if err := db.DB.Delete(&db.User{}, userID).Error; err != nil {
			http.Error(w, "Ошибка удаления пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}
