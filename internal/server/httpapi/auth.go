package httpapi

import (
	"encoding/json"
	"net/http"

	"gophkeeper/internal/shared/models"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r *Router) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) handleRegister(w http.ResponseWriter, req *http.Request) {
	var body registerRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if body.Email == "" || body.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password required"})
		return
	}
	user, err := r.services.Auth.Register(req.Context(), body.Email, body.Password)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	token, err := r.services.Auth.Login(req.Context(), body.Email, body.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	// also issue refresh token
	uid, _ := r.services.Auth.ParseToken(req.Context(), token)
	refresh, _ := r.services.Auth.IssueRefreshToken(req.Context(), uid, 30*24*60*60*1e9)
	writeJSON(w, http.StatusOK, models.TokenResponse{AccessToken: token, RefreshToken: refresh})
}

func (r *Router) handleRefresh(w http.ResponseWriter, req *http.Request) {
	var body refreshRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	token, err := r.services.Auth.Refresh(req.Context(), body.RefreshToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, models.TokenResponse{AccessToken: token})
}
