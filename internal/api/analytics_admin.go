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
<<<<<<< HEAD
=======

>>>>>>> 7259f56d86e51521de5028adf64898f696632916
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
<<<<<<< HEAD
	data, err := services.GetAnalyticsData(fromTime, toTime)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
=======

	// Всего пользователей
	var totalUsers int64
	db.DB.Model(&db.User{}).Count(&totalUsers)
	// Активных
	var activeUsers int64
	db.DB.Model(&db.User{}).Where("is_active = ?", true).Count(&activeUsers)
	// Забаненных
	var bannedUsers int64
	db.DB.Model(&db.User{}).Where("is_banned = ?", true).Count(&bannedUsers)
	// Общий баланс
	var totalBalance int64
	db.DB.Model(&db.User{}).Select("SUM(balance)").Scan(&totalBalance)
	// Новые пользователи за период
	var newUsers int64
	if !fromTime.IsZero() && !toTime.IsZero() {
		db.DB.Model(&db.User{}).Where("registered_at >= ? AND registered_at < ?", fromTime, toTime).Count(&newUsers)
	}
	// Сумма трат (amount < 0)
	var totalSpend int64
	db.DB.Model(&db.Transaction{}).Select("SUM(amount)").Where("amount < 0").Scan(&totalSpend)
	// Сумма начислений (amount > 0)
	var totalReward int64
	db.DB.Model(&db.Transaction{}).Select("SUM(amount)").Where("amount > 0").Scan(&totalReward)

	// Динамика по дням (трат и начислений)
	type dayStat struct {
		Date  string
		Spend int64
		Reward int64
	}
	var dayStats []dayStat
	db.DB.Raw(`
		SELECT to_char(created_at, 'YYYY-MM-DD') as date,
		SUM(CASE WHEN amount < 0 THEN amount ELSE 0 END) as spend,
		SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END) as reward
		FROM transactions
		WHERE 1=1
		` +
		ifCond(!fromTime.IsZero(), " AND created_at >= ?", fromTime) +
		ifCond(!toTime.IsZero(), " AND created_at < ?", toTime) +
		` GROUP BY date ORDER BY date`,
		).Scan(&dayStats)

	// Топ-10 по тратам
	type topUser struct {
		UserID    uint   `json:"user_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Total     int64  `json:"total"`
	}
	var topSpenders []topUser
	db.DB.Table("transactions").
		Select("users.id as user_id, users.first_name, users.last_name, SUM(transactions.amount) as total").
		Joins("JOIN users ON users.id = transactions.user_id").
		Where("transactions.amount < 0").
		Group("users.id, users.first_name, users.last_name").
		Order("total").Limit(10).Scan(&topSpenders)
	// Топ-10 по начислениям
	var topRewarded []topUser
	db.DB.Table("transactions").
		Select("users.id as user_id, users.first_name, users.last_name, SUM(transactions.amount) as total").
		Joins("JOIN users ON users.id = transactions.user_id").
		Where("transactions.amount > 0").
		Group("users.id, users.first_name, users.last_name").
		Order("total desc").Limit(10).Scan(&topRewarded)
	// Последние 10 транзакций
	type lastTx struct {
		ID        uint      `json:"id"`
		UserID    uint      `json:"user_id"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		Amount    int       `json:"amount"`
		Reason    string    `json:"reason"`
		Source    string    `json:"source"`
		CreatedAt time.Time `json:"created_at"`
	}
	var lastTransactions []lastTx
	db.DB.Table("transactions").
		Select("transactions.id, users.id as user_id, users.first_name, users.last_name, transactions.amount, transactions.reason, transactions.source, transactions.created_at").
		Joins("JOIN users ON users.id = transactions.user_id").
		Order("transactions.created_at desc").Limit(10).Scan(&lastTransactions)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_users": totalUsers,
		"active_users": activeUsers,
		"banned_users": bannedUsers,
		"total_balance": totalBalance,
		"new_users": newUsers,
		"total_spend": totalSpend,
		"total_reward": totalReward,
		"day_stats": dayStats,
		"top_spenders": topSpenders,
		"top_rewarded": topRewarded,
		"last_transactions": lastTransactions,
	})
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
}

// Вспомогательная функция для динамического SQL
func ifCond(cond bool, sql string, args ...interface{}) string {
	if cond {
		return sql
	}
	return ""
} 