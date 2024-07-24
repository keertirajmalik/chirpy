package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

func main() {
	const port = "8080"

	metrics := apiConfig{
		fileServerHits: 0,
	}

	mux := http.NewServeMux()

	mux.Handle("GET /app/*", metrics.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("GET /admin/metrics", metrics.handleMetrics)

	mux.HandleFunc("GET /api/reset", metrics.handleReset)

	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

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

func handleValidateChirp(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	type validResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		writer.WriteHeader(500)
		return
	}

	if len(params.Body) > 140 {

		response := errorResponse{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			writer.WriteHeader(500)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(400)
		writer.Write(dat)
		return
	}

	cleanMessage := validateMessage(params.Body)

	response := validResponse{
        CleanedBody: cleanMessage,
	}
	dat, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		writer.WriteHeader(500)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	writer.Write(dat)
	return
}

func validateMessage(message string) string {
	parts := strings.Split(message, " ")
	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	for i , part := range parts {
		if slices.Contains(badWords, strings.ToLower(part)) {
			parts[i] = "****"
		}
	}
	return strings.Join(parts, " ")
}
