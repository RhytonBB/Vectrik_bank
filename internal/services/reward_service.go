package services

import (
	"errors"
	"time"
	"strings"
	"fmt"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"gorm.io/gorm"
	"golang.org/x/text/unicode/norm"
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

func GetValidRewardReasons() map[string]bool {
	return map[string]bool{
		"+5 за активность в группе Вектора": true,
		"+10 за отличный вопрос или активность на блоке": true,
		"+15 за конструктивную ОС": true,
		"+20 за участие в задании на ИС по своей инициативе": true,
		"+25 за отличную игротехнику": true,
		"+30 за победу в соревнованиях": true,
		"+35 за что-то внезапное, креативное и хорошее": true,
		"+40 за преодоление себя": true,
		"+50 за высокий результат на экзамене": true,
		"+245 за прохождение вузовской Школы кураторов": true,
	}
}

func RewardUsers(userIDs []uint, amount int, reason, context string, adminID uint) error {
	reason = strings.TrimSpace(reason)
	reason = norm.NFC.String(reason)
	fmt.Printf("REWARD REASON: '%s' (len=%d, bytes=%v)\n", reason, len(reason), []byte(reason))
	if len(userIDs) == 0 || amount <= 0 || reason == "" {
		return errors.New("Все поля обязательны")
	}
	validReasons := GetValidRewardReasons()
	for k := range validReasons {
		fmt.Printf("VALID: '%s' (len=%d, bytes=%v)\n", k, len(k), []byte(k))
	}
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