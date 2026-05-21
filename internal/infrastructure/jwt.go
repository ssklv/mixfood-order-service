package infrastructure

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

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
		return p.signingKey, nil
	})
	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDFloat, ok := claims["sub"].(float64)
		if !ok {
			return 0, "", errors.New("invalid user id in token")
		}

		role, _ := claims["role"].(string)
		return int64(userIDFloat), role, nil
	}

	return 0, "", errors.New("invalid token")
}
