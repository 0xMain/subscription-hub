package service

import (
	"context"
	"errors"

	"github.com/0xMain/subscription-hub/internal/domain"
)

var (
	ErrUserNotInTenant = errors.New("пользователь не является сотрудником компании")
	ErrAccessDenied    = errors.New("недостаточно прав")
)

type roleGetter interface {
	Get(ctx context.Context, userID, tenantID int64) (*domain.UserTenant, error)
}

type AuthzService struct {
	roles roleGetter
}

func NewAuthzService(roles roleGetter) *AuthzService {
	return &AuthzService{roles: roles}
}

func (s *AuthzService) CheckAccess(ctx context.Context, userID, tenantID int64, allowedRoles ...domain.UserRole) error {
	role1, err := s.roles.Get(ctx, userID, tenantID)
	if err != nil {
		return err
	}
	role := role1.Role
	if role == "" {
		return ErrUserNotInTenant
	}

	if len(allowedRoles) == 0 {
		return ErrAccessDenied
	}
	for _, allowed := range allowedRoles {
		if role == allowed {
			return nil
		}
	}

	return ErrAccessDenied
}
