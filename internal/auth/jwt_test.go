package auth

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Fatal("Expected non-empty token")
	}
}

func TestValidateJWT_ValidToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	validatedUserID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if validatedUserID != userID {
		t.Fatalf("Expected userID %v, got %v", userID, validatedUserID)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := -time.Hour // Already expired

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("Expected error for expired token, got none")
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	wrongSecret := "wrong-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Fatal("Expected error for wrong secret, got none")
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	secret := "test-secret"
	invalidToken := "invalid.token.string"

	_, err := ValidateJWT(invalidToken, secret)
	if err == nil {
		t.Fatal("Expected error for invalid token, got none")
	}
}

func TestValidateJWT_EmptyToken(t *testing.T) {
	secret := "test-secret"

	_, err := ValidateJWT("", secret)
	if err == nil {
		t.Fatal("Expected error for empty token, got none")
	}
}

func TestGetBearerToken_ValidToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer abc123token")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token != "abc123token" {
		t.Fatalf("Expected token 'abc123token', got '%s'", token)
	}
}

func TestGetBearerToken_NoHeader(t *testing.T) {
	headers := http.Header{}

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatal("Expected error for missing auth header, got none")
	}

	expectedMsg := "no auth header has been provided"
	if err.Error() != expectedMsg {
		t.Fatalf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetBearerToken_InvalidFormat(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "InvalidFormatToken")

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatal("Expected error for invalid token format, got none")
	}

	expectedMsg := "invalid bearer token has been provided"
	if err.Error() != expectedMsg {
		t.Fatalf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetBearerToken_EmptyToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer ")

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatal("Expected error for empty token, got none")
	}

	expectedMsg := "invalid bearer token has been provided"
	if err.Error() != expectedMsg {
		t.Fatalf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetBearerToken_ExtraSpaces(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer  token_with_spaces")

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatal("Expected error for malformed token with extra spaces, got none")
	}

	expectedMsg := "invalid bearer token has been provided"
	if err.Error() != expectedMsg {
		t.Fatalf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestMakeRefreshToken(t *testing.T) {
	token, err := MakeRefreshToken()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(token) != 64 {
		t.Fatalf("Expected token length 64, got %d", len(token))
	}

	for _, char := range token {
		if !strings.Contains("0123456789abcdef", strings.ToLower(string(char))) {
			t.Fatalf("Token contains invalid hex character: %c", char)
		}
	}
}

func TestMakeRefreshToken_Uniqueness(t *testing.T) {
	token1, err := MakeRefreshToken()
	if err != nil {
		t.Fatalf("Expected no error for first token, got %v", err)
	}

	token2, err := MakeRefreshToken()
	if err != nil {
		t.Fatalf("Expected no error for second token, got %v", err)
	}

	if token1 == token2 {
		t.Fatal("Expected tokens to be unique, but they were identical")
	}
}

func TestMakeRefreshToken_MultipleGenerations(t *testing.T) {
	tokens := make(map[string]bool)
	
	for i := 0; i < 100; i++ {
		token, err := MakeRefreshToken()
		if err != nil {
			t.Fatalf("Expected no error on iteration %d, got %v", i, err)
		}
		
		if len(token) != 64 {
			t.Fatalf("Expected token length 64 on iteration %d, got %d", i, len(token))
		}
		
		if tokens[token] {
			t.Fatalf("Duplicate token generated on iteration %d: %s", i, token)
		}
		
		tokens[token] = true
	}
}