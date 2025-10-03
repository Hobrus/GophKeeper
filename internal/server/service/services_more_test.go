package service

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/repository/sqlite"
	"gophkeeper/internal/shared/models"
)

func TestAuth_ErrorsAndHelpers(t *testing.T) {
	repo, _ := sqlite.New("file:svc_errors_helpers?mode=memory&cache=shared")
	svcs := NewServices(repo, config.Config{JWTSecret: "test"})
	ctx := context.Background()

	// Register validations
	if _, err := svcs.Auth.Register(ctx, "", ""); err == nil {
		t.Fatalf("want error on empty inputs")
	}

	// Login invalid user
	if _, err := svcs.Auth.Login(ctx, "no@user", "x"); err == nil {
		t.Fatalf("want invalid credentials")
	}

	// Good user
	_, _ = svcs.Auth.Register(ctx, "u@example.com", "pass")
	if _, err := svcs.Auth.Login(ctx, "u@example.com", "wrong"); err == nil {
		t.Fatalf("want invalid creds on wrong pass")
	}

	token, _ := svcs.Auth.Login(ctx, "u@example.com", "pass")
	// Parse token invalid format
	if _, err := svcs.Auth.ParseToken(ctx, token+"broken"); err == nil {
		t.Fatalf("want parse error")
	}

	// Issue/parse custom access token
	at, err := svcs.Auth.IssueAccessToken("uid", 0)
	if err != nil {
		t.Fatalf("issue access: %v", err)
	}
	_, _ = svcs.Auth.ParseToken(ctx, at)

	// Unexpected signing method
	bad := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": "x"})
	s, _ := bad.SignedString([]byte(""))
	if _, err := svcs.Auth.ParseToken(ctx, s); err == nil {
		t.Fatalf("want error on none alg")
	}
}

func TestRecordsService_Validations(t *testing.T) {
	repo, _ := sqlite.New("file:svc_records_validations?mode=memory&cache=shared")
	svcs := NewServices(repo, config.Config{JWTSecret: "test"})
	ctx := context.Background()

	if _, err := svcs.Records.Upsert(ctx, models.Record{}); err == nil {
		t.Fatalf("missing owner/type should fail")
	}
	if _, err := svcs.Records.UpsertConditional(ctx, models.Record{}, 0); err == nil {
		t.Fatalf("missing owner/type should fail")
	}
}
