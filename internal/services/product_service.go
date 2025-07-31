package services

import (
	"errors"
	"fmt"
	"os"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
)

func CreateProduct(product *db.Product) error {
	if product.Name == "" || product.Price <= 0 || product.Type == "" || product.Stock < 0 {
		return errors.New("Заполните все обязательные поля корректно")
	}
	product.CreatedAt = time.Now()
	return db.DB.Create(product).Error
}

func GetProductByID(productID uint) (*db.Product, error) {
	var product db.Product
	if err := db.DB.First(&product, productID).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func GetAllProducts() ([]db.Product, error) {
	var products []db.Product
	if err := db.DB.Order("created_at desc").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func UpdateProduct(productID uint, updates map[string]interface{}) error {
	return db.DB.Model(&db.Product{}).Where("id = ?", productID).Updates(updates).Error
}

func DeleteProduct(productID uint) error {
	return db.DB.Delete(&db.Product{}, productID).Error
}

func ActivateProduct(productID uint) error {
	return db.DB.Model(&db.Product{}).Where("id = ?", productID).Update("is_active", true).Error
}

func DeactivateProduct(productID uint) error {
	return db.DB.Model(&db.Product{}).Where("id = ?", productID).Update("is_active", false).Error
}

func GetActiveProducts() ([]db.Product, error) {
	var products []db.Product
	err := db.DB.Where("is_active = ?", true).Order("created_at desc").Find(&products).Error
	return products, err
}

func PurchaseProduct(userID, productID uint, quantity int) error {
	tx := db.DB.Begin()
	
	// Получаем продукт
	var product db.Product
	if err := tx.First(&product, productID).Error; err != nil {
		tx.Rollback()
		return errors.New("Товар не найден")
	}
	
	if !product.IsActive {
		tx.Rollback()
		return errors.New("Товар недоступен")
	}
	
	if product.Stock < quantity {
		tx.Rollback()
		return errors.New("Недостаточно товара на складе")
	}
	
	// Получаем пользователя
	var user db.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		return errors.New("Пользователь не найден")
	}
	
	totalCost := product.Price * quantity
	
	if user.Balance < totalCost {
		tx.Rollback()
		return errors.New("Недостаточно средств")
	}
	
	// Создаем покупку
	purchase := db.Purchase{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Confirmed: true,
		CreatedAt: time.Now(),
	}
	
	if err := tx.Create(&purchase).Error; err != nil {
		tx.Rollback()
		return errors.New("Ошибка создания покупки")
	}
	
	// Создаем транзакцию
	transaction := db.Transaction{
		UserID:    userID,
		Amount:    -totalCost,
		Reason:    fmt.Sprintf("Покупка: %s", product.Name),
		Source:    "purchase",
		CreatedAt: time.Now(),
	}
	
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return errors.New("Ошибка создания транзакции")
	}
	
	// Обновляем баланс пользователя
	if err := tx.Model(&user).UpdateColumn("balance", user.Balance-totalCost).Error; err != nil {
		tx.Rollback()
		return errors.New("Ошибка обновления баланса")
	}
	
	// Обновляем остаток товара
	if err := tx.Model(&product).UpdateColumn("stock", product.Stock-quantity).Error; err != nil {
		tx.Rollback()
		return errors.New("Ошибка обновления остатка товара")
	}
	
	return tx.Commit().Error
}

func UploadProductMedia(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", fmt.Errorf("Файл не выбран")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("Ошибка открытия файла: %v", err)
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	fileName := fmt.Sprintf("product_%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join("internal/media", fileName)
	os.MkdirAll("internal/media", 0755)
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("Ошибка сохранения файла: %v", err)
	}
	defer out.Close()
	if _, err = io.Copy(out, file); err != nil {
		return "", fmt.Errorf("Ошибка копирования файла: %v", err)
	}
	return "/internal/media/" + fileName, nil
} 