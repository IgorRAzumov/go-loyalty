package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	accrualhttp "loyalty/internal/adapter/accrual/http"
	accrualmock "loyalty/internal/adapter/accrual/mock"
	"loyalty/internal/adapter/postgres"
	postgresrepo "loyalty/internal/adapter/postgres/repository"
	"loyalty/internal/adapter/postgres/util"
	tokensvc "loyalty/internal/adapter/token/jwt"
	"loyalty/internal/config"
	accrualclient "loyalty/internal/domain/accrual/client"
	"loyalty/internal/domain/auth/service/auth"
	"loyalty/internal/domain/auth/service/user"
	authusecase "loyalty/internal/domain/auth/usecase/auth"
	balanceappsvc "loyalty/internal/domain/balance/service/balance"
	balanceuc "loyalty/internal/domain/balance/usecase/balance"
	ordersappsvc "loyalty/internal/domain/order/service/orders"
	ordervalidator "loyalty/internal/domain/order/service/validator"
	orderusecase "loyalty/internal/domain/order/usecase/order"
	withdrawalsappsvc "loyalty/internal/domain/withdrawal/service/withdrawals"
	withdrawalusecase "loyalty/internal/domain/withdrawal/usecase/withdrawals"
	"loyalty/internal/logger"
	accrualworker "loyalty/internal/worker/accrual"
	"net/http"
	"os"
	"time"

	"loyalty/internal/controller/httpapi"
)

// Run запускает приложение: инициализирует зависимости, поднимает HTTP-сервер,
// запускает accrual воркер и корректно завершает их при отмене контекста.
func Run(ctx context.Context) error {
	appConfig := loadConfig()

	log, errLog := logger.NewLogger(appConfig.LogLevel)
	if errLog != nil {
		os.Exit(2)
	}

	db, errDb := initDb(ctx, appConfig, log)
	if errDb != nil {
		return errDb
	}
	defer func() { _ = db.Close() }()

	dependencies, worker := loadDependencies(appConfig, db, log)
	server, errChannel := httpapi.StartServer(appConfig, dependencies)

	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	go worker.Start(workerCtx)

	select {
	case <-ctx.Done():
		workerCancel()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error().Error(err).Message("http server shutdown failed")
		}
		return nil
	case err := <-errChannel:
		workerCancel()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func initDb(ctx context.Context, appConfig config.Config, log logger.Logger) (*sql.DB, error) {
	util.SetQueryTimeout(appConfig.DBQueryTimeout)

	poolCfg := postgres.PoolConfig{
		MaxOpenConns:    appConfig.DBMaxOpenConns,
		MaxIdleConns:    appConfig.DBMaxIdleConns,
		ConnMaxLifetime: appConfig.DBConnMaxLifetime,
		ConnMaxIdleTime: appConfig.DBConnMaxIdleTime,
	}
	db, err := postgres.OpenWithMigrations(ctx, appConfig.DatabaseURI, poolCfg, log)
	if err != nil {
		log.Error().Error(err).Message("database init failed")
		return nil, err
	}
	log.Info().
		Int("max_open_conns", poolCfg.MaxOpenConns).
		Int("max_idle_conns", poolCfg.MaxIdleConns).
		Duration("conn_max_lifetime", poolCfg.ConnMaxLifetime).
		Duration("conn_max_idle_time", poolCfg.ConnMaxIdleTime).
		Duration("query_timeout", appConfig.DBQueryTimeout).
		Message("database connection pool configured")
	return db, nil
}

func loadDependencies(appConfig config.Config, db *sql.DB, log logger.Logger) (httpapi.Deps, *accrualworker.Worker) {
	authRepo := postgresrepo.NewAuthUserRepository(db)
	ordersRepo := postgresrepo.NewLoyaltyOrdersRepository(db)
	accountRepo := postgresrepo.NewLoyaltyAccountRepository(db)
	withdrawalsRepo := postgresrepo.NewLoyaltyWithdrawalsRepository(db)

	tokenService := tokensvc.NewTokenService(appConfig.JWTSecret, appConfig.JWTTTL)
	authService := auth.NewAuthService()
	numberValidator := ordervalidator.NewValidator()
	ordersService := ordersappsvc.NewService(ordersRepo, numberValidator)
	balanceService := balanceappsvc.NewService(accountRepo)
	withdrawalsService := withdrawalsappsvc.NewService(accountRepo, withdrawalsRepo)

	accrualClient := createAccrualClient(appConfig, log)
	worker := accrualworker.NewWorker(ordersRepo, ordersService, accrualClient, accrualworker.DefaultConfig(), log)

	return httpapi.Deps{
		AuthUsecase:           authusecase.NewUsecase(user.NewUserService(authRepo), authService, tokenService),
		OrdersUsecase:         orderusecase.NewUsecase(ordersService),
		BalanceUsecase:        balanceuc.NewUsecase(balanceService),
		WithdrawalsUsecase:    withdrawalusecase.NewUsecase(withdrawalsService, numberValidator),
		TokenService:          tokenService,
		Logger:                log,
		EnableHTTPBodyLogging: appConfig.EnableHTTPBodyLogging,
		AuthRateLimitRPS:      appConfig.AuthRateLimitRPS,
		AuthRateLimitBurst:    appConfig.AuthRateLimitBurst,
	}, worker
}

func loadConfig() config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid flags: %v\n", err)
		os.Exit(2)
	}
	return cfg
}

// createAccrualClient создаёт клиент для системы accrual (HTTP или mock).
func createAccrualClient(cfg config.Config, log logger.Logger) accrualclient.AccrualClient {
	if cfg.AccrualSystemAddress == "" {
		log.Warn().Message("accrual system address not configured, using mock client")
		return accrualmock.NewClientWithDefaults()
	}

	log.Info().String("address", cfg.AccrualSystemAddress).Message("using HTTP accrual client")
	return accrualhttp.NewClient(cfg.AccrualSystemAddress, 5*time.Second)
}
