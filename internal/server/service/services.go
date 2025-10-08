package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gophkeeper/internal/server/config"
	"gophkeeper/internal/shared/models"
	"gophkeeper/internal/shared/passhash"
)

type Repository interface {
	CreateUser(ctx context.Context, email string, passwordHash []byte) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (id string, passwordHash []byte, err error)

	UpsertRecord(ctx context.Context, rec models.Record) (models.Record, error)
	UpsertRecordConditional(ctx context.Context, rec models.Record, expectedVersion int64) (models.Record, error)
	ListRecords(ctx context.Context, ownerID string) ([]models.Record, error)
	GetRecord(ctx context.Context, ownerID, id string) (models.Record, error)
	DeleteRecord(ctx context.Context, ownerID, id string) error

	CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (userID string, expiresAt time.Time, err error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

type Services struct {
	Auth    *AuthService
	Records *RecordsService
}

func NewServices(repo Repository, cfg config.Config) *Services {
	return &Services{
		Auth:    &AuthService{repo: repo, jwtSecret: []byte(cfg.JWTSecret)},
		Records: &RecordsService{repo: repo, maxPayloadBytes: cfg.MaxRecordPayloadBytes},
	}
}

// AuthService implements user registration, password verification,
// JWT access token issuance and refresh token rotation.
type AuthService struct {
	repo      Repository
	jwtSecret []byte
}

func (a *AuthService) Register(ctx context.Context, email, password string) (models.User, error) {
	if email == "" || password == "" {
		return models.User{}, errors.New("email and password required")
	}
	phc, err := passhash.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}
	return a.repo.CreateUser(ctx, email, []byte(phc))
}

func (a *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	id, hash, err := a.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	ok, err := passhash.VerifyPassword(string(hash), password)
	if err != nil || !ok {
		return "", errors.New("invalid credentials")
	}
	claims := jwt.MapClaims{
		"sub": id,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(a.jwtSecret)
}

func (a *AuthService) ParseToken(_ context.Context, token string) (string, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", errors.New("invalid token subject")
	}
	return sub, nil
}

func (a *AuthService) IssueAccessToken(userID string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{"sub": userID, "exp": time.Now().Add(ttl).Unix()}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(a.jwtSecret)
}

func (a *AuthService) IssueRefreshToken(ctx context.Context, userID string, ttl time.Duration) (string, error) {
	token := uuid4()
	expires := time.Now().Add(ttl)
	if err := a.repo.CreateRefreshToken(ctx, userID, token, expires); err != nil {
		return "", err
	}
	return token, nil
}

func (a *AuthService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	userID, exp, err := a.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}
	if time.Now().After(exp) {
		_ = a.repo.DeleteRefreshToken(ctx, refreshToken)
		return "", errors.New("refresh token expired")
	}
	// rotate refresh token
	_ = a.repo.DeleteRefreshToken(ctx, refreshToken)
	_, _ = a.IssueRefreshToken(ctx, userID, 30*24*time.Hour)
	return a.IssueAccessToken(userID, 24*time.Hour)
}

// RecordsService stores opaque, client-encrypted payloads with optional
// optimistic concurrency control based on monotonically increasing version.
type RecordsService struct {
	repo            Repository
	maxPayloadBytes int64
}

func (s *RecordsService) Upsert(ctx context.Context, rec models.Record) (models.Record, error) {
	if rec.OwnerID == "" {
		return models.Record{}, errors.New("owner_id required")
	}
	if rec.Type == "" {
		return models.Record{}, errors.New("type required")
	}
	if rec.Meta == nil {
		rec.Meta = map[string]string{}
	}
	if s.maxPayloadBytes > 0 && int64(len(rec.Payload)) > s.maxPayloadBytes {
		return models.Record{}, errors.New("payload too large")
	}
	return s.repo.UpsertRecord(ctx, rec)
}

func (s *RecordsService) UpsertConditional(ctx context.Context, rec models.Record, expectedVersion int64) (models.Record, error) {
	if rec.OwnerID == "" {
		return models.Record{}, errors.New("owner_id required")
	}
	if rec.Type == "" {
		return models.Record{}, errors.New("type required")
	}
	if rec.Meta == nil {
		rec.Meta = map[string]string{}
	}
	if s.maxPayloadBytes > 0 && int64(len(rec.Payload)) > s.maxPayloadBytes {
		return models.Record{}, errors.New("payload too large")
	}
	return s.repo.UpsertRecordConditional(ctx, rec, expectedVersion)
}

func (s *RecordsService) List(ctx context.Context, ownerID string) ([]models.Record, error) {
	return s.repo.ListRecords(ctx, ownerID)
}

func (s *RecordsService) Get(ctx context.Context, ownerID, id string) (models.Record, error) {
	return s.repo.GetRecord(ctx, ownerID, id)
}

func (s *RecordsService) Delete(ctx context.Context, ownerID, id string) error {
	return s.repo.DeleteRecord(ctx, ownerID, id)
}
