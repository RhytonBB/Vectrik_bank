package api

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
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
		var products []db.Product
		if err := db.DB.Order("created_at desc").Find(&products).Error; err != nil {
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
			log.Printf("[AdminProductsHandler] Ошибка обработки формы: %v", err)
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

		log.Printf("[AdminProductsHandler] name='%s', description='%s', price='%s', type='%s', stock='%s', media_url='%s', is_active=%v", name, description, priceStr, typeStr, stockStr, mediaURL, isActive)

		if name == "" || priceStr == "" || typeStr == "" || stockStr == "" {
			log.Printf("[AdminProductsHandler] Не все обязательные поля заполнены: name='%s', price='%s', type='%s', stock='%s'", name, priceStr, typeStr, stockStr)
			showProductsWithError(w, "Заполните все обязательные поля")
			return
		}
		price, err1 := strconv.Atoi(priceStr)
		if err1 != nil {
			log.Printf("[AdminProductsHandler] Некорректная цена: '%s'", priceStr)
			showProductsWithError(w, "Некорректная цена")
			return
		}
		stock, err2 := strconv.Atoi(stockStr)
		if err2 != nil {
			log.Printf("[AdminProductsHandler] Некорректный сток: '%s'", stockStr)
			showProductsWithError(w, "Некорректный сток")
			return
		}
		if typeStr != "regular" && typeStr != "drop" && typeStr != "shared" {
			log.Printf("[AdminProductsHandler] Некорректный тип товара: '%s'", typeStr)
			showProductsWithError(w, "Некорректный тип товара")
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
			CreatedAt:   time.Now(),
		}
		if err := db.DB.Create(&product).Error; err != nil {
			log.Printf("[AdminProductsHandler] Ошибка сохранения товара: %v", err)
			showProductsWithError(w, "Ошибка сохранения товара: "+err.Error())
			return
		}
		// Успех
		log.Printf("[AdminProductsHandler] Товар успешно добавлен: %+v", product)
		var products []db.Product
		db.DB.Order("created_at desc").Find(&products)
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

// Добавляю вспомогательную функцию для вывода ошибки
func showProductsWithError(w http.ResponseWriter, errMsg string) {
	var products []db.Product
	db.DB.Order("created_at desc").Find(&products)
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