package sqlite

import (
	"context"
	"testing"
	"time"

	"gophkeeper/internal/shared/models"
)

func TestUsersAndRecords(t *testing.T) {
	repo, err := New("file:repo_users_records?mode=memory&cache=shared&_journal=WAL")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	ctx := context.Background()
	user, err := repo.CreateUser(ctx, "u@example.com", []byte("h"))
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == "" {
		t.Fatalf("user id empty")
	}
	_, _, err = repo.GetUserByEmail(ctx, "u@example.com")
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}
	rec, err := repo.UpsertRecord(ctx, models.Record{OwnerID: user.ID, Type: models.RecordTypeText, Meta: map[string]string{"a": "b"}, Payload: []byte("x")})
	if err != nil {
		t.Fatal(err)
	}
	if rec.ID == "" || rec.Version == 0 {
		t.Fatalf("bad rec: %+v", rec)
	}
	list, err := repo.ListRecords(ctx, user.ID)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %d", err, len(list))
	}
	got, err := repo.GetRecord(ctx, user.ID, rec.ID)
	if err != nil || got.ID != rec.ID {
		t.Fatalf("get one: %v", err)
	}
	if err := repo.DeleteRecord(ctx, user.ID, rec.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestRefreshTokensAndConditional(t *testing.T) {
	repo, err := New("file:repo_refresh_conditional?mode=memory&cache=shared&_journal=WAL")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	ctx := context.Background()
	user, err := repo.CreateUser(ctx, "u2@example.com", []byte("h"))
	if err != nil {
		t.Fatal(err)
	}
	// refresh tokens
	if err := repo.CreateRefreshToken(ctx, user.ID, "tok", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	uid, exp, err := repo.GetRefreshToken(ctx, "tok")
	if err != nil || uid != user.ID || exp.IsZero() {
		t.Fatalf("get refresh: %v %s", err, uid)
	}
	if err := repo.DeleteRefreshToken(ctx, "tok"); err != nil {
		t.Fatalf("del refresh: %v", err)
	}

	// conditional upsert
	rec, err := repo.UpsertRecord(ctx, models.Record{OwnerID: user.ID, Type: models.RecordTypeText, Meta: map[string]string{}, Payload: []byte("x")})
	if err != nil {
		t.Fatal(err)
	}
	// wrong expected -> conflict
	_, err = repo.UpsertRecordConditional(ctx, models.Record{ID: rec.ID, OwnerID: user.ID, Type: models.RecordTypeText, Meta: map[string]string{}, Payload: []byte("y")}, rec.Version+1)
	if err == nil {
		t.Fatalf("expected conflict")
	}
	// correct expected -> ok
	rec2, err := repo.UpsertRecordConditional(ctx, models.Record{ID: rec.ID, OwnerID: user.ID, Type: models.RecordTypeText, Meta: map[string]string{}, Payload: []byte("y")}, rec.Version)
	if err != nil || rec2.Version != rec.Version+1 {
		t.Fatalf("cond update: %v", err)
	}
}
