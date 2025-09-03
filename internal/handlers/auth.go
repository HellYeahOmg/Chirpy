package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/HellYeahOmg/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *ApiConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds string `json:"expires_in_seconds"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("failed to decoded request body: %s\n", err)
		w.WriteHeader(500)
		return
	}

	row, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Incorrect email or password"))
		return
	}

	if auth.CheckPasswordHash(params.Password, row.HashedPassword) != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Incorrect email or password"))
		return
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}

	expires, err := time.ParseDuration(params.ExpiresInSeconds)
	if err != nil || expires.Hours() > 1 {
		expires = time.Hour
	}

	token, err := auth.MakeJWT(row.ID, cfg.JwtSecret, expires)
	if err != nil {
		log.Printf("failed to create jwt token: %s", err)
		w.WriteHeader(500)
		return
	}

	responseBody := response{
		ID:        row.ID,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		Email:     row.Email,
		Token:     token,
	}

	data, err := json.Marshal(responseBody)
	if err != nil {
		log.Printf("failed to marshal responseBody: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
