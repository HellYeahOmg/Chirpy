package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/HellYeahOmg/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("failed to open db connection: %s", err)
		panic(1)
	}

	dbQueries := database.New(db)

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

	sm.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email string `json:"email"`
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

		dbUser, err := dbQueries.CreateUser(r.Context(), params.Email)
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
	})

	sm.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
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
	})

	s.ListenAndServe()
}
