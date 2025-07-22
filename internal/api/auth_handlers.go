package api

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"gorm.io/gorm/clause"
	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/RhytonBB/vectrik-webbank/pkg/hash"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
)

var templates *template.Template

func init() {
	files, err := filepath.Glob("web/templates/*.html")
	if err != nil {
		log.Fatalf("❌ Failed to find templates: %v", err)
	}
	if len(files) == 0 {
		log.Fatal("❌ No templates found in web/templates/")
	}
	log.Printf("📄 Templates loaded: %v\n", files)

	templates = template.Must(template.ParseGlob("web/templates/*.html"))
}

func renderLogin(w http.ResponseWriter, errorMsg string) {
	log.Println("🔃 Rendering login page...")
	data := struct {
		Error string
	}{
		Error: errorMsg,
	}

	err := templates.ExecuteTemplate(w, "login.html", data)
	if err != nil {
		log.Printf("❌ Template execution error: %v", err)
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
	}
}

func renderDashboard(w http.ResponseWriter, isAdmin bool, balance int) {
	if isAdmin {
		data := struct {
			Balance int
		}{
			Balance: balance,
		}
		err := templates.ExecuteTemplate(w, "admin_dashboard.html", data)
		if err != nil {
			log.Printf("❌ Template execution error (admin_dashboard): %v", err)
			http.Error(w, "Ошибка загрузки панели администратора", http.StatusInternalServerError)
		}
	} else {
		data := struct {
			IsAdmin bool
			Balance int
		}{
			IsAdmin: false,
			Balance: balance,
		}
		err := templates.ExecuteTemplate(w, "dashboard.html", data)
		if err != nil {
			log.Printf("❌ Template execution error (dashboard): %v", err)
			http.Error(w, "Ошибка загрузки панели пользователя", http.StatusInternalServerError)
		}
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderLogin(w, "")
		return
	}

	isJSON := r.Header.Get("Content-Type") == "application/json" || r.Header.Get("Accept") == "application/json"

	var login, password string
	if isJSON {
		type loginRequest struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		var req loginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Некорректный JSON"}`))
			return
		}
		login = req.Login
		password = req.Password
	} else {
		if err := r.ParseForm(); err != nil {
			renderLogin(w, "Ошибка обработки формы.")
			return
		}
		login = r.FormValue("login")
		password = r.FormValue("password")
	}

	if login == "" || password == "" {
		if isJSON {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Заполните все поля."}`))
			return
		} else {
			renderLogin(w, "Заполните все поля.")
			return
		}
	}

	var user db.User
	result := db.DB.Where("login = ?", login).First(&user)
	if result.Error != nil {
		log.Printf("❌ Пользователь не найден: %v", result.Error)
		if isJSON {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Неверный логин или пароль."}`))
			return
		} else {
			renderLogin(w, "Неверный логин или пароль.")
			return
		}
	}

	if !hash.CheckPasswordHash(password, *user.PasswordHash) {
		log.Println("❌ Неверный пароль")
		if isJSON {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Неверный логин или пароль."}`))
			return
		} else {
			renderLogin(w, "Неверный логин или пароль.")
			return
		}
	}

	token, err := auth.GenerateJWT(user.ID, user.IsAdmin)
	if err != nil {
		log.Printf("❌ Ошибка генерации JWT: %v", err)
		if isJSON {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Ошибка авторизации"}`))
			return
		} else {
			http.Error(w, "Ошибка авторизации", http.StatusInternalServerError)
			return
		}
	}

	// ВСЕГДА устанавливай cookie!
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

	if isJSON {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"token": "` + token + `"}`))
		return
	}

	log.Printf("✅ Успешный вход: %s", login)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		log.Printf("❌ Ошибка разбора токена: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	isAdmin, ok := claims["is_admin"].(bool)
	if !ok {
		log.Println("❌ Не удалось определить роль пользователя")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
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
		log.Println("❌ Не удалось получить user_id из токена: неизвестный тип")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user db.User
	if err := db.DB.Select("balance").Where("id = ?", userID).First(&user).Error; err != nil {
		log.Printf("❌ Ошибка получения пользователя: %v", err)
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	renderDashboard(w, isAdmin, user.Balance)
}


func adminBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Рассылка — в разработке"))
}


func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// API: Получить баланс текущего пользователя/админа в JSON
func AdminBalanceAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	var userID uint
	switch v := claims["user_id"].(type) {
	case float64:
		userID = uint(v)
	case int:
		userID = uint(v)
	case uint:
		userID = v
	default:
		http.Error(w, "Invalid user_id", http.StatusInternalServerError)
		return
	}
	var user db.User
	if err := db.DB.Select("balance").Where("id = ?", userID).First(&user).Error; err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"balance": user.Balance})
}

// API: Получить информацию о текущем пользователе
func MeAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	userID, ok := claims["user_id"].(float64)
	if !ok {
		http.Error(w, "Invalid user_id", http.StatusInternalServerError)
		return
	}
	var user db.User
	if err := db.DB.Select("id, first_name, last_name, is_admin, balance").Where("id = ?", uint(userID)).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": user.ID,
		"first_name": user.FirstName,
		"last_name": user.LastName,
		"is_admin": user.IsAdmin,
		"balance": user.Balance,
	})
}

// API: Получить текущую тему пользователей
func AdminThemeGetHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := auth.ParseJWT(cookie.Value); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	theme := services.GetTheme()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"theme": "` + theme + `"}`))
}

// API: Установить тему пользователей
func AdminThemeSetHandler(w http.ResponseWriter, r *http.Request) {
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
		Theme string `json:"theme"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Theme != "green" && req.Theme != "pink" {
		http.Error(w, "Invalid theme", http.StatusBadRequest)
		return
	}
	setting := db.Setting{
		Key:   "user_theme",
		Value: req.Theme,
	}
	db.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&setting)
	services.SetTheme(req.Theme)
	w.WriteHeader(http.StatusOK)
}
