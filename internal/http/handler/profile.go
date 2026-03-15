package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/0xMain/subscription-hub/internal/domain"
	"github.com/0xMain/subscription-hub/internal/http/gen/profileapi"
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
	baseHelper
	svc      profileService
	validate *validator.Validate
}

func NewProfileHandler(svc profileService, validate *validator.Validate) *ProfileHandler {
	return &ProfileHandler{svc: svc, validate: validate}
}

func (h *ProfileHandler) GetCurrentProfile(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			h.sendError(c, http.StatusUnauthorized, msgUnauthorized, nil)
			return
		}
		log.Printf("внутренняя ошибка (метод=GetCurrentProfile, ID=%d): %v", userID, err)
		h.sendError(c, http.StatusInternalServerError, msgInternalError, nil)
		return
	}

	c.JSON(http.StatusOK, h.mapUserToResponse(user))
}

func (h *ProfileHandler) GetFullCurrentProfile(c *gin.Context, params profileapi.GetFullCurrentProfileParams) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	limit, offset := h.getPaginationParams(params.TenantsLimit, params.TenantsOffset)

	full, err := h.svc.GetFullProfile(c.Request.Context(), userID, limit, offset)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			h.sendError(c, http.StatusUnauthorized, msgUnauthorized, nil)
			return
		}
		log.Printf("внутренняя ошибка (метод=GetFullCurrentProfile, ID=%d): %v", userID, err)
		h.sendError(c, http.StatusInternalServerError, msgInternalError, nil)
		return
	}

	c.JSON(http.StatusOK, h.mapFullProfileToResponse(full, limit, offset))
}

func (h *ProfileHandler) UpdateCurrentProfile(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	var req profileapi.UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, msgInvalidFormat, nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.sendError(c, http.StatusBadRequest, msgValidationErr, h.formatValidationErrors(err))
		return
	}

	user, err := h.svc.Update(c.Request.Context(), userID, req.FirstName, req.LastName)
	if err != nil {
		log.Printf("внутренняя ошибка (метод=UpdateCurrentProfile, ID=%d): %v", userID, err)
		h.sendError(c, http.StatusInternalServerError, msgInternalError, nil)
		return
	}

	c.JSON(http.StatusOK, h.mapUserToResponse(user))
}

func (h *ProfileHandler) DeleteCurrentProfile(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	err := h.svc.Delete(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrCannotDeleteOwner) {
			h.sendError(c,
				http.StatusUnprocessableEntity,
				msgDeleteErr,
				map[string]string{"base": msgCannotDeleteOwner},
			)
			return
		}
		log.Printf("внутренняя ошибка (метод=DeleteCurrentProfile, ID=%d): %v", userID, err)
		h.sendError(c, http.StatusInternalServerError, msgInternalError, nil)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProfileHandler) mapUserToResponse(user *domain.User) profileapi.UserResponse {
	return profileapi.UserResponse{
		ID:        user.ID,
		Email:     types.Email(user.Email),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}
}

func (h *ProfileHandler) mapFullProfileToResponse(p *service.FullProfileResult, limit, offset int) profileapi.FullProfileResponse {
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
			Pagination profileapi.PaginationMeta         `json:"pagination"`
		}{
			Items: tenantItems,
			Pagination: profileapi.PaginationMeta{
				Limit:  limit,
				Offset: offset,
				Total:  p.MembershipsTotal,
			},
		},
	}
}
