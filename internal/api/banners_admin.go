package api

import (
	"encoding/json"
<<<<<<< HEAD
=======
	"fmt"
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
<<<<<<< HEAD
	"github.com/RhytonBB/vectrik-webbank/internal/services"
=======
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
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
		r.ParseMultipartForm(10 << 20)
		action := r.FormValue("action")
		bannerID := r.FormValue("banner_id")
		if action == "delete" && bannerID != "" {
			err := services.DeleteBannerByID(bannerID)
			if err != nil {
				errMsg = "Ошибка удаления баннера: " + err.Error()
			} else {
				message = "Баннер удалён"
			}
		} else if action == "toggle" && bannerID != "" {
			err := services.ToggleBannerByID(bannerID)
			if err != nil {
				errMsg = "Ошибка изменения статуса: " + err.Error()
			} else {
				message = "Статус баннера изменён"
			}
		} else if r.MultipartForm != nil && r.MultipartForm.File != nil && len(r.MultipartForm.File["file"]) > 0 {
			title := r.FormValue("title")
			fileHeader := r.MultipartForm.File["file"][0]
			if title == "" || fileHeader == nil {
				errMsg = "Заполните все поля и выберите файл"
			} else {
				_, err = services.UploadAndCreateBanner(title, fileHeader)
				if err != nil {
					errMsg = "Ошибка загрузки баннера: " + err.Error()
				} else {
					message = "Баннер успешно загружен"
				}
			}
		}
	}
	banners, _ := services.GetAllBanners()
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

<<<<<<< HEAD
=======
// API: Получить список баннеров в JSON
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
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
<<<<<<< HEAD
	if r.Method == http.MethodPost {
=======

	if r.Method == http.MethodPost {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
		title := r.FormValue("title")
		files := r.MultipartForm.File["file"]
		if title == "" || len(files) == 0 || files[0] == nil {
			http.Error(w, "Заполните все поля и выберите файл", http.StatusBadRequest)
			return
		}
<<<<<<< HEAD
		banner, err := services.UploadAndCreateBanner(title, files[0])
		if err != nil {
			http.Error(w, "Ошибка загрузки баннера: "+err.Error(), http.StatusInternalServerError)
			return
		}
=======
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
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(banner)
		return
	}
<<<<<<< HEAD
	banners, err := services.GetAllBanners()
	if err != nil {
=======

	// GET-запрос — как было
	var banners []db.Banner
	if err := db.DB.Order("created_at desc").Find(&banners).Error; err != nil {
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banners)
}

<<<<<<< HEAD
=======
// API: Активация/деактивация баннера
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
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
<<<<<<< HEAD
=======
	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
<<<<<<< HEAD
	err = services.SetBannerActiveStatus(bannerIDStr, req.IsActive)
	if err != nil {
=======
	if err := db.DB.Model(&db.Banner{}).Where("id = ?", bannerID).Update("is_active", req.IsActive).Error; err != nil {
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func AdminBannerDetailAPIHandler(w http.ResponseWriter, r *http.Request) {
<<<<<<< HEAD
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
		err := services.UpdateBannerByID(bannerIDStr, req.Title, req.ImageURL, req.IsActive)
		if err != nil {
			http.Error(w, "Ошибка обновления баннера: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method == http.MethodDelete {
		err := services.DeleteBannerByID(bannerIDStr)
		if err != nil {
			http.Error(w, "Ошибка удаления баннера: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
=======
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
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
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