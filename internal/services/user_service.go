package services

import (
	"errors"
	"time"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"gorm.io/gorm"
)

func GetUserByID(userID uint) (*db.User, error) {
	var user db.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetAllUsers() ([]db.User, error) {
	var users []db.User
	if err := db.DB.Order("first_name").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func CreateUser(user *db.User) error {
	if user.FirstName == "" || user.LastName == "" || user.TelegramUsername == "" {
		return errors.New("Имя, фамилия и Telegram username обязательны")
	}
	return db.DB.Create(user).Error
}

func UpdateUser(userID uint, updates map[string]interface{}) error {
	return db.DB.Model(&db.User{}).Where("id = ?", userID).Updates(updates).Error
}

func DeleteUser(userID uint) error {
	return db.DB.Delete(&db.User{}, userID).Error
}

func BanUser(userID uint) error {
	return db.DB.Model(&db.User{}).Where("id = ?", userID).Update("is_banned", true).Error
}

func UnbanUser(userID uint) error {
	return db.DB.Model(&db.User{}).Where("id = ?", userID).Update("is_banned", false).Error
}

func SetUserAdmin(userID uint, isAdmin bool) error {
	return db.DB.Model(&db.User{}).Where("id = ?", userID).Update("is_admin", isAdmin).Error
}

func ActivateUser(userID uint) error {
	return db.DB.Model(&db.User{}).Where("id = ?", userID).Update("is_active", true).Error
}

func DeactivateUser(userID uint) error {
	return db.DB.Model(&db.User{}).Where("id = ?", userID).Update("is_active", false).Error
}

func GetUserPurchases(userID uint) ([]db.Purchase, error) {
	var purchases []db.Purchase
	err := db.DB.Preload("Product").Where("user_id = ?", userID).Find(&purchases).Error
	return purchases, err
}

func GetUserTransactions(userID uint) ([]db.Transaction, error) {
	var transactions []db.Transaction
	err := db.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&transactions).Error
	return transactions, err
}

func AdjustUserBalance(userID uint, amount int, context string, adminID uint) error {
	if amount == 0 {
		return gorm.ErrInvalidData
	}
	tx := db.DB.Begin()
	transaction := db.Transaction{
		UserID:    userID,
		Amount:    amount,
		Reason:    "adjustment",
		Source:    "adjustment",
		CreatedAt: time.Now(),
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&db.User{}).Where("id = ?", userID).UpdateColumn("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
		tx.Rollback()
		return err
	}
	adminLog := db.AdminLog{
		AdminID:      adminID,
		TargetUserID: userID,
		Action:       "adjustment",
		Amount:       &amount,
		Context:      &context,
		CreatedAt:    time.Now(),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
} 