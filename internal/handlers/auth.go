package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/HellYeahOmg/Chirpy/internal/auth"
	"github.com/HellYeahOmg/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *ApiConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	accessToken, err := auth.MakeJWT(row.ID, cfg.JwtSecret)
	if err != nil {
		log.Printf("failed to create jwt token: %s", err)
		w.WriteHeader(500)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("failed to create refresh token: %s", err)
		w.WriteHeader(500)
		return
	}

	rtRow := database.AddRefreshTokenParams{
		Token:      refreshToken,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		UserID:     row.ID,
		ExperiesAt: time.Now().AddDate(0, 0, 60),
	}

	err = cfg.DB.AddRefreshToken(r.Context(), rtRow)
	if err != nil {
		log.Printf("failed to save refresh token to db: %s", err)
		w.WriteHeader(500)
		return
	}

	responseBody := response{
		ID:           row.ID,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		Email:        row.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}

	data, err := json.Marshal(responseBody)
	if err != nil {
		log.Printf("failed to marshal responseBody: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (cfg *ApiConfig) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	row, err := cfg.DB.GetRefreshToken(r.Context(), token)
	if err != nil || row.RevokedAt.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	accessToken, err := auth.MakeJWT(row.UserID, cfg.JwtSecret)
	if err != nil {
		log.Printf("failed to create jwt token: %s", err)
		w.WriteHeader(500)
		return

	}

	responseBody := response{
		Token: accessToken,
	}

	data, err := json.Marshal(responseBody)
	if err != nil {
		log.Printf("failed to marshal new jwt token using refresh token: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *ApiConfig) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	input := database.UpdateRefreshTokenParams{
		RevokedAt: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		Token: refreshToken,
	}
	err = cfg.DB.UpdateRefreshToken(r.Context(), input)
	if err != nil {
		log.Printf("failed to revoke rt: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
