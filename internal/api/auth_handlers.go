package api

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/RhytonBB/vectrik-webbank/pkg/hash"
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

	if err := r.ParseForm(); err != nil {
		renderLogin(w, "Ошибка обработки формы.")
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	if login == "" || password == "" {
		renderLogin(w, "Заполните все поля.")
		return
	}

	var user db.User
	result := db.DB.Where("login = ?", login).First(&user)
	if result.Error != nil {
		log.Printf("❌ Пользователь не найден: %v", result.Error)
		renderLogin(w, "Неверный логин или пароль.")
		return
	}

	if !hash.CheckPasswordHash(password, *user.PasswordHash) {
		log.Println("❌ Неверный пароль")
		renderLogin(w, "Неверный логин или пароль.")
		return
	}

	token, err := auth.GenerateJWT(user.ID, user.IsAdmin)
	if err != nil {
		log.Printf("❌ Ошибка генерации JWT: %v", err)
		http.Error(w, "Ошибка авторизации", http.StatusInternalServerError)
		return
	}


	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

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
