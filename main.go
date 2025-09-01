package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/HellYeahOmg/Chirpy/internal/database"
	"github.com/HellYeahOmg/Chirpy/internal/handlers"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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
	config := handlers.ApiConfig{}

	s := http.Server{
		Handler: sm,
		Addr:    ":8080",
	}

	sm.Handle("/app/", config.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir("./")))))
	sm.Handle("/app/assets", http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets/"))))

	sm.HandleFunc("GET /admin/metrics", config.HandleMetrics)
	sm.HandleFunc("POST /admin/reset", config.HandleReset)

	sm.HandleFunc("GET /api/healthz", handlers.HandleHealthz)

	sm.HandleFunc("POST /api/users", handlers.HandleCreateUser(dbQueries))

	sm.HandleFunc("POST /api/chirps", handlers.HandleCreateChirp(dbQueries))

	sm.HandleFunc("GET /api/chirps", handlers.HandleGetChirps(dbQueries))

	sm.HandleFunc("GET /api/chirps/{chirpId}", handlers.HandleGetChirp(dbQueries))

	sm.HandleFunc("POST /api/login", handlers.HandleLogin(dbQueries))

	s.ListenAndServe()
}
