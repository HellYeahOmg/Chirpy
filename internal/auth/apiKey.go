package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if token == "" {
		return "", errors.New("no auth header has been provided")
	}

	parts := strings.Split(token, " ")

	if len(parts) != 2 || parts[0] != "ApiKey" || strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("invalid api token has been provided")
	}

	return parts[1], nil
}
