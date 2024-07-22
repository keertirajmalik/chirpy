package main

import (
	"log"
	"net/http"
)

func handlerReadiness(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}


func main() {
	const port = "8080"

	metrics := apiConfig{
		fileServerHits: 0,
	}

	mux := http.NewServeMux()

	mux.Handle("/app/*", metrics.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /healthz", handlerReadiness)

	mux.HandleFunc("GET /metrics", metrics.handleMetrics)

	mux.HandleFunc("/reset", metrics.handleReset)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())

}
