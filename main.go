package main

import (
	"encoding/json"
	"fmt"
	"log"
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

	sm.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")

		type parameters struct {
			Body string `json:"body"`
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

		fmt.Println(len(params.Body))

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

		type returnValues struct {
			Valid bool `json:"valid"`
		}

		responseBody := returnValues{
			Valid: true,
		}

		dat, err := json.Marshal(responseBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(200)
		w.Write(dat)
	})

	s.ListenAndServe()
}
