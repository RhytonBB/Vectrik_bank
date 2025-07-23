package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/RhytonBB/vectrik-webbank/pkg/hash"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
)

// Удалить переменные userListTemplate, userDetailTemplate, userRegisterTemplate и все вызовы template.ParseFiles, template.Execute и т.д. Оставить только API-методы и бизнес-логику через сервисы.

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
		users, err := services.GetAllUsers()
		if err != nil {
			http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	case http.MethodPost:
		handleAdjustment(w, r, claims)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func AdminUserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// userRegisterTemplate.Execute(w, nil) // Removed
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			// userRegisterTemplate.Execute(w, struct{ Error string }{"Ошибка обработки формы"}) // Removed
			http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
			return
		}
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		tgUsername := strings.TrimSpace(r.FormValue("tg_username"))
		user := db.User{
			FirstName:        firstName,
			LastName:         lastName,
			TelegramUsername: tgUsername,
		}
		err := services.CreateUser(&user)
		if err != nil {
			// userRegisterTemplate.Execute(w, struct{ Error string }{"Ошибка: " + err.Error()}) // Removed
			http.Error(w, "Ошибка: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// userRegisterTemplate.Execute(w, struct{ Message string }{"Пользователь успешно создан"}) // Removed
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
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
	user, err := services.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}
	purchases, _ := services.GetUserPurchases(userID)
	transactions, _ := services.GetUserTransactions(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		User         db.User
		Purchases    []db.Purchase
		Transactions []db.Transaction
	}{
		User:         *user,
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
	err = services.AdjustUserBalance(userID, amount, context, adminID)
	if err != nil {
		http.Error(w, "Ошибка корректировки баланса: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
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
		user := db.User{
			FirstName: req.FirstName,
			LastName: req.LastName,
			TelegramUsername: req.TelegramUsername,
		}
		if err := services.CreateUser(&user); err != nil {
			log.Println("[AdminUsersAPIHandler] CreateUser error:", err)
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
	users, err := services.GetAllUsers()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// API: Получить список всех групп
func AdminGroupsAPIHandler(w http.ResponseWriter, r *http.Request) {
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
    if r.Method == http.MethodPost {
        var req struct {
            Name     string `json:"name"`
            AdminIDs []uint `json:"admin_ids"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }
        group, err := services.CreateGroupWithAdmins(req.Name, req.AdminIDs)
        if err != nil {
            http.Error(w, "Ошибка создания группы: "+err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(group)
        return
    }
    var groups []db.Group
    if err := db.DB.Order("name").Find(&groups).Error; err != nil {
        http.Error(w, "DB error", http.StatusInternalServerError)
        return
    }
    // Для каждого group добавить список админов
    type GroupWithAdmins struct {
        ID    uint        `json:"id"`
        Name  string      `json:"name"`
        Admins []db.User  `json:"admins"`
    }
    var result []GroupWithAdmins
    for _, g := range groups {
        var groupAdmins []db.GroupAdmin
        db.DB.Where("group_id = ?", g.ID).Find(&groupAdmins)
        var admins []db.User
        for _, ga := range groupAdmins {
            var admin db.User
            if err := db.DB.First(&admin, ga.AdminID).Error; err == nil {
                admins = append(admins, admin)
            }
        }
        result = append(result, GroupWithAdmins{
            ID: g.ID,
            Name: g.Name,
            Admins: admins,
        })
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
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
	user, err := services.GetUserByID(userID)
	if err != nil {
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
	var errAct error
	if req.IsActive {
		errAct = services.ActivateUser(userID)
	} else {
		errAct = services.DeactivateUser(userID)
	}
	if errAct != nil {
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
	if err := services.SetUserAdmin(userID, req.IsAdmin); err != nil {
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
	if err := services.BanUser(userID); err != nil {
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
	if err := services.UnbanUser(userID); err != nil {
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
			GroupID   *uint   `json:"group_id"`
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
		if req.GroupID != nil {
			updates["group_id"] = req.GroupID
		}
		if err := services.UpdateUser(userID, updates); err != nil {
			http.Error(w, "Ошибка обновления пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	case http.MethodDelete:
		if err := services.DeleteUser(userID); err != nil {
			http.Error(w, "Ошибка удаления пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func AdminUserAdjustBalanceAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		UserID  uint   `json:"user_id"`
		Amount  int    `json:"amount"`
		Context string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Amount == 0 {
		http.Error(w, "Amount must not be zero", http.StatusBadRequest)
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
	err = services.AdjustUserBalance(req.UserID, req.Amount, req.Context, adminID)
	if err != nil {
		http.Error(w, "Ошибка корректировки баланса: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func AdminThemeGetHandler(w http.ResponseWriter, r *http.Request) {
    theme := services.GetTheme()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"theme": theme})
}

func AdminThemeSetHandler(w http.ResponseWriter, r *http.Request) {
    var req struct{ Theme string `json:"theme"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    services.SetTheme(req.Theme)
    w.WriteHeader(http.StatusOK)
}
