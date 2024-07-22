package main

import (
	"fmt"
	"net/http"
)


type apiConfig struct {
	fileServerHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileServerHits)))
}

func (cfg *apiConfig) handleReset(writer http.ResponseWriter, request *http.Request) {
	cfg.fileServerHits = 0
    writer.WriteHeader(http.StatusOK)
    writer.Write([]byte("Hit reset to 0"))
}
