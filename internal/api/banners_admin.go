package api

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
)

const mediaDir = "internal/media/"

func AdminBannersHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	isAdmin, ok := claims["is_admin"].(bool)
	if !ok || !isAdmin {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
	}

	var message, errMsg string

	if r.Method == http.MethodPost {
		r.ParseMultipartForm(10 << 20) // 10 МБ
	
action := r.FormValue("action")
		bannerID := r.FormValue("banner_id")

		if action == "delete" && bannerID != "" {
			if err := deleteBanner(bannerID); err != nil {
				errMsg = "Ошибка удаления баннера: " + err.Error()
			} else {
				message = "Баннер удалён"
			}
		} else if action == "toggle" && bannerID != "" {
			if err := toggleBanner(bannerID); err != nil {
				errMsg = "Ошибка изменения статуса: " + err.Error()
			} else {
				message = "Статус баннера изменён"
			}
		} else if r.MultipartForm != nil && r.MultipartForm.File != nil && len(r.MultipartForm.File["file"]) > 0 {
			// Загрузка нового баннера
			title := r.FormValue("title")
			fileHeader := r.MultipartForm.File["file"][0]
			if title == "" || fileHeader == nil {
				errMsg = "Заполните все поля и выберите файл"
			} else {
				file, err := fileHeader.Open()
				if err != nil {
					errMsg = "Ошибка открытия файла: " + err.Error()
				} else {
					ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
					fileName := fmt.Sprintf("banner_%d%s", makeTimestamp(), ext)
					filePath := filepath.Join(mediaDir, fileName)
					os.MkdirAll(mediaDir, 0755)
					out, err := os.Create(filePath)
					if err != nil {
						errMsg = "Ошибка сохранения файла: " + err.Error()
					} else {
						_, err = io.Copy(out, file)
						out.Close()
						file.Close()
						if err != nil {
							errMsg = "Ошибка копирования файла: " + err.Error()
						} else {
							banner := db.Banner{
								Title:    title,
								ImageURL: "/" + filePath,
								CreatedAt: makeNow(),
							}
							if err := db.DB.Create(&banner).Error; err != nil {
								errMsg = "Ошибка сохранения баннера: " + err.Error()
							} else {
								message = "Баннер успешно загружен"
							}
						}
					}
				}
			}
		}
	}

	// GET или после POST — показать список баннеров
	var banners []db.Banner
	db.DB.Order("created_at desc").Find(&banners)
	tmpl := template.Must(template.ParseFiles("web/templates/admin_banners.html"))
	tmpl.Execute(w, struct {
		Banners []db.Banner
		Message string
		Error   string
	}{
		Banners: banners,
		Message: message,
		Error:   errMsg,
	})
}

func deleteBanner(idStr string) error {
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

func toggleBanner(idStr string) error {
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

func makeTimestamp() int64 {
	return makeNow().UnixNano()
}

func makeNow() time.Time {
	return time.Now()
} 