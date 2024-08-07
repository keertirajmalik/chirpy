package main

import (
	"log"
	"net/http"

	"github.com/keertirajmalik/chirpy/internal/database"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	metrics := apiConfig{
		fileServerHits: 0,
		DB:             db,
	}

	mux := http.NewServeMux()

	mux.Handle("GET /app/*", metrics.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("GET /admin/metrics", metrics.handleMetrics)

	mux.HandleFunc("GET /api/reset", metrics.handleReset)

	mux.HandleFunc("GET /api/chirps", metrics.handleChirpGet)

	mux.HandleFunc("POST /api/chirps", metrics.handleChirpCreate)

	mux.HandleFunc("GET /api/chirps/{chirpID}", metrics.handleChirpGetSpecific)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func handlerReadiness(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}
