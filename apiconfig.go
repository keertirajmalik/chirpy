package main

import (
	"fmt"
	"net/http"

	"github.com/keertirajmalik/chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits int
	DB             *database.DB
	jwtSecret      string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(http.StatusOK)

	html := `<html>

<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>`

	writer.Write([]byte(fmt.Sprintf(html, cfg.fileServerHits)))
}

func (cfg *apiConfig) handleReset(writer http.ResponseWriter, request *http.Request) {
	cfg.fileServerHits = 0
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("Hit reset to 0"))
}
