package authorization

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const Secret_KEY = "supersecretkey"

func CheckToken(tokenString string) (int, error) {
	// заадавем утверждения Claims
	var cl = Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &cl, func(t *jwt.Token) (interface{}, error) {
		return []byte(Secret_KEY), nil
	})
	if err != nil {
		return -1, err
	}
	if !token.Valid {
		err := fmt.Errorf("token is not valid")
		return -1, err
	}
	return cl.UserID, nil
}

func SetToken(id int) (string, error) {
	// задаем утверждения Claims
	var cl = Claims{
		UserID: id,
	}
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cl)
	// создаём строку токена
	tokenString, err := token.SignedString([]byte(Secret_KEY))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
