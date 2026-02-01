package jwt

import (
	"time"

	"loyalty/internal/domain/auth/model"
	"loyalty/internal/domain/auth/service"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type claims struct {
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
		parser: jwtlib.NewParser(jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()})),
	}
}

// IssueToken выпускает access-token для пользователя.
func (s *Service) IssueToken(userID int64, login string, now time.Time) (string, error) {
	if userID <= 0 {
		return "", model.ErrNotFound
	}
	c := claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   login,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(s.ttl)),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, c)
	return token.SignedString(s.secret)
}

// ParseToken проверяет и парсит access-token.
func (s *Service) ParseToken(token string) (*model.Claim, error) {
	var c claims
	parsed, err := s.parser.ParseWithClaims(token, &c, s.keyFunc)
	if err != nil {
		return nil, err
	}
	if !parsed.Valid || c.UserID <= 0 {
		return nil, model.ErrInvalidToken
	}
	return &model.Claim{
		UserID:    c.UserID,
		Login:     c.Login,
		IssuedAt:  c.IssuedAt.Time,
		ExpiresAt: c.ExpiresAt.Time,
	}, nil
}

func (s *Service) keyFunc(_ *jwtlib.Token) (any, error) {
	return s.secret, nil
}

var _ service.TokenService = (*Service)(nil)
