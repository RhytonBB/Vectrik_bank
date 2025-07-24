package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func InitJWTSecret() {
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	fmt.Println("[auth/jwt.go] jwtSecret (init):", string(jwtSecret))
}

// Теперь userID — числовой тип (uint или int64).
// Здесь лучше использовать int64, т.к. jwt.MapClaims поддерживает числа как float64.
func GenerateJWT(userID uint, isAdmin bool) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"is_admin": isAdmin,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	fmt.Println("[auth/jwt.go] Issued JWT:", tokenStr)
	return tokenStr, err
}

func ParseJWT(tokenStr string) (map[string]interface{}, error) {
	fmt.Println("[auth/jwt.go] Received JWT:", tokenStr)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		fmt.Println("[auth/jwt.go] JWT parse error:", err)
		return nil, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		fmt.Println("[auth/jwt.go] Claims:", claims)
		return claims, nil
	}

	fmt.Println("[auth/jwt.go] Invalid claims")
	return nil, errors.New("invalid claims")
}
