package main

import (
	"net/http"
)

func main() {
	sm := http.NewServeMux()
	s := http.Server{
		Handler: sm,
		Addr:    ":8080",
	}

	sm.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("./"))))
	sm.Handle("/app/assets", http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets/"))))

	sm.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	s.ListenAndServe()
}
