package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/HellYeahOmg/Chirpy/internal/auth"
	"github.com/HellYeahOmg/Chirpy/internal/database"
)

func HandleCreateUser(dbQueries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		dbUser, err := dbQueries.CreateUser(r.Context(), queryParams)
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
}
