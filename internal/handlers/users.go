package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/HellYeahOmg/Chirpy/internal/auth"
	"github.com/HellYeahOmg/Chirpy/internal/database"
)

func (cfg *ApiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("failed to decoded request body: %s\n", err)
		w.WriteHeader(500)
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("failed to hash the password: %v", err)
		w.WriteHeader(500)
		return
	}

	queryParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
	}

	dbUser, err := cfg.DB.CreateUser(r.Context(), queryParams)
	if err != nil {
		log.Printf("failed to create a new user: %s", err)
		w.WriteHeader(500)
		return
	}

	responseBody := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	data, err := json.Marshal(responseBody)
	if err != nil {
		log.Printf("failed to marshal responseBody: %s", err)
	}

	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *ApiConfig) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	id, err := auth.ValidateJWT(accessToken, cfg.JwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("failed to parse params in HandlerUpdateUser: %s", err)
		w.WriteHeader(500)
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("failed to hash the password: %v", err)
		w.WriteHeader(500)
		return
	}

	queryParams := database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
		ID:             id,
	}

	dbUser, err := cfg.DB.UpdateUser(r.Context(), queryParams)
	if err != nil {
		log.Printf("failed to create a new user: %s", err)
		w.WriteHeader(500)
		return
	}

	responseBody := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	data, err := json.Marshal(responseBody)
	if err != nil {
		log.Printf("failed to marshal responseBody: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
