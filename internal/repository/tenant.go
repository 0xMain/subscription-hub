package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/0xMain/subscription-hub/internal/domain"
	"github.com/0xMain/subscription-hub/internal/infra/postgres"
	"github.com/0xMain/subscription-hub/internal/pkg/pagination"
	"github.com/0xMain/subscription-hub/internal/pkg/tx"

	"github.com/lib/pq"
)

const (
	getTenantByIDQuery = `
		SELECT id, name, created_at
		FROM tenants
		WHERE id = $1
	`

	countTenantsQuery = `
		SELECT COUNT(*)
		FROM tenants
	`

	findAllTenantsQuery = `
		SELECT id, name, created_at
		FROM tenants
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	createTenantQuery = `
		INSERT INTO tenants (name, created_at)
		VALUES ($1, NOW())
		RETURNING id, created_at
	`

	updateTenantNameQuery = `
		UPDATE tenants 
		SET name = $1
		WHERE id = $2
		RETURNING id, name, created_at
	`

	deleteTenantQuery = `
		DELETE FROM tenants
		WHERE id = $1
	`
)

type TenantRepository struct {
	db *sql.DB
	tr *tx.Transactor
}

func NewTenantRepository(db *sql.DB, tr *tx.Transactor) *TenantRepository {
	return &TenantRepository{
		db: db,
		tr: tr,
	}
}

func (r *TenantRepository) GetByID(ctx context.Context, id int64) (*domain.Tenant, error) {
	tenant, err := r.scanTenant(r.tr.Executor(ctx, r.db).QueryRowContext(ctx, getTenantByIDQuery, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrTenantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении компании (id=%d) в БД: %w", id, err)
	}

	return tenant, nil
}

func (r *TenantRepository) List(ctx context.Context, limit, offset int) ([]domain.Tenant, error) {
	limit, offset = pagination.Normalize(limit, offset)

	rows, err := r.tr.Executor(ctx, r.db).QueryContext(ctx, findAllTenantsQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка компаний в БД: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("не удалось закрыть rows при получении списка компаний: %v", closeErr)
		}
	}()

	tenants := make([]domain.Tenant, 0, limit)
	for rows.Next() {
		tenant, err := r.scanTenant(rows)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании компании в БД: %w", err)
		}
		tenants = append(tenants, *tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении списка компаний в БД: %w", err)
	}

	return tenants, nil
}

func (r *TenantRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	err := r.tr.Executor(ctx, r.db).QueryRowContext(ctx, countTenantsQuery).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("ошибка при подсчете общего количества компаний: %w", err)
	}
	return total, nil
}

func (r *TenantRepository) Create(ctx context.Context, tenant *domain.Tenant) (*domain.Tenant, error) {
	err := r.tr.Executor(ctx, r.db).QueryRowContext(ctx, createTenantQuery,
		tenant.Name,
	).Scan(&tenant.ID, &tenant.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == postgres.ErrCodeUniqueViolation {
			return nil, domain.ErrTenantAlreadyExists
		}
		return nil, fmt.Errorf("ошибка при создании компании в БД: %w", err)
	}

	return tenant, nil
}

func (r *TenantRepository) UpdateName(ctx context.Context, id int64, name string) (*domain.Tenant, error) {
	var t domain.Tenant

	err := r.tr.Executor(ctx, r.db).QueryRowContext(ctx, updateTenantNameQuery,
		name,
		id,
	).Scan(&t.ID, &t.Name, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTenantNotFound
		}

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == postgres.ErrCodeUniqueViolation {
			return nil, domain.ErrTenantAlreadyExists
		}

		return nil, fmt.Errorf("ошибка при обновлении названия компании (id=%d) в БД: %w", id, err)
	}

	return &t, nil
}

func (r *TenantRepository) DeleteByID(ctx context.Context, id int64) error {
	res, err := r.tr.Executor(ctx, r.db).ExecContext(ctx, deleteTenantQuery, id)
	if err != nil {
		return fmt.Errorf("ошибка при удалении компании (id=%d) в БД: %w", id, err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при подтверждении удаления компании (id=%d): %w", id, err)
	}

	if count == 0 {
		return domain.ErrTenantNotFound
	}

	return nil
}

func (r *TenantRepository) scanTenant(row scanner) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := row.Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}
