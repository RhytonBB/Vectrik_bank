package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/xuri/excelize/v2"
)

type BulkUser struct {
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	TelegramUsername string `json:"telegram_username"`
	GroupID          *uint  `json:"group_id"`
}

type BulkPreviewResponse struct {
	Users []BulkUser `json:"users"`
}

type BulkRegisterResponse struct {
	Created int
	Errors  []string
}

type BulkRegisterRequest struct {
	Users []BulkUser `json:"users"`
}

func BulkPreviewUsers(users []BulkUser) BulkPreviewResponse {
	var validUsers []BulkUser
	for _, user := range users {
		if user.FirstName != "" && user.LastName != "" && user.TelegramUsername != "" {
			validUsers = append(validUsers, user)
		}
	}
	return BulkPreviewResponse{Users: validUsers}
}

func BulkRegisterUsers(users []BulkUser) BulkRegisterResponse {
	created := 0
	errors := []string{}
	for i, user := range users {
		if user.FirstName == "" || user.LastName == "" || user.TelegramUsername == "" {
			errors = append(errors, fmt.Sprintf("Row %d: missing required fields", i+1))
			continue
		}
		var existingUser db.User
		if err := db.DB.Where("telegram_username = ?", user.TelegramUsername).First(&existingUser).Error; err == nil {
			errors = append(errors, fmt.Sprintf("Row %d: telegram username '%s' already exists", i+1, user.TelegramUsername))
			continue
		}
		newUser := db.User{
			FirstName:        user.FirstName,
			LastName:         user.LastName,
			TelegramUsername: user.TelegramUsername,
			IsActive:         true,
			GroupID:          user.GroupID,
		}
		if err := db.DB.Create(&newUser).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: %s", i+1, err.Error()))
			continue
		}
		created++
	}
	return BulkRegisterResponse{Created: created, Errors: errors}
}

func ParseCSVFile(file io.Reader) ([]BulkUser, error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("file must contain at least header and one data row")
	}
	headers := records[0]
	firstNameIdx, lastNameIdx, telegramIdx, groupIdx := -1, -1, -1, -1
	for i, header := range headers {
		h := strings.ToLower(strings.TrimSpace(header))
		switch h {
		case "имя", "first name", "firstname", "name":
			firstNameIdx = i
		case "фамилия", "last name", "lastname", "surname":
			lastNameIdx = i
		case "telegram", "telegram username", "username", "tg", "telegram_username":
			telegramIdx = i
		case "group_id", "группа", "group":
			groupIdx = i
		}
	}
	if firstNameIdx == -1 || lastNameIdx == -1 || telegramIdx == -1 {
		return nil, fmt.Errorf("required columns not found. Expected: Имя/FirstName, Фамилия/LastName, Telegram/Username")
	}
	var users []BulkUser
	for _, record := range records[1:] {
		if len(record) <= firstNameIdx || len(record) <= lastNameIdx || len(record) <= telegramIdx {
			continue
		}
		var groupID *uint
		if groupIdx != -1 && len(record) > groupIdx {
			groupStr := strings.TrimSpace(record[groupIdx])
			if groupStr != "" {
				var gid uint
				_, err := fmt.Sscanf(groupStr, "%d", &gid)
				if err == nil {
					groupID = &gid
				}
			}
		}
		user := BulkUser{
			FirstName:        strings.TrimSpace(record[firstNameIdx]),
			LastName:         strings.TrimSpace(record[lastNameIdx]),
			TelegramUsername: strings.TrimSpace(record[telegramIdx]),
			GroupID:          groupID,
		}
		if strings.HasPrefix(user.TelegramUsername, "@") {
			user.TelegramUsername = user.TelegramUsername[1:]
		}
		users = append(users, user)
	}
	return users, nil
}

func ParseExcelFile(file io.Reader) ([]BulkUser, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
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
	headers := rows[0]
	firstNameIdx, lastNameIdx, telegramIdx, groupIdx := -1, -1, -1, -1
	for i, header := range headers {
		h := strings.ToLower(strings.TrimSpace(header))
		switch h {
		case "имя", "first name", "firstname", "name":
			firstNameIdx = i
		case "фамилия", "last name", "lastname", "surname":
			lastNameIdx = i
		case "telegram", "telegram username", "username", "tg", "telegram_username":
			telegramIdx = i
		case "group_id", "группа", "group":
			groupIdx = i
		}
	}
	if firstNameIdx == -1 || lastNameIdx == -1 || telegramIdx == -1 {
		return nil, fmt.Errorf("required columns not found. Expected: Имя/FirstName, Фамилия/LastName, Telegram/Username")
	}
	var users []BulkUser
	for _, row := range rows[1:] {
		if len(row) <= firstNameIdx || len(row) <= lastNameIdx || len(row) <= telegramIdx {
			continue
		}
		var groupID *uint
		if groupIdx != -1 && len(row) > groupIdx {
			groupStr := strings.TrimSpace(row[groupIdx])
			if groupStr != "" {
				var gid uint
				_, err := fmt.Sscanf(groupStr, "%d", &gid)
				if err == nil {
					groupID = &gid
				}
			}
		}
		user := BulkUser{
			FirstName:        strings.TrimSpace(row[firstNameIdx]),
			LastName:         strings.TrimSpace(row[lastNameIdx]),
			TelegramUsername: strings.TrimSpace(row[telegramIdx]),
			GroupID:          groupID,
		}
		if strings.HasPrefix(user.TelegramUsername, "@") {
			user.TelegramUsername = user.TelegramUsername[1:]
		}
		users = append(users, user)
	}
	return users, nil
} 