package app

import (
	"context"
	"database/sql"
	"errors"
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

	"github.com/rs/zerolog/log"

	"loyalty/internal/controller/httpapi"
)

// Run запускает приложение: инициализирует зависимости, поднимает HTTP-сервер,
// запускает accrual воркер и корректно завершает их при отмене контекста.
func Run(ctx context.Context) error {
	appConfig := loadConfig()
	initLogger(appConfig.LogLevel)

	db, errDb := initDb(ctx, appConfig)
	if errDb != nil {
		return errDb
	}
	defer func() { _ = db.Close() }()

	dependencies, worker := loadDependencies(appConfig, db)
	server, errChannel := httpapi.StartServer(appConfig, dependencies)

	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	go worker.Start(workerCtx)

	select {
	case <-ctx.Done():
		workerCancel() // Останавливаем воркер
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("http server shutdown failed")
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

func initDb(ctx context.Context, appConfig config.Config) (*sql.DB, error) {
	util.SetQueryTimeout(appConfig.DBQueryTimeout)

	poolCfg := postgres.PoolConfig{
		MaxOpenConns:    appConfig.DBMaxOpenConns,
		MaxIdleConns:    appConfig.DBMaxIdleConns,
		ConnMaxLifetime: appConfig.DBConnMaxLifetime,
		ConnMaxIdleTime: appConfig.DBConnMaxIdleTime,
	}
	db, err := postgres.OpenWithMigrations(ctx, appConfig.DatabaseURI, poolCfg)
	if err != nil {
		log.Error().Err(err).Msg("database init failed")
		return nil, err
	}
	log.Info().
		Int("max_open_conns", poolCfg.MaxOpenConns).
		Int("max_idle_conns", poolCfg.MaxIdleConns).
		Dur("conn_max_lifetime", poolCfg.ConnMaxLifetime).
		Dur("conn_max_idle_time", poolCfg.ConnMaxIdleTime).
		Dur("query_timeout", appConfig.DBQueryTimeout).
		Msg("database connection pool configured")
	return db, nil
}

func loadDependencies(appConfig config.Config, db *sql.DB) (httpapi.Deps, *accrualworker.Worker) {
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

	accrualClient := createAccrualClient(appConfig)
	worker := accrualworker.NewWorker(ordersRepo, ordersService, accrualClient, accrualworker.DefaultConfig())

	return httpapi.Deps{
		AuthUsecase:           authusecase.NewUsecase(user.NewUserService(authRepo), authService, tokenService),
		OrdersUsecase:         orderusecase.NewUsecase(ordersService),
		BalanceUsecase:        balanceuc.NewUsecase(balanceService),
		WithdrawalsUsecase:    withdrawalusecase.NewUsecase(withdrawalsService, numberValidator),
		TokenService:          tokenService,
		EnableHTTPBodyLogging: appConfig.EnableHTTPBodyLogging,
		AuthRateLimitRPS:      appConfig.AuthRateLimitRPS,
		AuthRateLimitBurst:    appConfig.AuthRateLimitBurst,
	}, worker
}

func initLogger(logLevel string) {
	if err := logger.InitLogger(logLevel); err != nil {
		os.Exit(2)
	}
}

func loadConfig() config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error().Err(err).Msg("invalid flags")
		os.Exit(2)
	}
	return cfg
}

// createAccrualClient создаёт клиент для системы accrual (HTTP или mock).
func createAccrualClient(cfg config.Config) accrualclient.AccrualClient {
	if cfg.AccrualSystemAddress == "" {
		log.Warn().Msg("accrual system address not configured, using mock client")
		return accrualmock.NewClientWithDefaults()
	}

	log.Info().Str("address", cfg.AccrualSystemAddress).Msg("using HTTP accrual client")
	return accrualhttp.NewClient(cfg.AccrualSystemAddress, 5*time.Second)
}
