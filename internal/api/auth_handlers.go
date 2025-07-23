package api

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/pkg/hash"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
)

// Удалить переменные templates, функции renderLogin, renderDashboard, init, и все связанные с ними вызовы.
// Везде, где был редирект на /login или /dashboard, возвращать ошибку авторизации или успешный JSON-ответ.

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Пожалуйста, авторизуйтесь."}`))
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
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Ошибка обработки формы."}`))
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
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Заполните все поля."}`))
			return
		}
	}

	user, err := services.GetUserByLogin(login)
	if err != nil {
		log.Printf("❌ Пользователь не найден: %v", err)
		if isJSON {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Неверный логин или пароль."}`))
			return
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Неверный логин или пароль."}`))
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
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Неверный логин или пароль."}`))
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
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Ошибка авторизации"}`))
			return
		}
	}

	// ВСЕГДА устанавливай cookie!
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		Domain:   "vectrik-bank.onrender.com", // добавлено для кросс-доменных cookie
		HttpOnly: true,
		Secure:   true, // обязательно для https
		SameSite: http.SameSiteNoneMode, // обязательно для кросс-доменных cookie
	})

	if isJSON {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"token": "` + token + `"}`))
		return
	}

	log.Printf("✅ Успешный вход: %s", login)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Успешный вход"}`))
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Пожалуйста, авторизуйтесь."}`))
		return
	}

	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		log.Printf("❌ Ошибка разбора токена: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Пожалуйста, авторизуйтесь."}`))
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
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Пожалуйста, авторизуйтесь."}`))
		return
	}

	balance, err := services.GetUserBalance(userID)
	if err != nil {
		log.Printf("❌ Ошибка получения пользователя: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Ошибка базы данных"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"balance": balance})
}


func adminBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"error": "Рассылка — в разработке."}`))
}


func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Успешный выход"}`))
}

// API: Получить баланс текущего пользователя/админа в JSON
func AdminBalanceAPIHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}
	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Unauthorized"}`))
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
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Invalid user_id"}`))
		return
	}
	balance, err := services.GetUserBalance(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "DB error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"balance": balance})
}

// API: Получить информацию о текущем пользователе
func MeAPIHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}
	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}
	userID, ok := claims["user_id"].(float64)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "Invalid user_id"}`))
		return
	}
	user, err := services.GetUserShortInfo(uint(userID))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "User not found"}`))
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
