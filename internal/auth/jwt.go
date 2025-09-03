package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type MyCustomClaims struct {
	jwt.RegisteredClaims
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	j := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": "chirpy",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
			"sub": userID.String(),
		})

	return j.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, err
	}

	subject, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, errors.New("invalid subject claim")
	}

	userID, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if token == "" {
		return "", errors.New("no auth header has been provided")
	}

	parts := strings.Split(token, " ")

	if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("invalid bearer token has been provided")
	}

	return parts[1], nil
}

func MakeRefreshToken() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}
