package api

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
)

func AdminAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Фильтры
	userIDStr := r.URL.Query().Get("user_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var filterUserID uint
	if userIDStr != "" {
		if id, err := strconv.Atoi(userIDStr); err == nil {
			filterUserID = uint(id)
		}
	}
	var filterFrom, filterTo string
	var fromTime, toTime time.Time
	if fromStr != "" {
		filterFrom = fromStr
		fromTime, _ = time.Parse("2006-01-02", fromStr)
	}
	if toStr != "" {
		filterTo = toStr
		toTime, _ = time.Parse("2006-01-02", toStr)
		toTime = toTime.Add(24 * time.Hour) // включительно
	}

	// Список пользователей для фильтра
	var users []db.User
	db.DB.Order("first_name").Find(&users)

	// Агрегация: траты по пользователям

type userSpend struct {
	UserID    uint
	FirstName string
	LastName  string
	Total     int
}

	var spends []userSpend
	query := db.DB.Table("transactions").
		Select("users.id as user_id, users.first_name, users.last_name, SUM(transactions.amount) as total").
		Joins("JOIN users ON users.id = transactions.user_id").
		Where("transactions.amount < 0")

	if filterUserID != 0 {
		query = query.Where("users.id = ?", filterUserID)
	}
	if !fromTime.IsZero() {
		query = query.Where("transactions.created_at >= ?", fromTime)
	}
	if !toTime.IsZero() {
		query = query.Where("transactions.created_at < ?", toTime)
	}
	query = query.Group("users.id, users.first_name, users.last_name").Order("total desc")
	query.Scan(&spends)

	// Формируем данные для Chart.js
	labels := []string{}
	data := []int{}
	for _, s := range spends {
		labels = append(labels, s.FirstName+" "+s.LastName)
		data = append(data, -s.Total) // сумма отрицательная, делаем положительной
	}
	labelsJSON, _ := json.Marshal(labels)
	dataJSON, _ := json.Marshal(data)

	tmpl := template.Must(template.ParseFiles("web/templates/admin_analytics.html"))
	tmpl.Execute(w, struct {
		Users        []db.User
		FilterUserID uint
		FilterFrom   string
		FilterTo     string
		ChartLabels  template.JS
		ChartData    template.JS
	}{
		Users:        users,
		FilterUserID: filterUserID,
		FilterFrom:   filterFrom,
		FilterTo:     filterTo,
		ChartLabels:  template.JS(labelsJSON),
		ChartData:    template.JS(dataJSON),
	})
} 