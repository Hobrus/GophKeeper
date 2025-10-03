package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/repository/sqlite"
	"gophkeeper/internal/server/service"
)

func TestRegister_BadJSON_And_Missing(t *testing.T) {
	ts := newTestServer(t)
	rr := doJSON(t, ts, "POST", "/api/v1/auth/register", "{bad", nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("bad json: %d", rr.Code)
	}
	rr = doJSON(t, ts, "POST", "/api/v1/auth/register", map[string]string{}, nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("missing fields: %d", rr.Code)
	}
}

func TestLogin_Invalid(t *testing.T) {
	ts := newTestServer(t)
	rr := doJSON(t, ts, "POST", "/api/v1/auth/login", map[string]string{"email": "no@user", "password": "x"}, nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", rr.Code)
	}
}

func TestRefresh_Invalid(t *testing.T) {
	ts := newTestServer(t)
	rr := doJSON(t, ts, "POST", "/api/v1/auth/refresh", map[string]string{"refresh_token": "bad"}, nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	ts := newTestServer(t)
	rr := doJSON(t, ts, "GET", "/api/v1/records", nil, map[string]string{"Authorization": "Bearer invalid"})
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", rr.Code)
	}
}

func TestNotFoundHandlers(t *testing.T) {
	// Build a separate server to avoid interference
	repo, err := sqlite.New("file:httpapi_notfound?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	svcs := service.NewServices(repo, config.Config{JWTSecret: "test"})
	ts := NewRouter(svcs, nil)
	// Register and login to get token
	rr := doJSON(t, ts, "POST", "/api/v1/auth/register", map[string]string{"email": "nf@example.com", "password": "p"}, nil)
	if rr.Code != http.StatusCreated {
		t.Fatalf("register: %d", rr.Code)
	}
	rr = doJSON(t, ts, "POST", "/api/v1/auth/login", map[string]string{"email": "nf@example.com", "password": "p"}, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("login: %d", rr.Code)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &tok)

	hdr := map[string]string{"Authorization": "Bearer " + tok.AccessToken}
	rr = doJSON(t, ts, "GET", "/api/v1/records/does-not-exist", nil, hdr)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("get 404: %d", rr.Code)
	}
	rr = doJSON(t, ts, "DELETE", "/api/v1/records/does-not-exist", nil, hdr)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("del 404: %d", rr.Code)
	}
}
