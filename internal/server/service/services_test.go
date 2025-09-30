package service

import (
	"context"
	"testing"

	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/repository/sqlite"
	"gophkeeper/internal/shared/models"
)

func TestAuthRegisterLogin(t *testing.T) {
	repo, err := sqlite.New("file:svc_auth_login?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	svcs := NewServices(repo, config.Config{JWTSecret: "test"})
	ctx := context.Background()
	_, err = svcs.Auth.Register(ctx, "u@example.com", "pass")
	if err != nil {
		t.Fatal(err)
	}
	token, err := svcs.Auth.Login(ctx, "u@example.com", "pass")
	if err != nil || token == "" {
		t.Fatalf("login failed: %v", err)
	}
	uid, err := svcs.Auth.ParseToken(ctx, token)
	if err != nil || uid == "" {
		t.Fatalf("parse failed: %v", err)
	}
}

func TestRefreshFlowAndRecordsService(t *testing.T) {
	repo, err := sqlite.New("file:svc_refresh_records?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	svcs := NewServices(repo, config.Config{JWTSecret: "test"})
	ctx := context.Background()
	u, err := svcs.Auth.Register(ctx, "u3@example.com", "pass")
	if err != nil {
		t.Fatal(err)
	}
	tok, err := svcs.Auth.Login(ctx, "u3@example.com", "pass")
	if err != nil {
		t.Fatal(err)
	}
	uid, err := svcs.Auth.ParseToken(ctx, tok)
	if err != nil || uid == "" {
		t.Fatal(err)
	}
	// issue refresh explicitly and rotate
	r, err := svcs.Auth.IssueRefreshToken(ctx, u.ID, 3600*1e9)
	if err != nil || r == "" {
		t.Fatalf("issue refresh: %v", err)
	}
	at, err := svcs.Auth.Refresh(ctx, r)
	if err != nil || at == "" {
		t.Fatalf("refresh: %v", err)
	}

	// Records service
	rec, err := svcs.Records.Upsert(ctx, models.Record{OwnerID: u.ID, Type: models.RecordTypeText, Meta: map[string]string{"k": "v"}, Payload: []byte("x")})
	if err != nil || rec.ID == "" {
		t.Fatalf("upsert: %v", err)
	}
	list, err := svcs.Records.List(ctx, u.ID)
	if err != nil || len(list) == 0 {
		t.Fatalf("list: %v", err)
	}
	_, err = svcs.Records.Get(ctx, u.ID, rec.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if err := svcs.Records.Delete(ctx, u.ID, rec.ID); err != nil {
		t.Fatalf("del: %v", err)
	}
}
