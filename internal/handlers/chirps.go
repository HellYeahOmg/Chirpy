package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/HellYeahOmg/Chirpy/internal/database"
	"github.com/google/uuid"
)

func HandleCreateChirp(dbQueries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body   string `json:"body"`
			UserID string `json:"user_id"`
		}

		type errorReturnValues struct {
			Error string `json:"valid"`
		}

		params := parameters{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(500)
			responseBody := errorReturnValues{
				Error: "Something went wrong",
			}

			dat, err := json.Marshal(responseBody)
			if err != nil {
				w.WriteHeader(500)
				log.Printf("Error marshalling JSON: %s", err)
				return
			}

			w.Write(dat)
			return
		}

		if len(params.Body) > 140 {
			w.WriteHeader(400)
			responseBody := errorReturnValues{
				Error: "Chirp is too long",
			}

			dat, err := json.Marshal(responseBody)
			if err != nil {
				w.WriteHeader(500)
				log.Printf("Error marshalling JSON: %s", err)
				return
			}

			w.Write(dat)
			return
		}

		parsedID, err := uuid.Parse(params.UserID)
		if err != nil {
			log.Printf("failed to parse user_id: %s", err)
			w.WriteHeader(500)
			return
		}

		newChirp := database.CreateChirpParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Body:      params.Body,
			UserID:    parsedID,
		}

		result, err := dbQueries.CreateChirp(r.Context(), newChirp)
		if err != nil {
			log.Printf("failed to create a new chirp: %s", err)
			w.WriteHeader(500)
			return
		}

		responseBody := Chirp{
			ID:        result.ID,
			UpdatedAt: result.UpdatedAt,
			CreatedAt: result.CreatedAt,
			Body:      result.Body,
			UserID:    result.UserID,
		}

		data, err := json.Marshal(responseBody)
		if err != nil {
			log.Printf("failed to marshal chirp: %s", err)
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(201)
		w.Write(data)
	}
}

func HandleGetChirps(dbQueries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := []Chirp{}

		rows, err := dbQueries.GetChirps(r.Context())
		if err != nil {
			log.Printf("failed to get chirps: %s", err)
			w.WriteHeader(500)
			return
		}

		for _, item := range rows {
			jsonItem := Chirp{
				ID:        item.ID,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
				Body:      item.Body,
				UserID:    item.UserID,
			}
			result = append(result, jsonItem)
		}

		data, err := json.Marshal(result)
		if err != nil {
			log.Printf("failed to marshal chirps: %s", err)
			w.WriteHeader(500)
			return
		}

		w.Write(data)
		w.WriteHeader(200)
	}
}

func HandleGetChirp(dbQueries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("chirpId")
		parsedID, err := uuid.Parse(id)
		if err != nil {
			log.Printf("failed to parse chirp id: %s", err)
			w.WriteHeader(500)
			return
		}

		row, err := dbQueries.GetChirp(r.Context(), parsedID)
		if err != nil {
			log.Printf("failed to find a chirp: %s", err)
			w.WriteHeader(404)
			return
		}

		chirp := Chirp{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			UserID:    row.UserID,
			Body:      row.Body,
		}

		data, err := json.Marshal(chirp)
		if err != nil {
			fmt.Printf("failed to marshal chirp: %s", err)
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(200)
		w.Write(data)
	}
}