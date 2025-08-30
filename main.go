package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) resetMetricsInc() {
	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	sm := http.NewServeMux()
	config := apiConfig{}

	s := http.Server{
		Handler: sm,
		Addr:    ":8080",
	}

	sm.Handle("/app/", config.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./")))))
	sm.Handle("/app/assets", http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets/"))))
	sm.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		s := fmt.Sprintf("Hits: %v", config.fileserverHits.Load())
		w.Write([]byte(s))
	})

	sm.HandleFunc("POST /reset", func(w http.ResponseWriter, r *http.Request) {
		config.resetMetricsInc()
		w.Write([]byte("OK"))
	})

	sm.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	s.ListenAndServe()
}
