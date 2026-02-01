package auth

import (
	"context"
	"errors"
	"loyalty/internal/domain/auth/model"
	"loyalty/internal/domain/auth/service"
	uc "loyalty/internal/domain/auth/usecase"
	"time"
)

// Usecase — сценарии аутентификации (оркестрация сервисов пользователя/паролей/токенов).
type Usecase struct {
	userService  service.UserService
	authService  service.AuthService
	tokenService service.TokenService
}

// NewUsecase создаёт usecase аутентификации с зависимостями на сервисы домена и токенов.
func NewUsecase(userService service.UserService, authService service.AuthService, tokenService service.TokenService) *Usecase {
	return &Usecase{
		userService:  userService,
		authService:  authService,
		tokenService: tokenService,
	}
}

// Register регистрирует пользователя и возвращает access-token.
func (usecase *Usecase) Register(ctx context.Context, login, password string) (token string, _ error) {
	hash, err := usecase.authService.HashPassword(password)
	if err != nil {
		return "", err
	}
	user, err := usecase.userService.CreateUser(ctx, login, hash)
	if err != nil {
		return "", err
	}
	return usecase.tokenService.IssueToken(user.ID, user.Login, time.Now())
}

// Login аутентифицирует пользователя и возвращает access-token.
func (usecase *Usecase) Login(ctx context.Context, login, password string) (token string, _ error) {
	user, err := usecase.userService.FindUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}
	if err := usecase.authService.ComparePassword(user.PasswordHash, password); err != nil {
		if errors.Is(err, model.ErrInvalidInput) || errors.Is(err, model.ErrPasswordTooShort) {
			return "", err
		}
		return "", model.ErrInvalidCreds
	}
	return usecase.tokenService.IssueToken(user.ID, user.Login, time.Now())
}

var _ uc.AuthUsecase = (*Usecase)(nil)
