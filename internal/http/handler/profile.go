package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/0xMain/subscription-hub/internal/domain"
	"github.com/0xMain/subscription-hub/internal/http/errs"
	commongen "github.com/0xMain/subscription-hub/internal/http/gen/common"
	"github.com/0xMain/subscription-hub/internal/http/gen/profileapi"
	usergen "github.com/0xMain/subscription-hub/internal/http/gen/user"
	"github.com/0xMain/subscription-hub/internal/http/req"
	"github.com/0xMain/subscription-hub/internal/http/res"
	"github.com/0xMain/subscription-hub/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/oapi-codegen/runtime/types"
)

type profileService interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetFullProfile(ctx context.Context, userID int64, limit, offset int) (*service.FullProfileResult, error)

	Update(ctx context.Context, userID int64, firstName, lastName *string) (*domain.User, error)
	Delete(ctx context.Context, userID int64) error
}

type ProfileHandler struct {
	svc      profileService
	validate *validator.Validate
}

func NewProfileHandler(svc profileService, validate *validator.Validate) *ProfileHandler {
	return &ProfileHandler{svc: svc, validate: validate}
}

func (h *ProfileHandler) GetCurrentProfile(c *gin.Context) {
	uid, ok := req.UserID(c)
	if !ok {
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			res.Error(c, http.StatusUnauthorized, errs.MsgUnauthorizedErr, nil)
			return
		}
		log.Printf("внутренняя ошибка (метод=GetCurrentProfile, ID=%d): %v", uid, err)
		res.Error(c, http.StatusInternalServerError, errs.MsgInternalErr, nil)
		return
	}

	res.OK(c, h.toUserResponse(user))
}

func (h *ProfileHandler) GetFullCurrentProfile(c *gin.Context, params profileapi.GetFullCurrentProfileParams) {
	uid, ok := req.UserID(c)
	if !ok {
		return
	}

	limit, offset := req.Pagination(params.TenantsLimit, params.TenantsOffset)

	full, err := h.svc.GetFullProfile(c.Request.Context(), uid, limit, offset)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			res.Error(c, http.StatusUnauthorized, errs.MsgUnauthorizedErr, nil)
			return
		}
		log.Printf("внутренняя ошибка (метод=GetFullCurrentProfile, ID=%d): %v", uid, err)
		res.Error(c, http.StatusInternalServerError, errs.MsgInternalErr, nil)
		return
	}

	res.OK(c, h.toFullProfileResponse(full, limit, offset))
}

func (h *ProfileHandler) UpdateCurrentProfile(c *gin.Context) {
	uid, ok := req.UserID(c)
	if !ok {
		return
	}

	var r profileapi.UpdateProfileRequest
	if !req.Body(c, &r) {
		return
	}

	user, err := h.svc.Update(c.Request.Context(), uid, r.FirstName, r.LastName)
	if err != nil {
		log.Printf("внутренняя ошибка (метод=UpdateCurrentProfile, ID=%d): %v", uid, err)
		res.Error(c, http.StatusInternalServerError, errs.MsgInternalErr, nil)
		return
	}

	res.OK(c, h.toUserResponse(user))
}

func (h *ProfileHandler) DeleteCurrentProfile(c *gin.Context) {
	uid, ok := req.UserID(c)
	if !ok {
		return
	}

	err := h.svc.Delete(c.Request.Context(), uid)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCannotDeleteOwner):
			res.Error(c,
				http.StatusConflict,
				errs.MsgDeleteErr,
				map[string][]string{"base": {errs.MsgCannotDeleteOwnerErr}},
			)
		case errors.Is(err, domain.ErrUserNotFound):
			res.Error(c, http.StatusNotFound, errs.MsgUserNotFoundErr, nil)
		default:
			log.Printf("внутренняя ошибка (метод=DeleteCurrentProfile, ID=%d): %v", uid, err)
			res.Error(c, http.StatusInternalServerError, errs.MsgInternalErr, nil)
		}
		return
	}

	res.NoContent(c)
}

func (h *ProfileHandler) toUserResponse(user *domain.User) usergen.UserResponse {
	return usergen.UserResponse{
		ID:        user.ID,
		Email:     types.Email(user.Email),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}
}

func (h *ProfileHandler) toFullProfileResponse(p *service.FullProfileResult, limit, offset int) profileapi.FullProfileResponse {
	tenantItems := make([]profileapi.MemberTenantResponse, len(p.Memberships))
	for i, m := range p.Memberships {
		tenantItems[i] = profileapi.MemberTenantResponse{
			TenantID:   m.TenantID,
			TenantName: m.TenantName,
			Role:       profileapi.MemberTenantResponseRole(m.Role),
		}
	}

	return profileapi.FullProfileResponse{
		ID:        p.User.ID,
		Email:     types.Email(p.User.Email),
		FirstName: p.User.FirstName,
		LastName:  p.User.LastName,
		CreatedAt: p.User.CreatedAt,

		Tenants: struct {
			Items      []profileapi.MemberTenantResponse `json:"items"`
			Pagination commongen.PaginationMeta          `json:"pagination"`
		}{
			Items: tenantItems,
			Pagination: commongen.PaginationMeta{
				Limit:  limit,
				Offset: offset,
				Total:  p.MembershipsTotal,
			},
		},
	}
}
