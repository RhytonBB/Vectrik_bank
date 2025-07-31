package services

import (
	"errors"
	"fmt"
	"os"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"strconv"
	"time"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
)

func CreateBanner(banner *db.Banner) error {
	if banner.Title == "" || banner.ImageURL == "" {
		return errors.New("Заполните все обязательные поля")
	}
	banner.CreatedAt = time.Now()
	return db.DB.Create(banner).Error
}

func GetBannerByID(bannerID uint) (*db.Banner, error) {
	var banner db.Banner
	if err := db.DB.First(&banner, bannerID).Error; err != nil {
		return nil, err
	}
	return &banner, nil
}

func GetAllBanners() ([]db.Banner, error) {
	var banners []db.Banner
	if err := db.DB.Order("created_at desc").Find(&banners).Error; err != nil {
		return nil, err
	}
	return banners, nil
}

func UpdateBanner(bannerID uint, updates map[string]interface{}) error {
	return db.DB.Model(&db.Banner{}).Where("id = ?", bannerID).Updates(updates).Error
}

func DeleteBanner(bannerID uint) error {
	return db.DB.Delete(&db.Banner{}, bannerID).Error
}

func ActivateBanner(bannerID uint) error {
	return db.DB.Model(&db.Banner{}).Where("id = ?", bannerID).Update("is_active", true).Error
}

func DeactivateBanner(bannerID uint) error {
	return db.DB.Model(&db.Banner{}).Where("id = ?", bannerID).Update("is_active", false).Error
}

func GetActiveBanners() ([]db.Banner, error) {
	var banners []db.Banner
	err := db.DB.Where("is_active = ?", true).Order("created_at desc").Find(&banners).Error
	return banners, err
}

func UploadAndCreateBanner(title string, fileHeader *multipart.FileHeader) (*db.Banner, error) {
	if title == "" || fileHeader == nil {
		return nil, fmt.Errorf("Заполните все поля и выберите файл")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("Ошибка открытия файла: %v", err)
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	fileName := fmt.Sprintf("banner_%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join("internal/media", fileName)
	os.MkdirAll("internal/media", 0755)
	out, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("Ошибка сохранения файла: %v", err)
	}
	defer out.Close()
	if _, err = io.Copy(out, file); err != nil {
		return nil, fmt.Errorf("Ошибка копирования файла: %v", err)
	}
	banner := db.Banner{
		Title:    title,
		ImageURL: "/internal/media/" + fileName,
		CreatedAt: time.Now(),
		IsActive: true,
	}
	if err := db.DB.Create(&banner).Error; err != nil {
		return nil, fmt.Errorf("Ошибка сохранения баннера: %v", err)
	}
	return &banner, nil
}

func DeleteBannerByID(idStr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}
	var banner db.Banner
	if err := db.DB.First(&banner, id).Error; err != nil {
		return err
	}
	if banner.ImageURL != "" {
		os.Remove(strings.TrimPrefix(banner.ImageURL, "/"))
	}
	return db.DB.Delete(&banner).Error
}

func ToggleBannerByID(idStr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}
	var banner db.Banner
	if err := db.DB.First(&banner, id).Error; err != nil {
		return err
	}
	return db.DB.Model(&banner).Update("is_active", !banner.IsActive).Error
}

func SetBannerActiveStatus(idStr string, isActive bool) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}
	return db.DB.Model(&db.Banner{}).Where("id = ?", id).Update("is_active", isActive).Error
}

func UpdateBannerByID(idStr string, title string, imageURL *string, isActive *bool) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return err
	}
	updates := map[string]interface{}{}
	if title != "" {
		updates["title"] = title
	}
	if imageURL != nil {
		updates["image_url"] = imageURL
	}
	if isActive != nil {
		updates["is_active"] = isActive
	}
	return db.DB.Model(&db.Banner{}).Where("id = ?", id).Updates(updates).Error
} 