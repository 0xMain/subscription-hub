package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/0xMain/subscription-hub/internal/domain"
	"github.com/0xMain/subscription-hub/internal/infra/postgres"
	"github.com/0xMain/subscription-hub/internal/pkg/tx"

	"github.com/lib/pq"
)

const (
	fkUser   = "fk_user_tenants_user"
	fkTenant = "fk_user_tenants_tenant"
)

const (
	getUserRoleQuery = `
		SELECT role
		FROM user_tenants
		WHERE user_id = $1 AND tenant_id = $2
	`

	listUserMembershipsQuery = `
		SELECT 
			ut.tenant_id, 
			t.name as tenant_name, 
			ut.role,
			COUNT(*) OVER() as total
		FROM user_tenants ut
		JOIN tenants t ON ut.tenant_id = t.id
		WHERE ut.user_id = $1
		ORDER BY t.name ASC
		LIMIT $2 OFFSET $3
`
	isOwnerAnywhereQuery = `
	SELECT EXISTS (
		SELECT 1 
		FROM user_tenants 
		WHERE user_id = $1 AND role = 'owner'
	)
`
	createUserTenantQuery = `
		INSERT INTO user_tenants (user_id, tenant_id, role)
		VALUES ($1, $2, $3)
	`

	updateUserRoleQuery = `
		UPDATE user_tenants
		SET role = $1
		WHERE user_id = $2 AND tenant_id = $3
	`

	deleteUserTenantQuery = `
		DELETE FROM user_tenants
		WHERE user_id = $1 AND tenant_id = $2
	`
)

type UserTenantRepository struct {
	db *sql.DB
	tr *tx.Transactor
}

func NewUserTenantRepository(db *sql.DB, tr *tx.Transactor) *UserTenantRepository {
	return &UserTenantRepository{
		db: db,
		tr: tr,
	}
}

func (r *UserTenantRepository) Get(ctx context.Context, userID, tenantID int64) (*domain.UserTenant, error) {
	var role string

	err := r.tr.Executor(ctx, r.db).QueryRowContext(ctx, getUserRoleQuery, userID, tenantID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotInTenant
		}
		return nil, fmt.Errorf("ошибка при получении роли участника компании (userID=%d, tenantID=%d): %w", userID, tenantID, err)
	}

	userRole, err := domain.ParseUserRole(role)
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге роли участника компании (userID=%d, tenantID=%d): %w", userID, tenantID, err)
	}

	return &domain.UserTenant{
		UserID:   userID,
		TenantID: tenantID,
		Role:     userRole,
	}, nil
}

func (r *UserTenantRepository) ListUserMemberships(ctx context.Context, userID int64, limit, offset int) ([]domain.UserTenantDetail, int64, error) {
	rows, err := r.tr.Executor(ctx, r.db).QueryContext(ctx, listUserMembershipsQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []domain.UserTenantDetail
	var total int64

	for rows.Next() {
		var item domain.UserTenantDetail
		var roleStr string

		// Сканируем всё за один проход
		err := rows.Scan(&item.TenantID, &item.TenantName, &roleStr, &total)
		if err != nil {
			return nil, 0, err
		}

		item.Role, _ = domain.ParseUserRole(roleStr)
		items = append(items, item)
	}

	return items, total, nil
}

func (r *UserTenantRepository) CountByTenantID(ctx context.Context, tenantID int64) (int64, error) {
	var total int64
	err := r.tr.Executor(ctx, r.db).QueryRowContext(ctx, countUsersByTenantIDQuery, tenantID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("ошибка при подсчете участников компании (tenantID=%d): %w", tenantID, err)
	}
	return total, nil
}

func (r *UserTenantRepository) IsOwnerAnywhere(ctx context.Context, userID int64) (bool, error) {
	var exists bool

	err := r.tr.Executor(ctx, r.db).QueryRowContext(ctx, isOwnerAnywhereQuery, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка при проверке статуса владельца (userID=%d): %w", userID, err)
	}

	return exists, nil
}

func (r *UserTenantRepository) Create(ctx context.Context, ut *domain.UserTenant) (*domain.UserTenant, error) {
	_, err := r.tr.Executor(ctx, r.db).ExecContext(ctx, createUserTenantQuery,
		ut.UserID,
		ut.TenantID,
		string(ut.Role),
	)
	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case postgres.ErrCodeUniqueViolation:
				return nil, domain.ErrUserAlreadyInTenant
			case postgres.ErrCodeForeignKeyViolation:
				switch pqErr.Constraint {
				case fkUser:
					return nil, domain.ErrUserNotFound
				case fkTenant:
					return nil, domain.ErrTenantNotFound
				}
			}
		}

		return nil, fmt.Errorf("ошибка при добавлении участника в компанию (userID=%d, tenantID=%d): %w", ut.UserID, ut.TenantID, err)
	}

	return ut, nil
}

func (r *UserTenantRepository) UpdateRole(ctx context.Context, userID, tenantID int64, newRole domain.UserRole) (*domain.UserTenant, error) {
	res, err := r.tr.Executor(ctx, r.db).ExecContext(ctx, updateUserRoleQuery,
		string(newRole),
		userID,
		tenantID,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == postgres.ErrCodeUniqueViolation {
			return nil, domain.ErrOwnerAlreadyExists
		}
		return nil, fmt.Errorf("ошибка обновления роли: %w", err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return nil, domain.ErrUserNotInTenant
	}

	return &domain.UserTenant{
		UserID:   userID,
		TenantID: tenantID,
		Role:     newRole,
	}, nil
}

func (r *UserTenantRepository) Delete(ctx context.Context, userID, tenantID int64) error {
	res, err := r.tr.Executor(ctx, r.db).ExecContext(ctx, deleteUserTenantQuery, userID, tenantID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении участника из компании (userID=%d, tenantID=%d): %w", userID, tenantID, err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при подтверждении удаления участника из компании (userID=%d, tenantID=%d): %w", userID, tenantID, err)
	}
	if count == 0 {
		return domain.ErrUserNotInTenant
	}

	return nil
}
