package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
)

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