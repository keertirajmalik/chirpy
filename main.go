package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/keertirajmalik/chirpy/internal/database"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load(".env")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET enviornment variable is not set")
	}

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if dbg != nil && *dbg {
		err := db.ResetDB()
		if err != nil {
			log.Fatal(err)
		}
	}

	config := apiConfig{
		fileServerHits: 0,
		DB:             db,
		jwtSecret:      jwtSecret,
	}

	mux := http.NewServeMux()

	mux.Handle("GET /app/*", config.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", config.handleMetrics)
	mux.HandleFunc("GET /api/reset", config.handleReset)

	mux.HandleFunc("GET /api/chirps", config.handleChirpGet)
	mux.HandleFunc("POST /api/chirps", config.handleChirpCreate)
	mux.HandleFunc("GET /api/chirps/{chirpID}", config.handleChirpGetSpecific)

	mux.HandleFunc("POST /api/users", config.handleUsersCreate)
	mux.HandleFunc("PUT /api/users", config.handleUsersUpdate)

	mux.HandleFunc("POST /api/login", config.handleLogin)
	mux.HandleFunc("POST /api/refresh", config.handleRefresh)
    mux.HandleFunc("POST /api/revoke", config.handleRevoke)

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
