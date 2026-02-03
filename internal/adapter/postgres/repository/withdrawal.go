package repository

import (
	"context"
	"database/sql"
	"fmt"
	"loyalty/internal/adapter/postgres/util"
	"time"

	withdrawalsmodel "loyalty/internal/domain/withdrawal/model"
	withdrawalsrepo "loyalty/internal/domain/withdrawal/repository"

	"github.com/shopspring/decimal"
)

// LoyaltyWithdrawalsRepository — PostgreSQL-реализация withdrawalsrepo.WithdrawalsRepository.
type LoyaltyWithdrawalsRepository struct {
	db *sql.DB
}

// NewLoyaltyWithdrawalsRepository создаёт репозиторий списаний на PostgreSQL.
func NewLoyaltyWithdrawalsRepository(db *sql.DB) *LoyaltyWithdrawalsRepository {
	return &LoyaltyWithdrawalsRepository{db: db}
}

// ListByUser возвращает списания пользователя (от новых к старым).
func (r *LoyaltyWithdrawalsRepository) ListByUser(ctx context.Context, userID int64) ([]withdrawalsmodel.Withdrawal, error) {
	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	rows, err := r.db.QueryContext(
		queryCtx,
		`SELECT order_number, sum, processed_at
		   FROM withdrawals
		  WHERE user_id = $1
		  ORDER BY processed_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("select withdrawals: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var out []withdrawalsmodel.Withdrawal
	for rows.Next() {
		var (
			orderNumber string
			sum         decimal.Decimal
			processedAt time.Time
		)
		if err := rows.Scan(&orderNumber, &sum, &processedAt); err != nil {
			return nil, fmt.Errorf("scan withdrawal: %w", err)
		}
		out = append(out, withdrawalsmodel.Withdrawal{
			UserID:      userID,
			OrderNumber: orderNumber,
			Sum:         sum,
			ProcessedAt: processedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate withdrawals: %w", err)
	}
	return out, nil
}

var _ withdrawalsrepo.WithdrawalsRepository = (*LoyaltyWithdrawalsRepository)(nil)
