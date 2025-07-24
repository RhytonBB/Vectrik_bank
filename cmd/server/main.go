package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/RhytonBB/vectrik-webbank/internal/api"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
	"github.com/RhytonBB/vectrik-webbank/internal/auth"
	"github.com/RhytonBB/vectrik-webbank/internal/config"
	"github.com/RhytonBB/vectrik-webbank/internal/services"
	"github.com/rs/cors"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	cfg := config.LoadConfig()

	if err := db.InitDB(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	auth.InitJWTSecret()
	services.InitBotToken(cfg.BotToken)

	api.RegisterRoutes()

	// Обработчик статики для медиафайлов (баннеры, фото товаров и т.д.)
	http.Handle("/internal/media/", http.StripPrefix("/internal/media/", http.FileServer(http.Dir("internal/media"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Логируем JWT_SECRET в файл для отладки
	f, err := os.OpenFile("secret.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString("JWT_SECRET: " + os.Getenv("JWT_SECRET") + "\n")
		f.Close()
	}

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://vectrik-bank-front.onrender.com", "https://vectrikbank-front.onrender.com"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
	}).Handler(http.DefaultServeMux)

	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
