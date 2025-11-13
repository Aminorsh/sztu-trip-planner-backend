package utils

import (
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/golang-jwt/jwt/v5"
)

var jwtPrivateKey = []byte(config.GetJWTKey())

type Claims struct {
	UserID uint64
	jwt.RegisteredClaims
}

func GenerateToken(userID uint64) (string, error) {
	expirationTime := jwt.NewNumericDate(time.Now().Add(time.Duration(config.GetJWTExpirationDuration()) * time.Hour))

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expirationTime,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtPrivateKey)
}

func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		return jwtPrivateKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
