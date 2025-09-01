package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/HellYeahOmg/Chirpy/internal/auth"
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

	responseBody := User{
		ID:        row.ID,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		Email:     row.Email,
	}

	data, err := json.Marshal(responseBody)
	if err != nil {
		log.Printf("failed to marshal responseBody: %s", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
