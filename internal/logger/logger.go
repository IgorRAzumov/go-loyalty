// Package logger инициализирует логирование приложения (zerolog).
package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger инициализирует глобальный логгер zerolog.
// Уровень логирования задаётся через переменную окружения LOG_LEVEL (debug|info|warn|error).
func InitLogger() error {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.Output(os.Stderr).With().Timestamp().Logger()

	if v := strings.TrimSpace(os.Getenv("LOG_LEVEL")); v != "" {
		level, err := zerolog.ParseLevel(strings.ToLower(v))
		if err != nil {
			return err
		}
		zerolog.SetGlobalLevel(level)
	}

	return nil
}
