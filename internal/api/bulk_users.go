package api

import (
<<<<<<< HEAD
	"encoding/json"
=======
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
	"net/http"
	"path/filepath"
	"strings"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
<<<<<<< HEAD
	"github.com/RhytonBB/vectrik-webbank/internal/services"
=======
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/xuri/excelize/v2"
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
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
<<<<<<< HEAD
=======

	// Проверка авторизации
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
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
<<<<<<< HEAD
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}
=======

	// Парсинг multipart формы
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

>>>>>>> 7259f56d86e51521de5028adf64898f696632916
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()
<<<<<<< HEAD
	ext := strings.ToLower(filepath.Ext(header.Filename))
	var users []services.BulkUser
	var parseErr error
	switch ext {
	case ".csv":
		users, parseErr = services.ParseCSVFile(file)
	case ".xlsx", ".xls":
		users, parseErr = services.ParseExcelFile(file)
=======

	// Определение типа файла
	ext := strings.ToLower(filepath.Ext(header.Filename))
	var users []BulkUser

	switch ext {
	case ".csv":
		users, err = parseCSVFile(file)
	case ".xlsx", ".xls":
		users, err = parseExcelFile(file)
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
	default:
		http.Error(w, "Unsupported file format. Use CSV or Excel files", http.StatusBadRequest)
		return
	}
<<<<<<< HEAD
	if parseErr != nil {
		http.Error(w, "Error parsing file: "+parseErr.Error(), http.StatusBadRequest)
		return
	}
	response := services.BulkPreviewUsers(users)
=======

	if err != nil {
		http.Error(w, "Error parsing file: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация данных
	var validUsers []BulkUser
	for i, user := range users {
		if user.FirstName == "" || user.LastName == "" || user.TelegramUsername == "" {
			log.Printf("Skipping invalid user at row %d: missing required fields", i+1)
			continue
		}
		validUsers = append(validUsers, user)
	}

	response := BulkPreviewResponse{
		Users: validUsers,
	}

>>>>>>> 7259f56d86e51521de5028adf64898f696632916
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

<<<<<<< HEAD
=======
// API: Массовая регистрация пользователей
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
func AdminBulkUsersRegisterAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
<<<<<<< HEAD
=======

	// Проверка авторизации
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
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
<<<<<<< HEAD
	var req services.BulkRegisterRequest
=======

	var req BulkRegisterRequest
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
<<<<<<< HEAD
=======

>>>>>>> 7259f56d86e51521de5028adf64898f696632916
	if len(req.Users) == 0 {
		http.Error(w, "No users to register", http.StatusBadRequest)
		return
	}
<<<<<<< HEAD
=======

>>>>>>> 7259f56d86e51521de5028adf64898f696632916
	if len(req.Users) > 1000 {
		http.Error(w, "Too many users. Maximum 1000 users per request", http.StatusBadRequest)
		return
	}
<<<<<<< HEAD
	response := services.BulkRegisterUsers(req.Users)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
=======

	// Массовая регистрация
	created := 0
	var errors []string

	for i, user := range req.Users {
		// Валидация
		if user.FirstName == "" || user.LastName == "" || user.TelegramUsername == "" {
			errors = append(errors, fmt.Sprintf("Row %d: missing required fields", i+1))
			continue
		}

		// Проверка уникальности Telegram username
		var existingUser db.User
		if err := db.DB.Where("telegram_username = ?", user.TelegramUsername).First(&existingUser).Error; err == nil {
			errors = append(errors, fmt.Sprintf("Row %d: telegram username '%s' already exists", i+1, user.TelegramUsername))
			continue
		}

		// Создание пользователя
		newUser := db.User{
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			TelegramUsername: user.TelegramUsername,
			IsActive:         true,
		}

		if err := db.DB.Create(&newUser).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: %s", i+1, err.Error()))
			continue
		}

		created++
	}

	response := BulkRegisterResponse{
		Created: created,
		Errors:  errors,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Парсинг CSV файла
func parseCSVFile(file io.Reader) ([]BulkUser, error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Разрешаем разное количество полей

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("file must contain at least header and one data row")
	}

	// Определяем индексы колонок по заголовкам
	headers := records[0]
	firstNameIdx := -1
	lastNameIdx := -1
	telegramIdx := -1

	for i, header := range headers {
		header = strings.ToLower(strings.TrimSpace(header))
		switch header {
		case "имя", "first name", "firstname", "name":
			firstNameIdx = i
		case "фамилия", "last name", "lastname", "surname":
			lastNameIdx = i
		case "telegram", "telegram username", "username", "tg", "telegram_username":
			telegramIdx = i
		}
	}

	if firstNameIdx == -1 || lastNameIdx == -1 || telegramIdx == -1 {
		return nil, fmt.Errorf("required columns not found. Expected: Имя/FirstName, Фамилия/LastName, Telegram/Username")
	}

	var users []BulkUser
	for i, record := range records[1:] {
		if len(record) <= firstNameIdx || len(record) <= lastNameIdx || len(record) <= telegramIdx {
			log.Printf("Skipping row %d: insufficient columns", i+2)
			continue
		}

		user := BulkUser{
			FirstName:        strings.TrimSpace(record[firstNameIdx]),
			LastName:         strings.TrimSpace(record[lastNameIdx]),
			TelegramUsername: strings.TrimSpace(record[telegramIdx]),
		}

		// Убираем @ из начала telegram username если есть
		if strings.HasPrefix(user.TelegramUsername, "@") {
			user.TelegramUsername = user.TelegramUsername[1:]
		}

		users = append(users, user)
	}

	return users, nil
}

// Парсинг Excel файла
func parseExcelFile(file io.Reader) ([]BulkUser, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Получаем первый лист
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("file must contain at least header and one data row")
	}

	// Определяем индексы колонок по заголовкам
	headers := rows[0]
	firstNameIdx := -1
	lastNameIdx := -1
	telegramIdx := -1

	for i, header := range headers {
		header = strings.ToLower(strings.TrimSpace(header))
		switch header {
		case "имя", "first name", "firstname", "name":
			firstNameIdx = i
		case "фамилия", "last name", "lastname", "surname":
			lastNameIdx = i
		case "telegram", "telegram username", "username", "tg", "telegram_username":
			telegramIdx = i
		}
	}

	if firstNameIdx == -1 || lastNameIdx == -1 || telegramIdx == -1 {
		return nil, fmt.Errorf("required columns not found. Expected: Имя/FirstName, Фамилия/LastName, Telegram/Username")
	}

	var users []BulkUser
	for i, row := range rows[1:] {
		if len(row) <= firstNameIdx || len(row) <= lastNameIdx || len(row) <= telegramIdx {
			log.Printf("Skipping row %d: insufficient columns", i+2)
			continue
		}

		user := BulkUser{
			FirstName:        strings.TrimSpace(row[firstNameIdx]),
			LastName:         strings.TrimSpace(row[lastNameIdx]),
			TelegramUsername: strings.TrimSpace(row[telegramIdx]),
		}

		// Убираем @ из начала telegram username если есть
		if strings.HasPrefix(user.TelegramUsername, "@") {
			user.TelegramUsername = user.TelegramUsername[1:]
		}

		users = append(users, user)
	}

	return users, nil
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
} 