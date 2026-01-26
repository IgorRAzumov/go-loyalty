// Package app содержит композиционный корень приложения (wiring зависимостей) и запуск рантайма.
package app

import (
	"context"
	"errors"
	"loyalty/internal/adapter/memory"
	tokensvc "loyalty/internal/adapter/token/jwt"
	"loyalty/internal/config"
	"loyalty/internal/domain/auth/service/auth"
	"loyalty/internal/domain/auth/service/user"
	authusecase "loyalty/internal/domain/auth/usecase/auth"
	"loyalty/internal/logger"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"loyalty/internal/controller/httpapi"
)

// Run запускает приложение: инициализирует зависимости, поднимает HTTP-сервер и
// корректно завершает его при отмене контекста.
func Run(ctx context.Context) error {
	initLogger()
	appConfig := loadConfig()

	tokenService := tokensvc.NewTokenService(appConfig.JWTSecret, appConfig.JWTTTL)
	authService := auth.NewAuthService()
	authUsecase := authusecase.NewUsecase(user.NewUserService(memory.NewRepository()), authService, tokenService)

	server, errChannel := httpapi.StartServer(appConfig, httpapi.AuthDeps{
		AuthUsecase:  authUsecase,
		TokenService: tokenService,
	})

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("http server shutdown failed")
		}
		return nil
	case err := <-errChannel:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func initLogger() {
	if err := logger.InitLogger(); err != nil {
		os.Exit(2)
	}
}

func loadConfig() config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error().Err(err).Msg("invalid flags")
		os.Exit(2)
	}
	return cfg
}
