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

type membershipFinder interface {
	Get(ctx context.Context, userID, tenantID int64) (*domain.UserTenant, error)
}

type AuthzService struct {
	ms membershipFinder
}

func NewAuthzService(ms membershipFinder) *AuthzService {
	return &AuthzService{
		ms: ms,
	}
}

func (s *AuthzService) CheckAccess(ctx context.Context, userID, tenantID int64, allowedRoles ...domain.UserRole) error {
	m, err := s.ms.Get(ctx, userID, tenantID)
	if err != nil {
		return err
	}

	if len(allowedRoles) == 0 {
		return ErrAccessDenied
	}

	for _, allowed := range allowedRoles {
		if m.Role == allowed {
			return nil
		}
	}

	return ErrAccessDenied
}
