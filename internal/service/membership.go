package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xMain/subscription-hub/internal/domain"
)

type (
	AddMemberParams struct {
		UserID   int64
		TenantID int64
		Role     domain.UserRole
	}

	UpdateMemberRoleParams struct {
		UserID   int64
		TenantID int64
		NewRole  domain.UserRole
	}
)

type (
	memberProvider interface {
		ByIDAndTenant(ctx context.Context, id, tenantID int64) (*domain.User, error)
		ListByTenant(ctx context.Context, tenantID int64, limit, offset int) ([]domain.User, error)
	}

	membershipStore interface {
		Get(ctx context.Context, userID, tenantID int64) (*domain.UserTenant, error)

		Create(ctx context.Context, ut *domain.UserTenant) (*domain.UserTenant, error)
		UpdateRole(ctx context.Context, userID, tenantID int64, newRole domain.UserRole) (*domain.UserTenant, error)
		Delete(ctx context.Context, userID, tenantID int64) error

		CountByTenantID(ctx context.Context, tenantID int64) (int64, error)
		TransferOwnership(ctx context.Context, tenantID, oldOwnerID, newOwnerID int64) error
	}
)

type MemberService struct {
	members membershipStore
	users   memberProvider
	tx      transactor
}

func NewMemberService(members membershipStore, users memberProvider, tx transactor) *MemberService {
	return &MemberService{
		members: members, users: users, tx: tx,
	}
}

func (s *MemberService) GetInTenant(ctx context.Context, userID, tenantID int64) (*domain.User, error) {
	return s.users.ByIDAndTenant(ctx, userID, tenantID)
}

func (s *MemberService) ListByTenant(ctx context.Context, tenantID int64, limit, offset int) ([]domain.User, int64, error) {
	total, err := s.members.CountByTenantID(ctx, tenantID)
	if err != nil || total == 0 {
		return []domain.User{}, 0, err
	}

	users, err := s.users.ListByTenant(ctx, tenantID, limit, offset)

	return users, total, err
}

func (s *MemberService) AddUserToTenant(ctx context.Context, actorID int64, p AddMemberParams) error {
	if p.Role == domain.RoleOwner {
		return domain.ErrIllegalOwnerAssignment
	}

	actor, err := s.getActor(ctx, actorID, p.TenantID)
	if err != nil {
		return err
	}

	if !s.canManageRole(actor.Role, p.Role) {
		return domain.ErrAccessDenied
	}

	_, err = s.members.Create(ctx, &domain.UserTenant{
		UserID: p.UserID, TenantID: p.TenantID, Role: p.Role,
	})

	return err
}

func (s *MemberService) UpdateUserRole(ctx context.Context, actorID int64, p UpdateMemberRoleParams) error {
	if p.NewRole == domain.RoleOwner {
		return domain.ErrIllegalOwnerAssignment
	}

	actor, target, err := s.getActorAndTarget(ctx, actorID, p.UserID, p.TenantID)
	if err != nil {
		return err
	}

	// Нельзя менять роль текущему владельцу или устанавливать роль владельца просто так
	if target.Role == domain.RoleOwner {
		return domain.ErrAccessDenied
	}

	if !s.canManageRole(actor.Role, target.Role) || !s.canManageRole(actor.Role, p.NewRole) {
		return domain.ErrAccessDenied
	}

	_, err = s.members.UpdateRole(ctx, p.UserID, p.TenantID, p.NewRole)
	return err
}

// TransferOwnership — единственно верный путь смены владельца (атомарная ротация)
func (s *MemberService) TransferOwnership(ctx context.Context, actorID, newOwnerID, tenantID int64) error {
	if actorID == newOwnerID {
		return errors.New("нельзя передать права самому себе")
	}

	actor, err := s.getActor(ctx, actorID, tenantID)
	if err != nil {
		return err
	}

	if actor.Role != domain.RoleOwner {
		return domain.ErrAccessDenied
	}

	// Транзакция гарантирует: либо оба сменили роли, либо никто
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		// 1. Понижаем старого (освобождаем уникальный индекс в БД)
		_, err := s.members.UpdateRole(txCtx, actorID, tenantID, domain.RoleAdmin)
		if err != nil {
			return fmt.Errorf("ошибка понижения старого владельца: %w", err)
		}

		// 2. Повышаем нового
		_, err = s.members.UpdateRole(txCtx, newOwnerID, tenantID, domain.RoleOwner)
		if err != nil {
			// Если тут ErrOwnerAlreadyExists — значит кто-то вклинился
			return fmt.Errorf("ошибка повышения нового владельца: %w", err)
		}

		return nil
	})
}

func (s *MemberService) RemoveFromTenant(ctx context.Context, actorID, userID, tenantID int64) error {
	actor, target, err := s.getActorAndTarget(ctx, actorID, userID, tenantID)
	if err != nil {
		return err
	}

	if target.Role == domain.RoleOwner {
		return domain.ErrCannotRemoveOwner
	}

	if !s.canManageRole(actor.Role, target.Role) {
		return domain.ErrAccessDenied
	}

	return s.members.Delete(ctx, userID, tenantID)
}

func (s *MemberService) getActor(ctx context.Context, actorID, tenantID int64) (*domain.UserTenant, error) {
	actor, err := s.members.Get(ctx, actorID, tenantID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotInTenant) {
			return nil, domain.ErrAccessDenied
		}
		return nil, err
	}
	return actor, nil
}

func (s *MemberService) getActorAndTarget(ctx context.Context, actorID, userID, tenantID int64) (*domain.UserTenant, *domain.UserTenant, error) {
	actor, err := s.getActor(ctx, actorID, tenantID)
	if err != nil {
		return nil, nil, err
	}

	target, err := s.members.Get(ctx, userID, tenantID)
	if err != nil {
		return nil, nil, err
	}

	return actor, target, nil
}

func (s *MemberService) canManageRole(actorRole, targetRole domain.UserRole) bool {
	switch actorRole {
	case domain.RoleOwner:
		return true
	case domain.RoleAdmin:
		return targetRole != domain.RoleOwner && targetRole != domain.RoleAdmin
	default:
		return false
	}
}
