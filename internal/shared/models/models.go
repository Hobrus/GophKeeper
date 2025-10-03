package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type RecordType string

const (
	RecordTypeLogin    RecordType = "login"
	RecordTypeText     RecordType = "text"
	RecordTypeBinary   RecordType = "binary"
	RecordTypeBankCard RecordType = "bank_card"
)

type Record struct {
	ID        string            `json:"id"`
	OwnerID   string            `json:"owner_id"`
	Type      RecordType        `json:"type"`
	Meta      map[string]string `json:"meta"`
	Payload   []byte            `json:"payload"`
	Version   int64             `json:"version"`
	UpdatedAt time.Time         `json:"updated_at"`
}
