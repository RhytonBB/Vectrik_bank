package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"database/sql"
	_ "github.com/lib/pq"
	"os"
)

func main() {
	// Получаем строку подключения к БД из переменной окружения
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Подключаемся к базе данных
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
	}

	fmt.Println("Connected to database successfully")

	// Получаем всех пользователей без номера счета
	rows, err := db.Query("SELECT id FROM users WHERE account_number = '00000000' OR account_number IS NULL")
	if err != nil {
		log.Fatal("Error querying users:", err)
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		err := rows.Scan(&userID)
		if err != nil {
			log.Printf("Error scanning user ID: %v", err)
			continue
		}
		userIDs = append(userIDs, userID)
	}

	fmt.Printf("Found %d users without account numbers\n", len(userIDs))

	// Обновляем каждого пользователя
	rand.Seed(time.Now().UnixNano())
	updatedCount := 0

	for _, userID := range userIDs {
		// Генерируем уникальный номер счета
		accountNumber := generateUniqueAccountNumber(db)
		
		// Обновляем пользователя
		_, err := db.Exec("UPDATE users SET account_number = $1 WHERE id = $2", accountNumber, userID)
		if err != nil {
			log.Printf("Error updating user %d: %v", userID, err)
			continue
		}
		
		fmt.Printf("Updated user %d with account number %s\n", userID, accountNumber)
		updatedCount++
	}

	fmt.Printf("Successfully updated %d users\n", updatedCount)
}

// Генерирует уникальный 8-значный номер счета
func generateUniqueAccountNumber(db *sql.DB) string {
	for {
		accountNumber := fmt.Sprintf("%08d", rand.Intn(100000000))
		
		// Проверяем, что номер уникален
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE account_number = $1)", accountNumber).Scan(&exists)
		if err != nil {
			log.Printf("Error checking account number uniqueness: %v", err)
			continue
		}
		
		if !exists {
			return accountNumber
		}
	}
} 