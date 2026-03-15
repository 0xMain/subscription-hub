package service

import (
	"context"

	"github.com/0xMain/subscription-hub/internal/domain"
)

type (
	tenantStore interface {
		GetByID(ctx context.Context, id int64) (*domain.Tenant, error)
		List(ctx context.Context, limit, offset int) ([]domain.Tenant, error)
		Count(ctx context.Context) (int64, error)

		Create(ctx context.Context, tenant *domain.Tenant) (*domain.Tenant, error)
		UpdateName(ctx context.Context, id int64, name string) (*domain.Tenant, error)
		DeleteByID(ctx context.Context, id int64) error
	}

	membershipBase interface {
		Get(ctx context.Context, userID, tenantID int64) (*domain.UserTenant, error)
		Create(ctx context.Context, ut *domain.UserTenant) (*domain.UserTenant, error)
	}
)

type TenantService struct {
	tenants tenantStore
	members membershipBase
	tx      transactor
}

func NewTenantService(tenants tenantStore, members membershipBase, tx transactor) *TenantService {
	return &TenantService{
		tenants: tenants, members: members, tx: tx,
	}
}

func (s *TenantService) GetByID(ctx context.Context, id int64) (*domain.Tenant, error) {
	return s.tenants.GetByID(ctx, id)
}

func (s *TenantService) List(ctx context.Context, limit, offset int) ([]domain.Tenant, int64, error) {
	total, err := s.tenants.Count(ctx)
	if err != nil || total == 0 {
		return []domain.Tenant{}, 0, err
	}

	tenants, err := s.tenants.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}

func (s *TenantService) Create(ctx context.Context, actorID int64, name string) (*domain.Tenant, error) {
	var created *domain.Tenant

	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error

		created, err = s.tenants.Create(txCtx, &domain.Tenant{Name: name})
		if err != nil {
			return err
		}

		_, err = s.members.Create(txCtx, &domain.UserTenant{
			UserID:   actorID,
			TenantID: created.ID,
			Role:     domain.RoleOwner,
		})

		return err
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *TenantService) UpdateName(ctx context.Context, tenantID int64, newName string) (*domain.Tenant, error) {
	return s.tenants.UpdateName(ctx, tenantID, newName)
}

func (s *TenantService) Delete(ctx context.Context, id int64) error {
	return s.tenants.DeleteByID(ctx, id)
}
