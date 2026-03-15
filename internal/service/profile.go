package service

import (
	"context"

	"github.com/0xMain/subscription-hub/internal/domain"
	"github.com/0xMain/subscription-hub/internal/pkg/strutil"
)

type FullProfileResult struct {
	User             *domain.User
	Memberships      []domain.UserTenantDetail
	MembershipsTotal int64
}

type (
	profileStore interface {
		ByID(ctx context.Context, id int64) (*domain.User, error)
		Update(ctx context.Context, id int64, firstName, lastName *string) (*domain.User, error)
		Delete(ctx context.Context, id int64) error
	}
	profileLinker interface {
		IsOwnerAnywhere(ctx context.Context, userID int64) (bool, error)
		ListUserMemberships(ctx context.Context, userID int64, limit, offset int) ([]domain.UserTenantDetail, int64, error)
	}
)

type ProfileService struct {
	store profileStore
	links profileLinker
}

func NewProfileService(store profileStore, links profileLinker) *ProfileService {
	return &ProfileService{
		store: store,
		links: links,
	}
}

func (s *ProfileService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.store.ByID(ctx, id)
}

func (s *ProfileService) GetFullProfile(ctx context.Context, userID int64, limit, offset int) (*FullProfileResult, error) {
	user, err := s.store.ByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	items, total, err := s.links.ListUserMemberships(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &FullProfileResult{
		User:             user,
		Memberships:      items,
		MembershipsTotal: total,
	}, nil
}

func (s *ProfileService) Update(ctx context.Context, userID int64, firstName, lastName *string) (*domain.User, error) {
	return s.store.Update(ctx, userID, strutil.CapitalizePtr(firstName), strutil.CapitalizePtr(lastName))
}

func (s *ProfileService) Delete(ctx context.Context, userID int64) error {
	isOwner, err := s.links.IsOwnerAnywhere(ctx, userID)
	if err != nil {
		return err
	}

	if isOwner {
		return domain.ErrCannotDeleteOwner
	}

	return s.store.Delete(ctx, userID)
}
