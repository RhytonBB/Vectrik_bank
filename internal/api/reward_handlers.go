package api

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
)

var rewardTemplate *template.Template

func init() {
	var err error
	rewardTemplate, err = template.ParseFiles("web/templates/admin_reward.html")
	if err != nil {
		log.Fatalf("❌ Failed to parse admin_reward template: %v", err)
	}
}

func AdminRewardHandler(w http.ResponseWriter, r *http.Request) {
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

	var adminID uint
	switch v := claims["user_id"].(type) {
	case float64:
		adminID = uint(v)
	case int:
		adminID = uint(v)
	case uint:
		adminID = v
	default:
		http.Error(w, "Неверный ID администратора", http.StatusInternalServerError)
		return
	}


	switch r.Method {
	case http.MethodGet:
		handleRewardGet(w, r)
	case http.MethodPost:
		handleRewardPost(w, r, adminID)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func handleRewardGet(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	users, err := services.RewardSearchUsers(query)
	if err != nil {
		log.Printf("❌ Ошибка выборки пользователей: %v", err)
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	data := struct {
		Users   []db.User
		Query   string
		Message string
		Error   string
	}{
		Users: users,
		Query: query,
	}
	if err := rewardTemplate.Execute(w, data); err != nil {
		log.Printf("❌ Ошибка рендера шаблона: %v", err)
	}
}

func handleRewardPost(w http.ResponseWriter, r *http.Request, adminID uint) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Неверные данные формы", http.StatusBadRequest)
		return
	}
	userIDsStr := r.Form["user_ids"]
	if len(userIDsStr) == 0 {
		renderRewardWithError(w, "Выберите хотя бы одного пользователя")
		return
	}
	var userIDs []uint
	for _, s := range userIDsStr {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			renderRewardWithError(w, "Некорректный ID пользователя")
			return
		}
		userIDs = append(userIDs, uint(id))
	}
	amountStr := r.FormValue("amount")
	reason := r.FormValue("reason")
	context := r.FormValue("context")
	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		renderRewardWithError(w, "Некорректная сумма")
		return
	}
	err = services.RewardUsers(userIDs, amount, reason, context, adminID)
	if err != nil {
		renderRewardWithError(w, err.Error())
		return
	}
	renderRewardWithMessage(w, "Вектрики успешно начислены")
}

func renderRewardWithError(w http.ResponseWriter, errMsg string) {
	data := struct {
		Error string
	}{
		Error: errMsg,
	}
	rewardTemplate.Execute(w, data)
}

func renderRewardWithMessage(w http.ResponseWriter, msg string) {
	data := struct {
		Message string
	}{
		Message: msg,
	}
	rewardTemplate.Execute(w, data)
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
	query := strings.TrimSpace(r.URL.Query().Get("q"))
<<<<<<< HEAD
	users, err := services.RewardSearchUsers(query)
	if err != nil {
=======
	var users []db.User
	tx := db.DB.Limit(20).Order("first_name")
	if query != "" {
		likeQuery := "%" + query + "%"
		tx = tx.Where(
			"first_name ILIKE ? OR last_name ILIKE ? OR telegram_username ILIKE ?",
			likeQuery, likeQuery, likeQuery,
		)
	}
	if err := tx.Find(&users).Error; err != nil {
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// API: Начисление вектриков через JSON
func AdminRewardAPIHandler(w http.ResponseWriter, r *http.Request) {
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
<<<<<<< HEAD
	err = services.RewardUsers(req.UserIDs, req.Amount, req.Reason, req.Context, adminID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
=======
	if len(req.UserIDs) == 0 || req.Amount <= 0 || req.Reason == "" {
		http.Error(w, "Все поля обязательны", http.StatusBadRequest)
		return
	}
	validReasons := map[string]bool{
		"drop":       true,
		"активность": true,
		"награда":    true,
	}
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
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}
