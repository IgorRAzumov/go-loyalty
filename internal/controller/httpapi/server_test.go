package httpapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	tokensvc "loyalty/internal/adapter/token/jwt"
	"loyalty/internal/config"

	"go.uber.org/mock/gomock"
)

func TestStartServer_Shutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := config.Config{
		RunAddress: ":0",
		JWTSecret:  "secret",
		JWTTTL:     time.Hour,
	}

	svc := tokensvc.NewTokenService("secret", time.Hour)
	deps := defaultDeps(t, ctrl)
	deps.TokenService = svc

	srv, errCh := StartServer(cfg, deps)
	if srv == nil || errCh == nil {
		t.Fatalf("expected server and channel")
	}
	if srv.ReadTimeout != readTimeout || srv.WriteTimeout != writeTimeout || srv.IdleTimeout != idleTimeout {
		t.Fatalf("unexpected timeouts")
	}

	time.Sleep(20 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("unexpected err: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for server error")
	}
}
