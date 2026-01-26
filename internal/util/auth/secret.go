package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// RandomSecret генерирует случайный секрет, пригодный для JWT/HMAC (URL-safe base64 без padding).
func RandomSecret() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(b[:])
}
