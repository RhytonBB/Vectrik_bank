package api

import (
	"encoding/json"
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
	"github.com/gorilla/mux"
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

// API: Получить список баннеров в JSON
func AdminBannersAPIHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	isAdmin, ok := claims["is_admin"].(bool)
	if !ok || !isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}
		title := r.FormValue("title")
		files := r.MultipartForm.File["file"]
		if title == "" || len(files) == 0 || files[0] == nil {
			http.Error(w, "Заполните все поля и выберите файл", http.StatusBadRequest)
			return
		}
		fileHeader := files[0]
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Ошибка открытия файла: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		fileName := fmt.Sprintf("banner_%d%s", makeTimestamp(), ext)
		filePath := filepath.Join(mediaDir, fileName)
		os.MkdirAll(mediaDir, 0755)
		out, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Ошибка сохранения файла: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer out.Close()
		if _, err = io.Copy(out, file); err != nil {
			http.Error(w, "Ошибка копирования файла: "+err.Error(), http.StatusInternalServerError)
			return
		}
		banner := db.Banner{
			Title:    title,
			ImageURL: "/" + filePath,
			CreatedAt: makeNow(),
		}
		if err := db.DB.Create(&banner).Error; err != nil {
			http.Error(w, "Ошибка сохранения баннера: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(banner)
		return
	}

	// GET-запрос — как было
	var banners []db.Banner
	if err := db.DB.Order("created_at desc").Find(&banners).Error; err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banners)
}

// API: Активация/деактивация баннера
func AdminBannerActivateAPIHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	claims, err := auth.ParseJWT(cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	isAdmin, ok := claims["is_admin"].(bool)
	if !ok || !isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	bannerIDStr := vars["id"]
	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := db.DB.Model(&db.Banner{}).Where("id = ?", bannerID).Update("is_active", req.IsActive).Error; err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func AdminBannerDetailAPIHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("token")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    claims, err := auth.ParseJWT(cookie.Value)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    isAdmin, ok := claims["is_admin"].(bool)
    if !ok || !isAdmin {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    vars := mux.Vars(r)
    bannerIDStr := vars["id"]
    bannerID, err := strconv.Atoi(bannerIDStr)
    if err != nil {
        http.Error(w, "Некорректный ID", http.StatusBadRequest)
        return
    }
    if r.Method == http.MethodPut {
        var req struct {
            Title    string  `json:"title"`
            ImageURL *string `json:"image_url"`
            IsActive *bool   `json:"is_active"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }
        updates := map[string]interface{}{}
        if req.Title != "" {
            updates["title"] = req.Title
        }
        if req.ImageURL != nil {
            updates["image_url"] = req.ImageURL
        }
        if req.IsActive != nil {
            updates["is_active"] = req.IsActive
        }
        if err := db.DB.Model(&db.Banner{}).Where("id = ?", bannerID).Updates(updates).Error; err != nil {
            http.Error(w, "Ошибка обновления баннера: "+err.Error(), http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
        return
    }
    if r.Method == http.MethodDelete {
        if err := db.DB.Delete(&db.Banner{}, bannerID).Error; err != nil {
            http.Error(w, "Ошибка удаления баннера: "+err.Error(), http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
        return
    }
    http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
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