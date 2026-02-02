package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"loyalty/internal/adapter/postgres/util"
	"time"

	balancemodel "loyalty/internal/domain/balance/model"
	balancerepo "loyalty/internal/domain/balance/repository"
	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"
	withdrawalsrepo "loyalty/internal/domain/withdrawal/repository"

	"github.com/shopspring/decimal"
)

// LoyaltyAccountRepository — PostgreSQL-реализация balance/withdrawals репозиториев счёта.
type LoyaltyAccountRepository struct {
	db *sql.DB
}

// NewLoyaltyAccountRepository создаёт репозиторий счетов на PostgreSQL.
func NewLoyaltyAccountRepository(db *sql.DB) *LoyaltyAccountRepository {
	return &LoyaltyAccountRepository{db: db}
}

// GetBalance возвращает баланс пользователя.
func (repository *LoyaltyAccountRepository) GetBalance(ctx context.Context, userID int64) (balancemodel.Balance, error) {
	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	var current, withdrawn decimal.Decimal
	if err := repository.db.QueryRowContext(
		queryCtx,
		`SELECT current, withdrawn FROM accounts WHERE user_id = $1`,
		userID,
	).Scan(&current, &withdrawn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ensureAccountExists(ctx, userID)
		}
		return balancemodel.Balance{}, fmt.Errorf("select account: %w", err)
	}
	return balancemodel.Balance{Current: current, Withdrawn: withdrawn}, nil
}

// Withdraw списывает сумму с баланса (атомарно), создавая запись о списании.
func (repository *LoyaltyAccountRepository) Withdraw(
	ctx context.Context,
	userID int64,
	orderNumber string,
	sum decimal.Decimal,
	now time.Time,
) error {
	if sum.LessThanOrEqual(decimal.Zero) {
		return withdrawalsmodel.ErrInvalidWithdrawSum
	}

	transaction, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	if exists, err := repository.checkExistingWithdrawal(ctx, transaction, orderNumber); err != nil {
		return err
	} else if exists {
		return nil
	}

	current, withdrawn, err := repository.getBalanceForWithdrawal(ctx, transaction, userID)
	if err != nil {
		return err
	}

	if current.LessThan(sum) {
		return withdrawalsmodel.ErrInsufficientFunds
	}

	if err := repository.executeWithdrawal(ctx, transaction, userID, orderNumber, current, withdrawn, sum, now); err != nil {
		return err
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (repository *LoyaltyAccountRepository) checkExistingWithdrawal(ctx context.Context, transaction *sql.Tx, orderNumber string) (bool, error) {
	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	var existingID int64
	err := transaction.QueryRowContext(
		queryCtx,
		`SELECT id FROM withdrawals WHERE order_number = $1`,
		orderNumber,
	).Scan(&existingID)

	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, fmt.Errorf("check existing withdrawal: %w", err)
}

func (repository *LoyaltyAccountRepository) getBalanceForWithdrawal(
	ctx context.Context,
	transaction *sql.Tx,
	userID int64,
) (current, withdrawn decimal.Decimal, err error) {
	err = repository.upsertAndGetAccount(ctx, transaction.QueryRowContext, userID, &current, &withdrawn)
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("getBalance: %w", err)
	}
	return current, withdrawn, nil
}

func (repository *LoyaltyAccountRepository) executeWithdrawal(
	ctx context.Context,
	transaction *sql.Tx,
	userID int64,
	orderNumber string,
	current, withdrawn, sum decimal.Decimal,
	now time.Time,
) error {
	newCurrent := current.Sub(sum)
	newWithdrawn := withdrawn.Add(sum)

	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	_, err := transaction.ExecContext(
		queryCtx,
		`WITH updated AS (
		   UPDATE accounts
		   SET current = $2, withdrawn = $3
		   WHERE user_id = $1
		   RETURNING user_id
		 )
		 INSERT INTO withdrawals(user_id, order_number, sum, processed_at)
		 SELECT $1, $4, $5, $6 FROM updated`,
		userID,
		newCurrent,
		newWithdrawn,
		orderNumber,
		sum,
		now,
	)
	if err != nil {
		return fmt.Errorf("execute withdrawal: %w", err)
	}

	return nil
}

func (repository *LoyaltyAccountRepository) ensureAccountExists(ctx context.Context, userID int64) (balancemodel.Balance, error) {
	var current, withdrawn decimal.Decimal
	if err := repository.upsertAndGetAccount(ctx, repository.db.QueryRowContext, userID, &current, &withdrawn); err != nil {
		return balancemodel.Balance{}, fmt.Errorf("init account: %w", err)
	}
	return balancemodel.Balance{Current: current, Withdrawn: withdrawn}, nil
}

func (repository *LoyaltyAccountRepository) upsertAndGetAccount(
	ctx context.Context,
	queryRowFunc func(context.Context, string, ...any) *sql.Row,
	userID int64,
	current, withdrawn *decimal.Decimal,
) error {
	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	return queryRowFunc(
		queryCtx,
		`INSERT INTO accounts(user_id, current, withdrawn)
		 VALUES ($1, 0, 0)
		 ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
		 RETURNING current, withdrawn`,
		userID,
	).Scan(current, withdrawn)
}

var _ balancerepo.BalanceRepository = (*LoyaltyAccountRepository)(nil)
var _ withdrawalsrepo.AccountRepository = (*LoyaltyAccountRepository)(nil)
