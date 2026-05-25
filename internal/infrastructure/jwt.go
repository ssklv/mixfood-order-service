package infrastructure

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// TokenProvider реализует интерфейс, который мы позже определим в usecase
type TokenProvider struct {
	signingKey []byte
}

func NewTokenProvider(key string) *TokenProvider {
	return &TokenProvider{
		signingKey: []byte(key),
	}
}

func (p *TokenProvider) ParseToken(tokenString string) (int64, string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return p.signingKey, nil
	})

	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDVal, ok := claims["sub"].(float64)
		if !ok {
			return 0, "", errors.New("invalid sub claim")
		}

		role, _ := claims["role"].(string)
		return int64(userIDVal), role, nil
	}

	return 0, "", errors.New("invalid token claims")
}
