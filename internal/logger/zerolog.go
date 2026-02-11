package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type zerologEvent struct{ e *zerolog.Event }

func (z zerologEvent) String(key, value string) Event  { return zerologEvent{z.e.Str(key, value)} }
func (z zerologEvent) Int(key string, value int) Event { return zerologEvent{z.e.Int(key, value)} }
func (z zerologEvent) Error(err error) Event           { return zerologEvent{z.e.Err(err)} }
func (z zerologEvent) Interface(key string, value interface{}) Event {
	return zerologEvent{z.e.Interface(key, value)}
}
func (z zerologEvent) Duration(key string, value time.Duration) Event {
	return zerologEvent{z.e.Dur(key, value)}
}
func (z zerologEvent) Time(key string, value time.Time) Event {
	return zerologEvent{z.e.Time(key, value)}
}
func (z zerologEvent) Message(msg string) { z.e.Msg(msg) }

type zerologLogger struct{ log zerolog.Logger }

func (z zerologLogger) Info() Event  { return zerologEvent{z.log.Info()} }
func (z zerologLogger) Warn() Event  { return zerologEvent{z.log.Warn()} }
func (z zerologLogger) Error() Event { return zerologEvent{z.log.Error()} }
func (z zerologLogger) Debug() Event { return zerologEvent{z.log.Debug()} }

// NewLogger создаёт логгер, реализующий интерфейс Logger.
// Уровень задаётся параметром logLevel (debug|info|warn|error). Пустая строка — info.
func NewLogger(logLevel string) (Logger, error) {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	if logLevel != "" {
		level, err := zerolog.ParseLevel(strings.TrimSpace(strings.ToLower(logLevel)))
		if err != nil {
			return nil, err
		}
		log = log.Level(level)
	}

	return zerologLogger{log: log}, nil
}

// NewNopLogger возвращает логгер, который ничего не выводит (для тестов).
func NewNopLogger() Logger {
	return zerologLogger{log: zerolog.New(io.Discard).Level(zerolog.Disabled)}
}
