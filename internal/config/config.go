package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    DatabaseURL string
    Port        string
    JWTSecret   string
}

func LoadConfig() *Config {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    return &Config{
        DatabaseURL: os.Getenv("DATABASE_URL"),
        Port:        os.Getenv("PORT"),
        JWTSecret:   os.Getenv("JWT_SECRET"),
    }
}
