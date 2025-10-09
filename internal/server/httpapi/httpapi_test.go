package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/repository/sqlite"
	"gophkeeper/internal/server/service"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	repo, err := sqlite.New("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("repo: %v", err)
	}
	svcs := service.NewServices(repo, config.Config{JWTSecret: "test", MaxRequestBytes: 1 << 20, MaxRecordPayloadBytes: 1 << 20})
	return NewRouter(svcs, nil, 1<<20)
}

func doJSON(t *testing.T, ts http.Handler, method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = &bytes.Buffer{}
	}
	req, _ := http.NewRequest(method, path, buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rr := httptest.NewRecorder()
	ts.ServeHTTP(rr, req)
	return rr
}

func TestHealth(t *testing.T) {
	ts := newTestServer(t)
	rr := doJSON(t, ts, "GET", "/health", nil, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("health status: %d", rr.Code)
	}
}

func TestAuthAndCRUD(t *testing.T) {
	ts := newTestServer(t)

	// Register
	rr := doJSON(t, ts, "POST", "/api/v1/auth/register", map[string]string{"email": "u@example.com", "password": "pass"}, nil)
	if rr.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", rr.Code, rr.Body.String())
	}

	// Login -> tokens
	rr = doJSON(t, ts, "POST", "/api/v1/auth/login", map[string]string{"email": "u@example.com", "password": "pass"}, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("login: %d %s", rr.Code, rr.Body.String())
	}
	var tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &tokens)
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("tokens empty")
	}

	// Refresh
	rr = doJSON(t, ts, "POST", "/api/v1/auth/refresh", map[string]string{"refresh_token": tokens.RefreshToken}, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("refresh: %d %s", rr.Code, rr.Body.String())
	}

	authz := map[string]string{"Authorization": "Bearer " + tokens.AccessToken}

	// Create record
	body := map[string]any{"type": "text", "meta": map[string]string{"title": "t"}, "payload": []byte("x")}
	rr = doJSON(t, ts, "POST", "/api/v1/records", body, authz)
	if rr.Code != http.StatusOK {
		t.Fatalf("create rec: %d %s", rr.Code, rr.Body.String())
	}
	etag := rr.Header().Get("ETag")
	if etag == "" {
		t.Fatalf("missing ETag")
	}
	var rec struct {
		ID      string `json:"id"`
		Version int64  `json:"version"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &rec)
	if rec.ID == "" || rec.Version == 0 {
		t.Fatalf("bad rec: %+v", rec)
	}

	// List
	rr = doJSON(t, ts, "GET", "/api/v1/records", nil, authz)
	if rr.Code != http.StatusOK {
		t.Fatalf("list: %d", rr.Code)
	}

	// Get one
	rr = doJSON(t, ts, "GET", "/api/v1/records/"+rec.ID, nil, authz)
	if rr.Code != http.StatusOK {
		t.Fatalf("get: %d", rr.Code)
	}

	// Conditional update: wrong If-Match -> 412
	badBody := map[string]any{"id": rec.ID, "type": "text", "meta": map[string]string{"title": "t2"}, "payload": []byte("y")}
	hdr := map[string]string{"Authorization": authz["Authorization"], "If-Match": "999"}
	rr = doJSON(t, ts, "POST", "/api/v1/records", badBody, hdr)
	if rr.Code != http.StatusPreconditionFailed {
		t.Fatalf("want 412 got %d", rr.Code)
	}

	// Conditional update: correct If-Match
	hdr["If-Match"] = etag
	goodBody := map[string]any{"id": rec.ID, "type": "text", "meta": map[string]string{"title": "t3"}, "payload": []byte("z")}
	rr = doJSON(t, ts, "POST", "/api/v1/records", goodBody, hdr)
	if rr.Code != http.StatusOK {
		t.Fatalf("upsert cond: %d", rr.Code)
	}

	// Delete
	rr = doJSON(t, ts, "DELETE", "/api/v1/records/"+rec.ID, nil, authz)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", rr.Code)
	}
}

func TestAuthMiddleware_Unauthorized(t *testing.T) {
	ts := newTestServer(t)
	rr := doJSON(t, ts, "GET", "/api/v1/records", nil, nil)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", rr.Code)
	}
}
