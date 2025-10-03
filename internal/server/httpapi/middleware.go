package httpapi

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func (r *Router) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authz := req.Header.Get("Authorization")
		if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(authz, "Bearer ")
		userID, err := r.services.Auth.ParseToken(req.Context(), token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}
		ctx := context.WithValue(req.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func getUserID(ctx context.Context) string {
	if v := ctx.Value(userIDContextKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
