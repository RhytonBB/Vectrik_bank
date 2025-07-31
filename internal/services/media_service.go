package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RhytonBB/vectrik-webbank/internal/config"
)

type MediaService struct {
	config *config.Config
}

var mediaService *MediaService

func InitMediaService(cfg *config.Config) {
	mediaService = &MediaService{
		config: cfg,
	}
}

func GetMediaService() *MediaService {
	return mediaService
}

// UploadMedia загружает медиафайл в зависимости от настроек
func (ms *MediaService) UploadMedia(fileHeader *multipart.FileHeader, folder string) (string, error) {
	switch ms.config.MediaStorageType {
	case "cloudinary":
		return ms.uploadToCloudinary(fileHeader, folder)
	case "local":
		fallthrough
	default:
		return ms.uploadToLocal(fileHeader, folder)
	}
}

// uploadToLocal загружает файл локально
func (ms *MediaService) uploadToLocal(fileHeader *multipart.FileHeader, folder string) (string, error) {
	// Создаем уникальное имя файла
	ext := filepath.Ext(fileHeader.Filename)
	fileName := fmt.Sprintf("%s_%d%s", folder, time.Now().UnixNano(), ext)
	
	// Создаем директорию если её нет
	uploadDir := "internal/media"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("ошибка создания директории: %v", err)
	}
	
	// Создаем файл
	filePath := filepath.Join(uploadDir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания файла: %v", err)
	}
	defer dst.Close()
	
	// Открываем исходный файл
	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer src.Close()
	
	// Копируем содержимое
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("ошибка копирования файла: %v", err)
	}
	
	// Возвращаем URL для доступа к файлу
	return ms.config.MediaBaseURL + fileName, nil
}

// uploadToCloudinary загружает файл в Cloudinary
func (ms *MediaService) uploadToCloudinary(fileHeader *multipart.FileHeader, folder string) (string, error) {
	// Если Cloudinary не настроен, используем локальное хранение
	if ms.config.CloudinaryURL == "" {
		return ms.uploadToLocal(fileHeader, folder)
	}
	
	// TODO: Реализовать загрузку в Cloudinary
	// Пока используем локальное хранение как fallback
	return ms.uploadToLocal(fileHeader, folder)
}

// GetMediaURL возвращает полный URL для медиафайла
func (ms *MediaService) GetMediaURL(path string) string {
	if strings.HasPrefix(path, "http") {
		return path // Уже полный URL
	}
	
	if strings.HasPrefix(path, "/") {
		path = path[1:] // Убираем начальный слеш
	}
	
	switch ms.config.MediaStorageType {
	case "cloudinary":
		// TODO: Вернуть Cloudinary URL
		return ms.config.MediaBaseURL + path
	case "local":
		fallthrough
	default:
		return ms.config.MediaBaseURL + path
	}
}

// DeleteMedia удаляет медиафайл
func (ms *MediaService) DeleteMedia(path string) error {
	switch ms.config.MediaStorageType {
	case "cloudinary":
		// TODO: Удалить из Cloudinary
		return nil
	case "local":
		fallthrough
	default:
		// Убираем префикс пути если есть
		if strings.HasPrefix(path, "/internal/media/") {
			path = strings.TrimPrefix(path, "/internal/media/")
		}
		
		filePath := filepath.Join("internal/media", path)
		return os.Remove(filePath)
	}
} 