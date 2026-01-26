// Package jwt содержит инфраструктурный сервис токенов (JWT) для TokenService.
package jwt

import (
	"time"

	"loyalty/internal/domain/auth/model"
	"loyalty/internal/domain/auth/service"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type jwtClaims struct {
	UserID int64  `json:"uid"`
	Login  string `json:"login"`
	jwtlib.RegisteredClaims
}

// Service — JWT-реализация service.TokenService (HS256).
type Service struct {
	secret []byte
	ttl    time.Duration
	parser *jwtlib.Parser
}

// NewTokenService создаёт JWT-сервис токенов с заданным секретом и TTL.
func NewTokenService(secret string, ttl time.Duration) *Service {
	return &Service{
		secret: []byte(secret),
		ttl:    ttl,
		parser: jwtlib.NewParser(
			jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
		),
	}
}

// IssueToken выпускает access-token для пользователя.
func (service *Service) IssueToken(userID int64, login string, now time.Time) (string, error) {
	if userID <= 0 {
		return "", model.ErrNotFound
	}
	claims := jwtClaims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   login,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(service.ttl)),
		},
	}
	tok := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return tok.SignedString(service.secret)
}

// ParseToken проверяет и парсит access-token.
func (service *Service) ParseToken(token string) (*model.Claim, error) {
	var claims jwtClaims
	parsed, err := service.parser.ParseWithClaims(token, &claims, func(_ *jwtlib.Token) (any, error) {
		return service.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid || claims.UserID <= 0 {
		return nil, model.ErrInvalidToken
	}

	var issuedAt time.Time
	if claims.IssuedAt != nil {
		issuedAt = claims.IssuedAt.Time
	}
	var expiresAt time.Time
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}
	return &model.Claim{
		UserID:    claims.UserID,
		Login:     claims.Login,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}, nil
}

var _ service.TokenService = (*Service)(nil)
