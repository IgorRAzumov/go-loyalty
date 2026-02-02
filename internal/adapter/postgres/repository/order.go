package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"loyalty/internal/adapter/postgres/util"
	"time"

	ordersmodel "loyalty/internal/domain/order/model"
	ordersrepo "loyalty/internal/domain/order/repository"

	"github.com/shopspring/decimal"
)

// LoyaltyOrdersRepository — PostgreSQL-реализация ordersrepo.OrdersRepository.
type LoyaltyOrdersRepository struct {
	db *sql.DB
}

// NewLoyaltyOrdersRepository создаёт репозиторий заказов на PostgreSQL.
func NewLoyaltyOrdersRepository(db *sql.DB) *LoyaltyOrdersRepository {
	return &LoyaltyOrdersRepository{db: db}
}

// Create создаёт заказ со статусом NEW или возвращает ошибки
func (repository *LoyaltyOrdersRepository) Create(ctx context.Context, userID int64, number string) error {
	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	var existingUserID int64
	var inserted bool
	if err := repository.db.QueryRowContext(
		queryCtx,
		`INSERT INTO orders(number, user_id, status) VALUES ($1, $2, $3)
		 ON CONFLICT (number) DO UPDATE SET number = EXCLUDED.number
		 RETURNING user_id, (xmax = 0) AS inserted`,
		number,
		userID,
		string(ordersmodel.StatusNew),
	).Scan(&existingUserID, &inserted); err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	if inserted {
		return nil
	}
	if existingUserID == userID {
		return ordersmodel.ErrOrderAlreadyUploaded
	}
	return ordersmodel.ErrOrderAlreadyUploadedByAnother
}

// ListByUser возвращает заказы пользователя по времени загрузки (от новых к старым).
func (repository *LoyaltyOrdersRepository) ListByUser(ctx context.Context, userID int64) ([]ordersmodel.Order, error) {
	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	rows, err := repository.db.QueryContext(
		queryCtx,
		`SELECT number, status, accrual, uploaded_at
		   FROM orders
		  WHERE user_id = $1
		  ORDER BY uploaded_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("select orders: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var out []ordersmodel.Order
	for rows.Next() {
		var (
			number     string
			status     string
			accrual    decimal.NullDecimal
			uploadedAt sql.NullTime
		)
		if err := rows.Scan(&number, &status, &accrual, &uploadedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		var accrualPtr *decimal.Decimal
		if accrual.Valid {
			accrualPtr = &accrual.Decimal
		}
		out = append(out, ordersmodel.Order{
			Number:     number,
			UserID:     userID,
			Status:     ordersmodel.Status(status),
			Accrual:    accrualPtr,
			UploadedAt: uploadedAt.Time,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}
	return out, nil
}

// ListPending возвращает все заказы в статусах NEW/PROCESSING для фоновой обработки.
func (repository *LoyaltyOrdersRepository) ListPending(ctx context.Context) ([]ordersmodel.Order, error) {
	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	rows, err := repository.db.QueryContext(
		queryCtx,
		`SELECT number, user_id, status, uploaded_at
		   FROM orders
		  WHERE status IN ($1, $2)
		  ORDER BY uploaded_at ASC`,
		string(ordersmodel.StatusNew),
		string(ordersmodel.StatusProcessing),
	)
	if err != nil {
		return nil, fmt.Errorf("select pending orders: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var out []ordersmodel.Order
	for rows.Next() {
		var (
			number     string
			userID     int64
			status     string
			uploadedAt time.Time
		)
		if err := rows.Scan(&number, &userID, &status, &uploadedAt); err != nil {
			return nil, fmt.Errorf("scan pending order: %w", err)
		}
		out = append(out, ordersmodel.Order{
			Number:     number,
			UserID:     userID,
			Status:     ordersmodel.Status(status),
			UploadedAt: uploadedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending orders: %w", err)
	}
	return out, nil
}

// UpdateFromAccrual обновляет заказ и (идемпотентно) зачисляет начисление на счёт.
func (repository *LoyaltyOrdersRepository) UpdateFromAccrual(
	ctx context.Context,
	number string,
	status ordersmodel.Status,
	accrual *decimal.Decimal,
) error {
	transaction, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	userID, alreadyApplied, err := repository.lockOrder(ctx, transaction, number)
	if err != nil {
		return err
	}
	if userID == 0 {
		return nil
	}

	shouldApplyAccrual := status == ordersmodel.StatusProcessed && !alreadyApplied
	if err := repository.updateOrderStatus(ctx, transaction, number, status, accrual, shouldApplyAccrual); err != nil {
		return err
	}

	if shouldApplyAccrual && accrual != nil && accrual.GreaterThan(decimal.Zero) {
		if err := repository.applyAccrualToAccount(ctx, transaction, userID, *accrual); err != nil {
			return err
		}
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (repository *LoyaltyOrdersRepository) lockOrder(
	ctx context.Context,
	transaction *sql.Tx,
	number string,
) (userID int64, alreadyApplied bool, err error) {
	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	err = transaction.QueryRowContext(
		queryCtx,
		`SELECT user_id, accrual_applied FROM orders WHERE number = $1 FOR UPDATE`,
		number,
	).Scan(&userID, &alreadyApplied)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("lock order: %w", err)
	}
	return userID, alreadyApplied, nil
}

func (repository *LoyaltyOrdersRepository) updateOrderStatus(
	ctx context.Context,
	transaction *sql.Tx,
	number string,
	status ordersmodel.Status,
	accrual *decimal.Decimal,
	setApplied bool,
) error {
	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	var accrualVal any
	if accrual != nil {
		accrualVal = *accrual
	}

	_, err := transaction.ExecContext(
		queryCtx,
		`UPDATE orders
		    SET status = $2,
		        accrual = $3,
		        accrual_applied = $4
		  WHERE number = $1`,
		number,
		string(status),
		accrualVal,
		setApplied,
	)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}
	return nil
}

func (repository *LoyaltyOrdersRepository) applyAccrualToAccount(ctx context.Context, transaction *sql.Tx, userID int64, accrual decimal.Decimal) error {
	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	_, err := transaction.ExecContext(
		queryCtx,
		`UPDATE accounts SET current = current + $2 WHERE user_id = $1`,
		userID,
		accrual,
	)
	if err != nil {
		return fmt.Errorf("apply accrual: %w", err)
	}
	return nil
}

var _ ordersrepo.OrdersRepository = (*LoyaltyOrdersRepository)(nil)
