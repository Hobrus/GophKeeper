package service

import (
	"context"
	"testing"
	"time"

	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/repository/sqlite"
	"gophkeeper/internal/shared/models"
)

func TestRefresh_ExpiredToken(t *testing.T) {
	repo, err := sqlite.New("file:svc_refresh_expired?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	svcs := NewServices(repo, config.Config{JWTSecret: "test"})
	ctx := context.Background()

	user, err := svcs.Auth.Register(ctx, "exp@example.com", "pass")
	if err != nil {
		t.Fatal(err)
	}
	// Issue already expired refresh token
	tok, err := svcs.Auth.IssueRefreshToken(ctx, user.ID, -1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svcs.Auth.Refresh(ctx, tok); err == nil {
		t.Fatalf("expected error on expired refresh token")
	}
}

func TestRecordsService_MetaNilPaths(t *testing.T) {
	repo, err := sqlite.New("file:svc_records_meta_nil?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	svcs := NewServices(repo, config.Config{JWTSecret: "test"})
	ctx := context.Background()

	u, err := svcs.Auth.Register(ctx, "meta@example.com", "pass")
	if err != nil {
		t.Fatal(err)
	}

	// Upsert with nil meta should be normalized to empty map
	rec, err := svcs.Records.Upsert(ctx, models.Record{OwnerID: u.ID, Type: models.RecordTypeText, Meta: nil, Payload: []byte("x")})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Meta == nil {
		t.Fatalf("meta should be normalized, got nil")
	}

	// Conditional upsert with nil meta path
	rec2, err := svcs.Records.UpsertConditional(ctx, models.Record{ID: rec.ID, OwnerID: u.ID, Type: models.RecordTypeText, Meta: nil, Payload: []byte("y")}, rec.Version)
	if err != nil {
		t.Fatal(err)
	}
	if rec2.Version != rec.Version+1 {
		t.Fatalf("version not incremented: %d -> %d", rec.Version, rec2.Version)
	}
	if rec2.Meta == nil {
		t.Fatalf("meta should be normalized, got nil")
	}
}
