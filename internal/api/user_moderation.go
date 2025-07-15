package api

import (
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
)

var (
	userListTemplate     = template.Must(template.ParseFiles("web/templates/admin_users.html"))
	userDetailTemplate   = template.Must(template.ParseFiles("web/templates/admin_user_detail.html"))
	userRegisterTemplate = template.Must(template.ParseFiles("web/templates/admin_user_register.html"))
)

func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	isAdmin, ok := claims["is_admin"].(bool)
	if !ok || !isAdmin {
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
