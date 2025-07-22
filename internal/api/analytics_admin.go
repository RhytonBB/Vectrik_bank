package api

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
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
	userIDStr := r.URL.Query().Get("user_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	var filterUserID uint
	if userIDStr != "" {
		if id, err := strconv.Atoi(userIDStr); err == nil {
			filterUserID = uint(id)
		}
	}
	var fromTime, toTime time.Time
	if fromStr != "" {
		fromTime, _ = time.Parse("2006-01-02", fromStr)
	}
	if toStr != "" {
		toTime, _ = time.Parse("2006-01-02", toStr)
		toTime = toTime.Add(24 * time.Hour)
	}
	users, _ := services.GetUsersList()
	spends, _ := services.GetUserSpends(filterUserID, fromTime, toTime)
	labels := []string{}
	data := []int{}
	for _, s := range spends {
		labels = append(labels, s.FirstName+" "+s.LastName)
		data = append(data, -s.Total)
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
		FilterFrom:   fromStr,
		FilterTo:     toStr,
		ChartLabels:  template.JS(labelsJSON),
		ChartData:    template.JS(dataJSON),
	})
}

// API: Получить аналитику в JSON
func AdminAnalyticsAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	var fromTime, toTime time.Time
	if fromStr != "" {
		fromTime, _ = time.Parse("2006-01-02", fromStr)
	}
	if toStr != "" {
		toTime, _ = time.Parse("2006-01-02", toStr)
		toTime = toTime.Add(24 * time.Hour)
	}
	data, err := services.GetAnalyticsData(fromTime, toTime)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Вспомогательная функция для динамического SQL
func ifCond(cond bool, sql string, args ...interface{}) string {
	if cond {
		return sql
	}
	return ""
} 