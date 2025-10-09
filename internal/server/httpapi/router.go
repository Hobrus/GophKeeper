package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"gophkeeper/internal/server/service"
)

type Router struct {
	services        *service.Services
	logger          *log.Logger
	maxRequestBytes int64
}

func NewRouter(services *service.Services, logger *log.Logger, maxRequestBytes int64) http.Handler {
	r := &Router{services: services, logger: logger, maxRequestBytes: maxRequestBytes}
	mux := chi.NewRouter()

	mux.Get("/health", r.handleHealth)
	mux.Get("/swagger.yaml", r.handleSwagger)
	mux.Post("/api/v1/auth/register", r.handleRegister)
	mux.Post("/api/v1/auth/login", r.handleLogin)
	mux.Post("/api/v1/auth/refresh", r.handleRefresh)

	mux.Group(func(pr chi.Router) {
		pr.Use(r.authMiddleware)
		pr.Get("/api/v1/records", r.handleListRecords)
		pr.Post("/api/v1/records", r.handleUpsertRecord)
		pr.Get("/api/v1/records/{id}", r.handleGetRecord)
		pr.Delete("/api/v1/records/{id}", r.handleDeleteRecord)
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
