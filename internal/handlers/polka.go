package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/HellYeahOmg/Chirpy/internal/auth"
	"github.com/HellYeahOmg/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *ApiConfig) HandlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if apiKey != cfg.PolkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	parsedUserId, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	queryArg := database.SetIsRedChirpyUserParams{
		IsChirpyRed: sql.NullBool{Valid: true, Bool: true},
		ID:          parsedUserId,
	}
	err = cfg.DB.SetIsRedChirpyUser(r.Context(), queryArg)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
