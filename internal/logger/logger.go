package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger инициализирует глобальный логгер zerolog.
// Уровень логирования задаётся параметром logLevel (debug|info|warn|error).
func InitLogger(logLevel string) error {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.Output(os.Stderr).With().Timestamp().Logger()

	if logLevel != "" {
		level, err := zerolog.ParseLevel(strings.ToLower(logLevel))
		if err != nil {
			return err
		}
		zerolog.SetGlobalLevel(level)
	}

	return nil
}
