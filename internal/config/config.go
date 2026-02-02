package config

import (
	"flag"
	"io"
	"loyalty/internal/util/auth"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config содержит параметры запуска и подключения к внешним зависимостям.
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string

	JWTSecret string
	JWTTTL    time.Duration

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	EnableHTTPBodyLogging bool

	DBQueryTimeout time.Duration

	AuthRateLimitRPS   int
	AuthRateLimitBurst int

	LogLevel string
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
		RunAddress:            runAddr,
		DatabaseURI:           os.Getenv("DATABASE_URI"),
		AccrualSystemAddress:  os.Getenv("ACCRUAL_SYSTEM_ADDRESS"),
		JWTSecret:             jwtSecret,
		JWTTTL:                jwtTTL,
		DBMaxOpenConns:        parseIntEnv("DB_MAX_OPEN_CONNS", 100),
		DBMaxIdleConns:        parseIntEnv("DB_MAX_IDLE_CONNS", 25),
		DBConnMaxLifetime:     parseDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		DBConnMaxIdleTime:     parseDurationEnv("DB_CONN_MAX_IDLE_TIME", 1*time.Minute),
		DBQueryTimeout:        parseDurationEnv("DB_QUERY_TIMEOUT", 3*time.Second),
		EnableHTTPBodyLogging: parseBoolEnv("LOG_HTTP_BODIES", false),
		AuthRateLimitRPS:      parseIntEnv("AUTH_RATE_LIMIT_RPS", 100),
		AuthRateLimitBurst:    parseIntEnv("AUTH_RATE_LIMIT_BURST", 20),
		LogLevel:              strings.TrimSpace(os.Getenv("LOG_LEVEL")),
	}

	if err := applyFlags(&cfg, os.Args[1:]); err != nil {
		return Config{}, err
	}

	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = "postgres://localhost:5432/postgres?sslmode=disable"
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

func parseIntEnv(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}

func parseDurationEnv(key string, defaultValue time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	seconds, err := strconv.Atoi(val)
	if err != nil || seconds <= 0 {
		return defaultValue
	}
	return time.Duration(seconds) * time.Second
}

func parseBoolEnv(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	switch val {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}
