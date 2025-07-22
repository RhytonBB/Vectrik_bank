package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
	"github.com/gorilla/mux"
)

func AdminProductsHandler(w http.ResponseWriter, r *http.Request) {
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
	if r.Method == http.MethodGet {
		products, err := services.GetAllProducts()
		if err != nil {
			http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
			return
		}
		data := struct {
			Products []db.Product
			Message  string
			Error    string
		}{
			Products: products,
		}
		tmpl := template.Must(template.ParseFiles("web/templates/admin_products.html"))
		tmpl.Execute(w, data)
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			showProductsWithError(w, "Ошибка обработки формы")
			return
		}
		name := r.FormValue("name")
		description := r.FormValue("description")
		priceStr := r.FormValue("price")
		typeStr := r.FormValue("type")
		stockStr := r.FormValue("stock")
		mediaURL := r.FormValue("media_url")
		isActive := r.FormValue("is_active") == "on"
		price, err1 := strconv.Atoi(priceStr)
		stock, err2 := strconv.Atoi(stockStr)
		if name == "" || priceStr == "" || typeStr == "" || stockStr == "" || err1 != nil || err2 != nil {
			showProductsWithError(w, "Заполните все обязательные поля корректно")
			return
		}
		var mediaPtr *string
		if mediaURL != "" {
			mediaPtr = &mediaURL
		}
		product := db.Product{
			Name:        name,
			Description: description,
			Price:       price,
			Type:        typeStr,
			Stock:       stock,
			MediaURL:    mediaPtr,
			IsActive:    isActive,
		}
		err := services.CreateProduct(&product)
		if err != nil {
			showProductsWithError(w, "Ошибка сохранения товара: "+err.Error())
			return
		}
		products, _ := services.GetAllProducts()
		data := struct {
			Products []db.Product
			Message  string
			Error    string
		}{
			Products: products,
			Message:  "Товар успешно добавлен",
		}
		tmpl := template.Must(template.ParseFiles("web/templates/admin_products.html"))
		tmpl.Execute(w, data)
		return
	}
	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

func AdminProductsAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
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
		var req struct {
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Price       int     `json:"price"`
			Stock       int     `json:"stock"`
			Type        string  `json:"type"`
			MediaURL    *string `json:"media_url"`
			IsActive    bool    `json:"is_active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		product := db.Product{
			Name:        req.Name,
			Description: req.Description,
			Price:       req.Price,
			Type:        req.Type,
			Stock:       req.Stock,
			MediaURL:    req.MediaURL,
			IsActive:    req.IsActive,
		}
		err = services.CreateProduct(&product)
		if err != nil {
			http.Error(w, "Ошибка сохранения товара: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
		return
	}
	if r.Method == http.MethodGet {
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
		products, err := services.GetAllProducts()
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
		return
	}
	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

func AdminProductDetailAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	productIDStr := vars["id"]
	productID64, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}
	productID := uint(productID64)
	if r.Method == http.MethodGet {
		product, err := services.GetProductByID(productID)
		if err != nil {
			http.Error(w, "Товар не найден", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
		return
	}
	if r.Method == http.MethodPut {
		var req struct {
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Price       int     `json:"price"`
			Stock       int     `json:"stock"`
			Type        string  `json:"type"`
			MediaURL    *string `json:"media_url"`
			IsActive    bool    `json:"is_active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		updates := map[string]interface{}{
			"name":        req.Name,
			"description": req.Description,
			"price":       req.Price,
			"stock":       req.Stock,
			"type":        req.Type,
			"media_url":   req.MediaURL,
			"is_active":   req.IsActive,
		}
		err := services.UpdateProduct(productID, updates)
		if err != nil {
			http.Error(w, "Ошибка обновления товара: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method == http.MethodDelete {
		err := services.DeleteProduct(productID)
		if err != nil {
			http.Error(w, "Ошибка удаления товара: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

func AdminProductUploadAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["file"]
	if len(files) == 0 || files[0] == nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	url, err := services.UploadProductMedia(files[0])
	if err != nil {
		http.Error(w, "Ошибка загрузки файла: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"url": "%s"}`, url)))
}

func showProductsWithError(w http.ResponseWriter, errMsg string) {
	products, _ := services.GetAllProducts()
	data := struct {
		Products []db.Product
		Message  string
		Error    string
	}{
		Products: products,
		Error:    errMsg,
	}
	tmpl := template.Must(template.ParseFiles("web/templates/admin_products.html"))
	tmpl.Execute(w, data)
} 