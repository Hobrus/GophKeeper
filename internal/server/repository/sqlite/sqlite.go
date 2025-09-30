package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"gophkeeper/internal/server/repository"
	"gophkeeper/internal/shared/models"
)

type Repository struct {
	db *sql.DB
}

func New(dsn string) (*Repository, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash BLOB NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
		CREATE TABLE IF NOT EXISTS records (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL,
			type TEXT NOT NULL,
			meta BLOB NOT NULL,
			payload BLOB NOT NULL,
			version INTEGER NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY(owner_id) REFERENCES users(id)
		);
        CREATE TABLE IF NOT EXISTS refresh_tokens (
            token TEXT PRIMARY KEY,
            user_id TEXT NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            created_at TIMESTAMP NOT NULL,
            FOREIGN KEY(user_id) REFERENCES users(id)
        );
	`); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// Auth

func (r *Repository) CreateUser(ctx context.Context, email string, passwordHash []byte) (models.User, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `INSERT INTO users(id,email,password_hash,created_at) VALUES(?,?,?,?)`, id, email, passwordHash, now)
	if err != nil {
		return models.User{}, err
	}
	return models.User{ID: id, Email: email, CreatedAt: now}, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (id string, passwordHash []byte, err error) {
	row := r.db.QueryRowContext(ctx, `SELECT id,password_hash FROM users WHERE email = ?`, email)
	if err = row.Scan(&id, &passwordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, sql.ErrNoRows
		}
		return "", nil, err
	}
	return
}

// Records

func (r *Repository) UpsertRecord(ctx context.Context, rec models.Record) (models.Record, error) {
	if rec.ID == "" {
		rec.ID = uuid.NewString()
	}
	if rec.Version == 0 {
		rec.Version = 1
	} else {
		rec.Version++
	}
	rec.UpdatedAt = time.Now().UTC()
	metaJSON, _ := json.Marshal(rec.Meta)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO records(id, owner_id, type, meta, payload, version, updated_at)
		VALUES(?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			owner_id=excluded.owner_id,
			type=excluded.type,
			meta=excluded.meta,
			payload=excluded.payload,
			version=excluded.version,
			updated_at=excluded.updated_at
    `, rec.ID, rec.OwnerID, string(rec.Type), metaJSON, rec.Payload, rec.Version, rec.UpdatedAt)
	if err != nil {
		return models.Record{}, err
	}
	return rec, nil
}

func (r *Repository) UpsertRecordConditional(ctx context.Context, rec models.Record, expectedVersion int64) (models.Record, error) {
	now := time.Now().UTC()
	metaJSON, _ := json.Marshal(rec.Meta)
	if rec.ID == "" {
		rec.ID = uuid.NewString()
	}
	// Try insert when expectedVersion == 0 and record not exists
	if expectedVersion == 0 {
		rec.Version = 1
		rec.UpdatedAt = now
		_, err := r.db.ExecContext(ctx, `INSERT INTO records(id, owner_id, type, meta, payload, version, updated_at) VALUES(?,?,?,?,?,?,?)`, rec.ID, rec.OwnerID, string(rec.Type), metaJSON, rec.Payload, rec.Version, rec.UpdatedAt)
		if err == nil {
			return rec, nil
		}
		// If insert failed (exists), fall through to conditional update
	}
	// Conditional update when current version matches expectedVersion
	res, err := r.db.ExecContext(ctx, `UPDATE records SET type=?, meta=?, payload=?, version=?, updated_at=? WHERE id=? AND owner_id=? AND version=?`, string(rec.Type), metaJSON, rec.Payload, expectedVersion+1, now, rec.ID, rec.OwnerID, expectedVersion)
	if err != nil {
		return models.Record{}, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return models.Record{}, repository.ErrVersionConflict
	}
	rec.Version = expectedVersion + 1
	rec.UpdatedAt = now
	return rec, nil
}

func (r *Repository) ListRecords(ctx context.Context, ownerID string) ([]models.Record, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, owner_id, type, meta, payload, version, updated_at FROM records WHERE owner_id = ? ORDER BY updated_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Record
	for rows.Next() {
		var rec models.Record
		var typ string
		var metaBytes []byte
		if err := rows.Scan(&rec.ID, &rec.OwnerID, &typ, &metaBytes, &rec.Payload, &rec.Version, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		rec.Type = models.RecordType(typ)
		if len(metaBytes) > 0 {
			var meta map[string]string
			_ = json.Unmarshal(metaBytes, &meta)
			rec.Meta = meta
		}
		out = append(out, rec)
	}
	return out, nil
}

func (r *Repository) GetRecord(ctx context.Context, ownerID, id string) (models.Record, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, owner_id, type, meta, payload, version, updated_at FROM records WHERE owner_id = ? AND id = ?`, ownerID, id)
	var rec models.Record
	var typ string
	var metaBytes []byte
	if err := row.Scan(&rec.ID, &rec.OwnerID, &typ, &metaBytes, &rec.Payload, &rec.Version, &rec.UpdatedAt); err != nil {
		return models.Record{}, err
	}
	rec.Type = models.RecordType(typ)
	if len(metaBytes) > 0 {
		var meta map[string]string
		_ = json.Unmarshal(metaBytes, &meta)
		rec.Meta = meta
	}
	return rec, nil
}

func (r *Repository) DeleteRecord(ctx context.Context, ownerID, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM records WHERE owner_id = ? AND id = ?`, ownerID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Refresh tokens

func (r *Repository) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO refresh_tokens(token, user_id, expires_at, created_at) VALUES(?,?,?,?)`, token, userID, expiresAt, time.Now().UTC())
	return err
}

func (r *Repository) GetRefreshToken(ctx context.Context, token string) (userID string, expiresAt time.Time, err error) {
	row := r.db.QueryRowContext(ctx, `SELECT user_id, expires_at FROM refresh_tokens WHERE token = ?`, token)
	err = row.Scan(&userID, &expiresAt)
	return
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token = ?`, token)
	return err
}
