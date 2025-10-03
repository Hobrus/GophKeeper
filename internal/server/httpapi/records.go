package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"gophkeeper/internal/server/repository"
	"gophkeeper/internal/shared/models"
)

func (r *Router) handleListRecords(w http.ResponseWriter, req *http.Request) {
	userID := getUserID(req.Context())
	records, err := r.services.Records.List(req.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (r *Router) handleGetRecord(w http.ResponseWriter, req *http.Request) {
	userID := getUserID(req.Context())
	id := chi.URLParam(req, "id")
	rec, err := r.services.Records.Get(req.Context(), userID, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (r *Router) handleUpsertRecord(w http.ResponseWriter, req *http.Request) {
	userID := getUserID(req.Context())
	var body models.Record
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	body.OwnerID = userID
	// Optional optimistic concurrency via If-Match: <version>
	ifMatch := req.Header.Get("If-Match")
	if ifMatch != "" {
		var expected int64
		_, _ = fmt.Sscanf(ifMatch, "%d", &expected)
		rec, err := r.services.Records.UpsertConditional(req.Context(), body, expected)
		if err != nil {
			if errors.Is(err, repository.ErrVersionConflict) {
				writeJSON(w, http.StatusPreconditionFailed, map[string]string{"error": "version conflict"})
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("ETag", fmt.Sprintf("%d", rec.Version))
		writeJSON(w, http.StatusOK, rec)
		return
	}
	rec, err := r.services.Records.Upsert(req.Context(), body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("ETag", fmt.Sprintf("%d", rec.Version))
	writeJSON(w, http.StatusOK, rec)
}

func (r *Router) handleDeleteRecord(w http.ResponseWriter, req *http.Request) {
	userID := getUserID(req.Context())
	id := chi.URLParam(req, "id")
	if err := r.services.Records.Delete(req.Context(), userID, id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
