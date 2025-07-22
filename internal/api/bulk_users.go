package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
)

type BulkUser struct {
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	TelegramUsername string `json:"telegram_username"`
}

type BulkPreviewResponse struct {
	Users []BulkUser `json:"users"`
}

type BulkRegisterRequest struct {
	Users []BulkUser `json:"users"`
}

type BulkRegisterResponse struct {
	Created int      `json:"created"`
	Errors  []string `json:"errors,omitempty"`
}

// API: Предварительный просмотр файла с пользователями
func AdminBulkUsersPreviewAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
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
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	var users []services.BulkUser
	var parseErr error
	switch ext {
	case ".csv":
		users, parseErr = services.ParseCSVFile(file)
	case ".xlsx", ".xls":
		users, parseErr = services.ParseExcelFile(file)
	default:
		http.Error(w, "Unsupported file format. Use CSV or Excel files", http.StatusBadRequest)
		return
	}
	if parseErr != nil {
		http.Error(w, "Error parsing file: "+parseErr.Error(), http.StatusBadRequest)
		return
	}
	response := services.BulkPreviewUsers(users)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func AdminBulkUsersRegisterAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
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
	var req services.BulkRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if len(req.Users) == 0 {
		http.Error(w, "No users to register", http.StatusBadRequest)
		return
	}
	if len(req.Users) > 1000 {
		http.Error(w, "Too many users. Maximum 1000 users per request", http.StatusBadRequest)
		return
	}
	response := services.BulkRegisterUsers(req.Users)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
} 