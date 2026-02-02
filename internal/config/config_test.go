package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig_EnvAndDefaults(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Setenv("RUN_ADDRESS", "")
	t.Setenv("PORT", "9999")
	t.Setenv("DATABASE_URI", "postgres://x")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_TTL_SECONDS", "")

	os.Args = []string{"cmd"} // no flags
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.RunAddress != ":9999" {
		t.Fatalf("want %q, got %q", ":9999", cfg.RunAddress)
	}
	if cfg.DatabaseURI != "postgres://x" {
		t.Fatalf("unexpected DatabaseURI: %q", cfg.DatabaseURI)
	}
	if cfg.AccrualSystemAddress != "http://accrual" {
		t.Fatalf("unexpected AccrualSystemAddress: %q", cfg.AccrualSystemAddress)
	}
	if cfg.JWTSecret == "" {
		t.Fatalf("expected JWTSecret to be generated")
	}
	if cfg.JWTTTL != 24*time.Hour {
		t.Fatalf("expected default ttl 24h, got %v", cfg.JWTTTL)
	}
}

func TestLoadConfig_FlagsOverride(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Setenv("RUN_ADDRESS", ":1111")
	t.Setenv("DATABASE_URI", "postgres://env")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://env")
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("JWT_TTL_SECONDS", "3600")

	os.Args = []string{"cmd", "-a", ":2222", "-d", "postgres://flag", "-r", "http://flag"}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.RunAddress != ":2222" {
		t.Fatalf("want %q, got %q", ":2222", cfg.RunAddress)
	}
	if cfg.DatabaseURI != "postgres://flag" {
		t.Fatalf("want %q, got %q", "postgres://flag", cfg.DatabaseURI)
	}
	if cfg.AccrualSystemAddress != "http://flag" {
		t.Fatalf("want %q, got %q", "http://flag", cfg.AccrualSystemAddress)
	}
	if cfg.JWTSecret != "s" {
		t.Fatalf("unexpected JWTSecret: %q", cfg.JWTSecret)
	}
	if cfg.JWTTTL != time.Hour {
		t.Fatalf("expected 1h, got %v", cfg.JWTTTL)
	}
}

func TestLoadConfig_DefaultDatabaseURIWhenEmpty(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Setenv("RUN_ADDRESS", "")
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URI", "")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "")
	t.Setenv("JWT_SECRET", "s")
	t.Setenv("JWT_TTL_SECONDS", "")

	os.Args = []string{"cmd"} // no flags
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.RunAddress != ":8080" {
		t.Fatalf("want %q, got %q", ":8080", cfg.RunAddress)
	}
	if cfg.DatabaseURI != "postgres://localhost:5432/postgres?sslmode=disable" {
		t.Fatalf("unexpected DatabaseURI: %q", cfg.DatabaseURI)
	}
}

func TestLoadConfig_BodyLoggingDefaults(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Setenv("LOG_HTTP_BODIES", "")
	t.Setenv("JWT_SECRET", "s")
	os.Args = []string{"cmd"}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.EnableHTTPBodyLogging != false {
		t.Fatalf("expected EnableHTTPBodyLogging=false by default, got %v", cfg.EnableHTTPBodyLogging)
	}
}

func TestLoadConfig_BodyLoggingEnabled(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Setenv("LOG_HTTP_BODIES", "true")
	t.Setenv("JWT_SECRET", "s")
	os.Args = []string{"cmd"}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.EnableHTTPBodyLogging != true {
		t.Fatalf("expected EnableHTTPBodyLogging=true, got %v", cfg.EnableHTTPBodyLogging)
	}
}

func TestLoadConfig_DBPoolSettings(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Setenv("DB_MAX_OPEN_CONNS", "150")
	t.Setenv("DB_MAX_IDLE_CONNS", "30")
	t.Setenv("DB_CONN_MAX_LIFETIME", "600")
	t.Setenv("DB_CONN_MAX_IDLE_TIME", "120")
	t.Setenv("DB_QUERY_TIMEOUT", "5")
	t.Setenv("JWT_SECRET", "s")
	os.Args = []string{"cmd"}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cfg.DBMaxOpenConns != 150 {
		t.Fatalf("expected DBMaxOpenConns=150, got %d", cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != 30 {
		t.Fatalf("expected DBMaxIdleConns=30, got %d", cfg.DBMaxIdleConns)
	}
	if cfg.DBConnMaxLifetime != 600*time.Second {
		t.Fatalf("expected DBConnMaxLifetime=600s, got %v", cfg.DBConnMaxLifetime)
	}
	if cfg.DBConnMaxIdleTime != 120*time.Second {
		t.Fatalf("expected DBConnMaxIdleTime=120s, got %v", cfg.DBConnMaxIdleTime)
	}
	if cfg.DBQueryTimeout != 5*time.Second {
		t.Fatalf("expected DBQueryTimeout=5s, got %v", cfg.DBQueryTimeout)
	}
}

func TestParseBoolEnv(t *testing.T) {
	tests := []struct {
		value    string
		defVal   bool
		expected bool
	}{
		{"", false, false},
		{"", true, true},
		{"true", false, true},
		{"1", false, true},
		{"yes", false, true},
		{"on", false, true},
		{"false", true, false},
		{"0", true, false},
		{"no", true, false},
		{"off", true, false},
		{"invalid", false, false},
		{"invalid", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if tt.value != "" {
				t.Setenv("TEST_BOOL", tt.value)
			} else {
				t.Setenv("TEST_BOOL", "")
			}
			got := parseBoolEnv("TEST_BOOL", tt.defVal)
			if got != tt.expected {
				t.Errorf("parseBoolEnv(%q, %v) = %v, want %v", tt.value, tt.defVal, got, tt.expected)
			}
		})
	}
}
