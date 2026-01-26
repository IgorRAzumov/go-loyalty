// Package config отвечает за конфигурацию приложения (env + флаги) и её валидацию.
package config

import (
	"flag"
	"io"
	"loyalty/internal/util/auth"
	"os"
	"strconv"
	"time"
)

// Config содержит параметры запуска и подключения к внешним зависимостям.
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string

	JWTSecret string
	JWTTTL    time.Duration
}

// LoadConfig загружает конфигурацию из env и CLI-флагов.
// Приоритет: флаги (-a/-d/-r) перекрывают переменные окружения.
func LoadConfig() (Config, error) {
	runAddr := os.Getenv("RUN_ADDRESS")
	if runAddr == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		runAddr = ":" + port
	}

	jwtTTL := 24 * time.Hour
	if ttlEnv := os.Getenv("JWT_TTL_SECONDS"); ttlEnv != "" {
		if sec, err := strconv.Atoi(ttlEnv); err == nil && sec > 0 {
			jwtTTL = time.Duration(sec) * time.Second
		}
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = auth.RandomSecret()
	}

	cfg := Config{
		RunAddress:           runAddr,
		DatabaseURI:          os.Getenv("DATABASE_URI"),
		AccrualSystemAddress: os.Getenv("ACCRUAL_SYSTEM_ADDRESS"),
		JWTSecret:            jwtSecret,
		JWTTTL:               jwtTTL,
	}

	if err := applyFlags(&cfg, os.Args[1:]); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func applyFlags(cfg *Config, args []string) error {
	flagSet := flag.NewFlagSet("app", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	a := flagSet.String("a", "", "service run address (overrides RUN_ADDRESS)")
	d := flagSet.String("d", "", "database uri (overrides DATABASE_URI)")
	r := flagSet.String("r", "", "accrual system address (overrides ACCRUAL_SYSTEM_ADDRESS)")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if *a != "" {
		cfg.RunAddress = *a
	}
	if *d != "" {
		cfg.DatabaseURI = *d
	}
	if *r != "" {
		cfg.AccrualSystemAddress = *r
	}

	return nil
}
