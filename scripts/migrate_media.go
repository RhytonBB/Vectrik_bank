package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/RhytonBB/vectrik-webbank/internal/config"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()
	
	// Инициализируем базу данных
	if err := db.InitDB(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	
	// Инициализируем медиа-сервис
	services.InitMediaService(cfg)
	
	fmt.Println("Начинаем миграцию медиафайлов...")
	
	// Мигрируем баннеры
	migrateBanners()
	
	// Мигрируем товары
	migrateProducts()
	
	fmt.Println("Миграция завершена!")
}

func migrateBanners() {
	fmt.Println("Мигрируем баннеры...")
	
	var banners []db.Banner
	if err := db.DB.Find(&banners).Error; err != nil {
		log.Printf("Ошибка получения баннеров: %v", err)
		return
	}
	
	for _, banner := range banners {
		if banner.ImageURL != "" && strings.HasPrefix(banner.ImageURL, "/internal/media/") {
			fmt.Printf("Мигрируем баннер %d: %s\n", banner.ID, banner.ImageURL)
			
			// TODO: Здесь должна быть логика загрузки в облачное хранилище
			// Пока просто выводим информацию
			fmt.Printf("  Файл: %s\n", banner.ImageURL)
		}
	}
}

func migrateProducts() {
	fmt.Println("Мигрируем товары...")
	
	var products []db.Product
	if err := db.DB.Find(&products).Error; err != nil {
		log.Printf("Ошибка получения товаров: %v", err)
		return
	}
	
	for _, product := range products {
		if product.MediaURL != nil && *product.MediaURL != "" && strings.HasPrefix(*product.MediaURL, "/internal/media/") {
			fmt.Printf("Мигрируем товар %d: %s\n", product.ID, *product.MediaURL)
			
			// TODO: Здесь должна быть логика загрузки в облачное хранилище
			// Пока просто выводим информацию
			fmt.Printf("  Файл: %s\n", *product.MediaURL)
		}
	}
}

func checkLocalFiles() {
	fmt.Println("Проверяем локальные файлы...")
	
	mediaDir := "internal/media"
	if _, err := os.Stat(mediaDir); os.IsNotExist(err) {
		fmt.Println("Директория internal/media не существует")
		return
	}
	
	err := filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fmt.Printf("Найден файл: %s\n", path)
		}
		return nil
	})
	
	if err != nil {
		log.Printf("Ошибка обхода директории: %v", err)
	}
} 