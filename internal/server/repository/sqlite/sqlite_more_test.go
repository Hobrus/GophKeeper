package sqlite

import (
	"context"
	"testing"

	"gophkeeper/internal/shared/models"
)

func TestGetUserByEmail_NoRows(t *testing.T) {
	repo, _ := New("file:repo_no_user?mode=memory&cache=shared")
	if _, _, err := repo.GetUserByEmail(context.Background(), "none@example.com"); err == nil {
		t.Fatalf("expected error for missing user")
	}
}

func TestDeleteRecord_NoRows(t *testing.T) {
	repo, _ := New("file:repo_delete_no_rows?mode=memory&cache=shared")
	ctx := context.Background()
	u, _ := repo.CreateUser(ctx, "u@example.com", []byte("h"))
	if err := repo.DeleteRecord(ctx, u.ID, "no-such"); err == nil {
		t.Fatalf("expected sql.ErrNoRows, got nil")
	}
}

func TestListRecords_MetaDecode(t *testing.T) {
	repo, _ := New("file:repo_list_meta?mode=memory&cache=shared")
	ctx := context.Background()
	u, _ := repo.CreateUser(ctx, "m@example.com", []byte("h"))
	_, err := repo.UpsertRecord(ctx, models.Record{OwnerID: u.ID, Type: models.RecordTypeText, Meta: map[string]string{"k": "v"}, Payload: []byte("x")})
	if err != nil {
		t.Fatal(err)
	}
	list, err := repo.ListRecords(ctx, u.ID)
	if err != nil || len(list) != 1 {
		t.Fatalf("list err: %v", err)
	}
	if list[0].Meta["k"] != "v" {
		t.Fatalf("meta not decoded")
	}
}
