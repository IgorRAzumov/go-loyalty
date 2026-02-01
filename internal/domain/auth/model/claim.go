package model

import "time"

// Claim — доменные claims (данные), которые мы кладём в access-token и извлекаем из него.
type Claim struct {
	UserID    int64     `json:"uid"`
	Login     string    `json:"login"`
	IssuedAt  time.Time `json:"iat,omitempty"`
	ExpiresAt time.Time `json:"exp,omitempty"`
}
