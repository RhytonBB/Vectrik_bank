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

// Глобальная переменная шаблонов
var templates *template.Template

func init() {
	// Проверяем наличие шаблонов
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

func renderDashboard(w http.ResponseWriter, isAdmin bool) {
	log.Println("🔃 Rendering dashboard page...")
	data := struct {
		IsAdmin bool
	}{
		IsAdmin: isAdmin,
	}

	err := templates.ExecuteTemplate(w, "dashboard.html", data)
	if err != nil {
		log.Printf("❌ Template execution error: %v", err)
		http.Error(w, "Ошибка загрузки панели", http.StatusInternalServerError)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderLogin(w, "")
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

	token, err := auth.GenerateJWT(user.ID.String(), user.IsAdmin)
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

	renderDashboard(w, isAdmin)
}
