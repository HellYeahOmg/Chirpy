package handlers

import (
	"net/http"
	"sync/atomic"

	"github.com/HellYeahOmg/Chirpy/internal/database"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	JwtSecret      string
	PolkaKey       string
}

func (cfg *ApiConfig) ResetMetricsInc() {
	cfg.FileserverHits.Store(0)
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
