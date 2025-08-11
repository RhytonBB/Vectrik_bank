package api

import (
	"encoding/json"
	"net/http"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/gorilla/mux"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
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
	if r.Method == http.MethodPost {
		r.ParseMultipartForm(100 << 20)
		action := r.FormValue("action")
		bannerID := r.FormValue("banner_id")
		if action == "delete" && bannerID != "" {
			err := services.DeleteBannerByID(bannerID)
			if err != nil {
				// Ошибка удаления баннера
			} else {
				// Баннер удалён
			}
		} else if action == "toggle" && bannerID != "" {
			err := services.ToggleBannerByID(bannerID)
			if err != nil {
				// Ошибка изменения статуса
			} else {
				// Статус баннера изменён
			}
		} else if r.MultipartForm != nil && r.MultipartForm.File != nil && len(r.MultipartForm.File["file"]) > 0 {
			title := r.FormValue("title")
			fileHeader := r.MultipartForm.File["file"][0]
			if title == "" || fileHeader == nil {
				// Заполните все поля и выберите файл
			} else {
				_, err = services.UploadAndCreateBanner(title, fileHeader)
				if err != nil {
					// Ошибка загрузки баннера
				} else {
					// Баннер успешно загружен
				}
			}
		}
	}
}

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
		title := r.FormValue("title")
		files := r.MultipartForm.File["file"]
		if title == "" || len(files) == 0 || files[0] == nil {
			http.Error(w, "Заполните все поля и выберите файл", http.StatusBadRequest)
			return
		}
		banner, err := services.UploadAndCreateBanner(title, files[0])
		if err != nil {
			http.Error(w, "Ошибка загрузки баннера: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(banner)
		return
	}
	banners, err := services.GetAllBanners()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banners)
}

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
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	err = services.SetBannerActiveStatus(bannerIDStr, req.IsActive)
	if err != nil {
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
}

// API: Получить активные баннеры для клиентской части
func ClientBannersAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	
	// Проверяем, что пользователь не админ (только для клиентской части)
	isAdmin, ok := claims["is_admin"].(bool)
	if ok && isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	banners, err := services.GetActiveBanners()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banners)
} 