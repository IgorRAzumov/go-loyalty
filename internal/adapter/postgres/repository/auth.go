package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"loyalty/internal/adapter/postgres/util"

	authmodel "loyalty/internal/domain/auth/model"
	authrepo "loyalty/internal/domain/auth/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

// AuthUserRepository — PostgreSQL-реализация authrepo.UserRepository.
type AuthUserRepository struct {
	db *sql.DB
}

// NewAuthUserRepository создаёт репозиторий пользователей для подсистемы auth на PostgreSQL.
func NewAuthUserRepository(db *sql.DB) *AuthUserRepository {
	return &AuthUserRepository{db: db}
}

// Create создаёт пользователя и инициализирует его накопительный счёт.
func (repository *AuthUserRepository) Create(ctx context.Context, login string, passwordHash []byte) (authmodel.User, error) {
	transaction, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return authmodel.User{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()

	insertCtx, insertCancel := util.WithQueryTimeout(ctx)
	defer insertCancel()

	var id int64
	if err := transaction.QueryRowContext(
		insertCtx,
		`INSERT INTO users(login, password_hash) VALUES ($1, $2) RETURNING id`,
		login,
		passwordHash,
	).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			return authmodel.User{}, authmodel.ErrLoginTaken
		}
		return authmodel.User{}, fmt.Errorf("insert user: %w", err)
	}

	accountCtx, accountCancel := util.WithQueryTimeout(ctx)
	defer accountCancel()
	if _, err := transaction.ExecContext(accountCtx, `INSERT INTO accounts(user_id) VALUES ($1) ON CONFLICT DO NOTHING`, id); err != nil {
		return authmodel.User{}, fmt.Errorf("init account: %w", err)
	}

	if err := transaction.Commit(); err != nil {
		return authmodel.User{}, fmt.Errorf("commit: %w", err)
	}
	return authmodel.User{ID: id, Login: login, PasswordHash: passwordHash}, nil
}

// FindByLogin возвращает пользователя по логину или authmodel.ErrNotFound.
func (repository *AuthUserRepository) FindByLogin(ctx context.Context, login string) (authmodel.User, error) {
	queryCtx, cancel := util.WithQueryTimeout(ctx)
	defer cancel()

	var user authmodel.User
	var hash []byte
	if err := repository.db.QueryRowContext(
		queryCtx,
		`SELECT id, login, password_hash FROM users WHERE login = $1`,
		login,
	).Scan(&user.ID, &user.Login, &hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return authmodel.User{}, authmodel.ErrNotFound
		}
		return authmodel.User{}, fmt.Errorf("select user: %w", err)
	}
	user.PasswordHash = hash
	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

var _ authrepo.UserRepository = (*AuthUserRepository)(nil)
