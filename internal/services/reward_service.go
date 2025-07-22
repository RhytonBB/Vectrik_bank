package services

import (
	"errors"
	"time"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"gorm.io/gorm"
)

func RewardSearchUsers(query string) ([]db.User, error) {
	var users []db.User
	tx := db.DB.Limit(20).Order("first_name")
	if query != "" {
		likeQuery := "%" + query + "%"
		tx = tx.Where(
			"first_name ILIKE ? OR last_name ILIKE ? OR telegram_username ILIKE ?",
			likeQuery, likeQuery, likeQuery,
		)
	}
	if err := tx.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func RewardUsers(userIDs []uint, amount int, reason, context string, adminID uint) error {
	if len(userIDs) == 0 || amount <= 0 || reason == "" {
		return errors.New("Все поля обязательны")
	}
	validReasons := map[string]bool{"drop": true, "активность": true, "награда": true}
	if !validReasons[reason] {
		return errors.New("Некорректная причина")
	}
	tx := db.DB.Begin()
	for _, userID := range userIDs {
		transaction := db.Transaction{
			UserID:    userID,
			Amount:    amount,
			Reason:    reason,
			Source:    "reward",
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&transaction).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(&db.User{}).
			Where("id = ?", userID).
			UpdateColumn("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			tx.Rollback()
			return err
		}
		adminLog := db.AdminLog{
			AdminID:      adminID,
			TargetUserID: userID,
			Action:       "reward",
			Amount:       &amount,
			CreatedAt:    time.Now(),
		}
		if context != "" {
			adminLog.Context = &context
		}
		if err := tx.Create(&adminLog).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
} 