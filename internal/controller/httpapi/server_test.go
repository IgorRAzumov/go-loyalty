package httpapi

import (
	"context"
	tokensvc "loyalty/internal/adapter/token/jwt"
	"net/http"
	"testing"
	"time"

	"loyalty/internal/config"
)

func TestStartServer_Shutdown(t *testing.T) {
	cfg := config.Config{
		RunAddress: ":0",
		JWTSecret:  "secret",
		JWTTTL:     time.Hour,
	}

	svc := tokensvc.NewTokenService("secret", time.Hour)
	srv, errCh := StartServer(cfg, AuthDeps{
		AuthUsecase: &mockAuthUsecase{
			registerFn: func(context.Context, string, string) (string, error) { return "", nil },
			loginFn:    func(context.Context, string, string) (string, error) { return "", nil },
		},
		TokenService: svc,
	})
	if srv == nil || errCh == nil {
		t.Fatalf("expected server and channel")
	}
	if srv.ReadTimeout != readTimeout || srv.WriteTimeout != writeTimeout || srv.IdleTimeout != idleTimeout {
		t.Fatalf("unexpected timeouts")
	}

	// Let ListenAndServe start.
	time.Sleep(20 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	select {
	case err := <-errCh:
		// We accept both nil and ErrServerClosed here due to timing.
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("unexpected err: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for server error")
	}
}
