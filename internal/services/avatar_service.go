package services

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"golang.org/x/image/draw"
)

// UploadUserAvatar загружает и обрабатывает аватарку пользователя
func UploadUserAvatar(userID uint, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", errors.New("файл не выбран")
	}

	// Проверяем тип файла
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", errors.New("поддерживаются только изображения в форматах JPG, PNG")
	}

	// Проверяем размер файла (максимум 10MB)
	if fileHeader.Size > 10*1024*1024 {
		return "", errors.New("размер файла не должен превышать 10MB")
	}

	// Открываем файл
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	// Декодируем изображение
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования изображения: %v", err)
	}

	// Обрезаем изображение до квадрата
	squareImg := cropToSquare(img)

	// Изменяем размер до 200x200 пикселей
	resizedImg := resizeImage(squareImg, 200, 200)

	// Генерируем имя файла
	fileName := fmt.Sprintf("avatar_%d_%d%s", userID, time.Now().UnixNano(), ext)
	filePath := filepath.Join("internal/media/avatars", fileName)

	// Создаем директорию если не существует
	os.MkdirAll("internal/media/avatars", 0755)

	// Сохраняем обработанное изображение
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания файла: %v", err)
	}
	defer out.Close()

	// Кодируем в зависимости от расширения
	switch ext {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(out, resizedImg, &jpeg.Options{Quality: 90})
	case ".png":
		err = png.Encode(out, resizedImg)
	default:
		err = jpeg.Encode(out, resizedImg, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return "", fmt.Errorf("ошибка кодирования изображения: %v", err)
	}

	// Возвращаем URL для сохранения в базе
	return "/internal/media/avatars/" + fileName, nil
}

// cropToSquare обрезает изображение до квадрата
func cropToSquare(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Определяем размер стороны квадрата (минимальная из ширины и высоты)
	size := width
	if height < width {
		size = height
	}

	// Вычисляем координаты для обрезки (центрируем)
	x := (width - size) / 2
	y := (height - size) / 2

	// Создаем новое изображение
	square := image.NewRGBA(image.Rect(0, 0, size, size))

	// Копируем центральную часть
	draw.Draw(square, square.Bounds(), img, image.Pt(x, y), draw.Over)

	return square
}

// resizeImage изменяет размер изображения
func resizeImage(img image.Image, targetWidth, targetHeight int) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// Создаем новое изображение с целевыми размерами
	resized := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// Используем высококачественное масштабирование
	draw.ApproxBiLinear.Scale(resized, resized.Bounds(), img, image.Rect(0, 0, srcWidth, srcHeight), draw.Over, nil)

	return resized
}

// DeleteUserAvatar удаляет аватарку пользователя
func DeleteUserAvatar(avatarURL string) error {
	if avatarURL == "" {
		return nil
	}

	// Убираем ведущий слеш
	filePath := strings.TrimPrefix(avatarURL, "/")
	
	// Проверяем, что путь безопасный
	if !strings.HasPrefix(filePath, "internal/media/avatars/") {
		return errors.New("неверный путь к файлу")
	}

	// Удаляем файл
	return os.Remove(filePath)
}

// GetDefaultAvatar возвращает URL дефолтной аватарки
func GetDefaultAvatar() string {
	return "/default-avatar.png"
}
