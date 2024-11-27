package authorization

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/h1067675/shortUrl/internal/logger"
	"go.uber.org/zap"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const SecretKEY = "supersecretkey"

func CheckToken(tokenString string) (int, error) {
	// заадавем утверждения Claims
	logger.Log.Debug("token", zap.String("token", tokenString))

	var cl = Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &cl, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKEY), nil
	})
	if err != nil {
		return -1, err
	}
	if !token.Valid {
		err := fmt.Errorf("token is not valid")
		logger.Log.Debug("token is not valid")
		return -1, err
	}
	logger.Log.Debug("user id restore from token", zap.Int("userid", cl.UserID))
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
	tokenString, err := token.SignedString([]byte(SecretKEY))
	if err != nil {
		return "", err
	}
	logger.Log.Debug("create new token", zap.Int("userid", id))
	return tokenString, nil
}
