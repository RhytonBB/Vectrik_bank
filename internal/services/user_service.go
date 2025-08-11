package services

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"gorm.io/gorm"
)

func GetUserByID(userID uint) (*db.User, error) {
	var user db.User
	if err := db.DB.Preload("Group").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetAllUsers() ([]db.User, error) {
	var users []db.User
	if err := db.DB.Preload("Group").Order("first_name").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Генерирует уникальный 8-значный номер счета
func generateAccountNumber() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%08d", rand.Intn(100000000))
}

func CreateUser(user *db.User) error {
	if user.FirstName == "" || user.LastName == "" || user.TelegramUsername == "" {
		return errors.New("Имя, фамилия и Telegram username обязательны")
	}
	
	// Генерируем уникальный номер счета
	for {
		accountNumber := generateAccountNumber()
		var existingUser db.User
		if err := db.DB.Where("account_number = ?", accountNumber).First(&existingUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				user.AccountNumber = accountNumber
				break
			}
			return err
		}
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

func GetGroupUsersByAdmin(adminID uint) ([]db.User, error) {
    var groupAdmin db.GroupAdmin
    if err := db.DB.Where("admin_id = ?", adminID).First(&groupAdmin).Error; err != nil {
        return []db.User{}, nil // Нет группы — пустой список
    }
    var users []db.User
    db.DB.Where("group_id = ?", groupAdmin.GroupID).Find(&users)
    return users, nil
}

func GetUserByLogin(login string) (*db.User, error) {
    var user db.User
    if err := db.DB.Where("login = ?", login).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

func GetUserBalance(userID uint) (int, error) {
    var user db.User
    if err := db.DB.Select("balance").Where("id = ?", userID).First(&user).Error; err != nil {
        return 0, err
    }
    return user.Balance, nil
}

func GetUserShortInfo(userID uint) (*db.User, error) {
    var user db.User
    if err := db.DB.Select("id, first_name, last_name, is_admin, balance, account_number, telegram_username").Where("id = ?", userID).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

func CreateGroupWithAdmins(name string, adminIDs []uint) (*db.Group, error) {
    if name == "" {
        return nil, errors.New("Название группы обязательно")
    }
    group := db.Group{Name: name}
    if err := db.DB.Create(&group).Error; err != nil {
        return nil, err
    }
    for _, adminID := range adminIDs {
        db.DB.Create(&db.GroupAdmin{GroupID: group.ID, AdminID: adminID})
    }
    return &group, nil
}

func DeleteGroup(groupID uint) error {
    tx := db.DB.Begin()
    
    // Удаляем связи администраторов с группой
    if err := tx.Where("group_id = ?", groupID).Delete(&db.GroupAdmin{}).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // Обновляем пользователей, убирая ссылку на группу
    if err := tx.Model(&db.User{}).Where("group_id = ?", groupID).Update("group_id", nil).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // Удаляем саму группу
    if err := tx.Delete(&db.Group{}, groupID).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit().Error
} 