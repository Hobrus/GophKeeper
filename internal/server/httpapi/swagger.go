package httpapi

import (
	"net/http"
	"os"
)

func (r *Router) handleSwagger(w http.ResponseWriter, req *http.Request) {
	b, err := os.ReadFile("internal/server/httpapi/swagger/openapi.yaml")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}
