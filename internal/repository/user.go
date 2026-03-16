package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/0xMain/subscription-hub/internal/domain"
	"github.com/0xMain/subscription-hub/internal/infra/postgres"
	"github.com/0xMain/subscription-hub/internal/pkg/pagination"
	"github.com/0xMain/subscription-hub/internal/pkg/tx"
	"github.com/0xMain/subscription-hub/internal/repository/dao"

	"github.com/lib/pq"
)

const (
	getUserByIDQuery = `
		SELECT id, email, first_name, last_name, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	getUserByEmailQuery = `
		SELECT id, email, first_name, last_name, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	getUserByIDAndTenantIDQuery = `
		SELECT u.id, u.email, u.first_name, u.last_name, u.password_hash, u.created_at
		FROM users u
		JOIN user_tenants ut ON u.id = ut.user_id
		WHERE u.id = $1 AND ut.tenant_id = $2
	`

	listUsersByTenantIDQuery = `
		SELECT u.id, u.email, u.first_name, u.last_name, u.password_hash, u.created_at
		FROM users u
		JOIN user_tenants ut ON u.id = ut.user_id
		WHERE ut.tenant_id = $1
		ORDER BY u.id DESC
		LIMIT $2 OFFSET $3
	`

	createUserQuery = `
		INSERT INTO users (email, first_name, last_name, password_hash, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`

	updateUserQuery = `
		UPDATE users 
		SET 
			first_name = COALESCE($1, first_name),
			last_name  = COALESCE($2, last_name)
		WHERE id = $3
		RETURNING id, email, first_name, last_name, password_hash, created_at
	`

	deleteUserByIDQuery = `
		DELETE FROM users
		WHERE id = $1
	`
)

type UserRepository struct {
	db *sql.DB
	tr *tx.Transactor
}

func NewUserRepository(db *sql.DB, tr *tx.Transactor) *UserRepository {
	return &UserRepository{
		db: db,
		tr: tr,
	}
}

func (r *UserRepository) ByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := r.scan(r.tr.Executor(ctx, r.db).QueryRowContext(ctx, getUserByIDQuery, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя (id=%d) в БД: %w", id, err)
	}

	return user.ToDomain(), nil
}

func (r *UserRepository) ByEmail(ctx context.Context, email string) (*domain.UserWithPassword, error) {
	user, err := r.scan(r.tr.Executor(ctx, r.db).QueryRowContext(ctx, getUserByEmailQuery, email))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя по email в БД: %w", err)
	}

	return &domain.UserWithPassword{User: user.ToDomain(), PasswordHash: user.PasswordHash}, nil
}

func (r *UserRepository) ByIDAndTenant(ctx context.Context, userID, tenantID int64) (*domain.User, error) {
	user, err := r.scan(r.tr.Executor(ctx, r.db).QueryRowContext(ctx, getUserByIDAndTenantIDQuery, userID, tenantID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя (userID=%d) для компании (tenantID=%d): %w", userID, tenantID, err)
	}

	return user.ToDomain(), nil
}

func (r *UserRepository) ListByTenant(ctx context.Context, tenantID int64, limit, offset int) ([]domain.User, int64, error) {
	limit, offset = pagination.Normalize(limit, offset)

	executor := r.tr.Executor(ctx, r.db)

	var total int64
	err := executor.QueryRowContext(ctx, countByTenantIDQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка при подсчете участников компании (tenantID=%d): %w", tenantID, err)
	}

	if total == 0 || int64(offset) >= total {
		return []domain.User{}, total, nil
	}

	rows, err := executor.QueryContext(ctx, listUsersByTenantIDQuery, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка при получении списка участников компании (tenantID=%d) в БД: %w", tenantID, err)
	}

	defer func() { _ = rows.Close() }()

	users := make([]domain.User, 0, pagination.CalculateCapacity(total, limit, offset))

	for rows.Next() {
		user, err := r.scan(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("ошибка при сканировании участника компании в БД: %w", err)
		}
		users = append(users, *user.ToDomain())
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ошибка при чтении списка участников компании (tenantID=%d) в БД: %w", tenantID, err)
	}

	return users, total, nil
}

func (r *UserRepository) Create(ctx context.Context, user *dao.User) (*domain.User, error) {
	err := r.tr.Executor(ctx, r.db).QueryRowContext(ctx, createUserQuery,
		user.Email,
		user.FirstName,
		user.LastName,
		user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == postgres.ErrCodeUniqueViolation {
			return nil, domain.ErrUserAlreadyRegistered
		}
		return nil, fmt.Errorf("ошибка при создании пользователя в БД: %w", err)
	}

	return user.ToDomain(), nil
}

func (r *UserRepository) Update(ctx context.Context, id int64, firstName, lastName *string) (*domain.User, error) {
	if firstName == nil && lastName == nil {
		return r.ByID(ctx, id)
	}

	user, err := r.scan(r.tr.Executor(ctx, r.db).QueryRowContext(ctx, updateUserQuery, firstName, lastName, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("ошибка при обновлении пользователя (id=%d) в БД: %w", id, err)
	}

	return user.ToDomain(), nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.tr.Executor(ctx, r.db).ExecContext(ctx, deleteUserByIDQuery, id)
	if err != nil {
		return fmt.Errorf("ошибка при удалении пользователя (id=%d) в БД: %w", id, err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при подтверждении удаления пользователя (id=%d): %w", id, err)
	}

	if count == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) scan(row scanner) (*dao.User, error) {
	var user dao.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
