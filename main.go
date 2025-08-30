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

	sm.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		s := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, config.fileserverHits.Load())
		w.Write([]byte(s))
	})

	sm.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		config.resetMetricsInc()
		w.Write([]byte("OK"))
	})

	sm.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	s.ListenAndServe()
}
