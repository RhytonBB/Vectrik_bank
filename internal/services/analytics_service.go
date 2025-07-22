package services

import (
	"log"
	"time"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
)

type UserSpend struct {
	UserID    uint
	FirstName string
	LastName  string
	Total     int
}

type DayStat struct {
	Date   string `json:"date"`
	Spend  int64  `json:"spend"`
	Reward int64  `json:"reward"`
}

type TopUser struct {
	UserID    uint   `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Total     int64  `json:"total"`
}

type LastTx struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Amount    int       `json:"amount"`
	Reason    string    `json:"reason"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

type AnalyticsData struct {
	TotalUsers      int64      `json:"total_users"`
	ActiveUsers     int64      `json:"active_users"`
	BannedUsers     int64      `json:"banned_users"`
	TotalBalance    int64      `json:"total_balance"`
	NewUsers        int64      `json:"new_users"`
	TotalSpend      int64      `json:"total_spend"`
	TotalReward     int64      `json:"total_reward"`
	DayStats        []DayStat  `json:"day_stats"`
	TopSpenders     []TopUser  `json:"top_spenders"`
	TopRewarded     []TopUser  `json:"top_rewarded"`
	LastTransactions []LastTx  `json:"last_transactions"`
}

func GetUsersList() ([]db.User, error) {
	var users []db.User
	err := db.DB.Order("first_name").Find(&users).Error
	if err != nil {
		log.Printf("[Analytics] Ошибка выборки пользователей: %v", err)
		return nil, err
	}
	return users, nil
}

func GetUserSpends(filterUserID uint, fromTime, toTime time.Time) ([]UserSpend, error) {
	var spends []UserSpend
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
	if err := query.Scan(&spends).Error; err != nil {
		log.Printf("[Analytics] Ошибка выборки трат: %v", err)
		return nil, err
	}
	return spends, nil
}

func GetAnalyticsData(fromTime, toTime time.Time) (AnalyticsData, error) {
	var data AnalyticsData
	db.DB.Model(&db.User{}).Count(&data.TotalUsers)
	db.DB.Model(&db.User{}).Where("is_active = ?", true).Count(&data.ActiveUsers)
	db.DB.Model(&db.User{}).Where("is_banned = ?", true).Count(&data.BannedUsers)
	db.DB.Model(&db.User{}).Select("SUM(balance)").Scan(&data.TotalBalance)
	if !fromTime.IsZero() && !toTime.IsZero() {
		db.DB.Model(&db.User{}).Where("registered_at >= ? AND registered_at < ?", fromTime, toTime).Count(&data.NewUsers)
	}
	db.DB.Model(&db.Transaction{}).Select("SUM(amount)").Where("amount < 0").Scan(&data.TotalSpend)
	db.DB.Model(&db.Transaction{}).Select("SUM(amount)").Where("amount > 0").Scan(&data.TotalReward)
	// Динамика по дням
	var dayStats []DayStat
	db.DB.Raw(`SELECT to_char(created_at, 'YYYY-MM-DD') as date, SUM(CASE WHEN amount < 0 THEN amount ELSE 0 END) as spend, SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END) as reward FROM transactions WHERE 1=1`+
		ifCond(!fromTime.IsZero(), " AND created_at >= ?", fromTime)+
		ifCond(!toTime.IsZero(), " AND created_at < ?", toTime)+
		` GROUP BY date ORDER BY date`).Scan(&dayStats)
	data.DayStats = dayStats
	// Топ-10 по тратам
	var topSpenders []TopUser
	db.DB.Table("transactions").
		Select("users.id as user_id, users.first_name, users.last_name, SUM(transactions.amount) as total").
		Joins("JOIN users ON users.id = transactions.user_id").
		Where("transactions.amount < 0").
		Group("users.id, users.first_name, users.last_name").
		Order("total").Limit(10).Scan(&topSpenders)
	data.TopSpenders = topSpenders
	// Топ-10 по начислениям
	var topRewarded []TopUser
	db.DB.Table("transactions").
		Select("users.id as user_id, users.first_name, users.last_name, SUM(transactions.amount) as total").
		Joins("JOIN users ON users.id = transactions.user_id").
		Where("transactions.amount > 0").
		Group("users.id, users.first_name, users.last_name").
		Order("total desc").Limit(10).Scan(&topRewarded)
	data.TopRewarded = topRewarded
	// Последние 10 транзакций
	var lastTransactions []LastTx
	db.DB.Table("transactions").
		Select("transactions.id, users.id as user_id, users.first_name, users.last_name, transactions.amount, transactions.reason, transactions.source, transactions.created_at").
		Joins("JOIN users ON users.id = transactions.user_id").
		Order("transactions.created_at desc").Limit(10).Scan(&lastTransactions)
	data.LastTransactions = lastTransactions
	return data, nil
}

func ifCond(cond bool, sql string, args ...interface{}) string {
	if cond {
		return sql
	}
	return ""
} 