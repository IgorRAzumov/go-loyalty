package httpapi

import (
	"loyalty/internal/config"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
)

// StartServer поднимает HTTP-сервер и запускает его в отдельной горутине.
// Возвращает сам сервер (для Shutdown) и канал, в который будет отправлена ошибка ListenAndServe.
func StartServer(appConfig config.Config, deps Deps) (*http.Server, <-chan error) {
	srv := &http.Server{
		Addr:              appConfig.RunAddress,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		Handler:           InitRouter(deps),
	}

	errChannel := make(chan error, 1)
	go func() {
		if appConfig.JWTSecret == "" {
			log.Warn().Msg("JWT_SECRET is empty; auth tokens may be insecure")
		}
		log.Info().Str("addr", appConfig.RunAddress).Msg("starting http server listening")

		if err := srv.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("http server failed")
			errChannel <- err
		}
	}()

	return srv, errChannel
}
