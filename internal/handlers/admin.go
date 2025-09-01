package handlers

import (
	"fmt"
	"net/http"
)

func (cfg *ApiConfig) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	s := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.FileserverHits.Load())
	w.Write([]byte(s))
}

func (cfg *ApiConfig) HandleReset(w http.ResponseWriter, r *http.Request) {
	cfg.ResetMetricsInc()
	w.Write([]byte("OK"))
}