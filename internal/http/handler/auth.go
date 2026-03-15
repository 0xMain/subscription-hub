package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/0xMain/subscription-hub/internal/domain"
	"github.com/0xMain/subscription-hub/internal/http/gen/authapi"
	"github.com/0xMain/subscription-hub/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/oapi-codegen/runtime/types"
)

type authService interface {
	SignUp(ctx context.Context, p service.SignUpParams) (*domain.User, error)
	SignIn(ctx context.Context, email, password string) (*service.SignInResult, error)
}

type AuthHandler struct {
	baseHelper
	svc      authService
	validate *validator.Validate
}

func NewAuthHandler(svc authService, validate *validator.Validate) *AuthHandler {
	return &AuthHandler{svc: svc, validate: validate}
}

func (h *AuthHandler) SignUp(c *gin.Context) {
	var req authapi.SignUpRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, msgInvalidFormat, nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.sendError(c, http.StatusBadRequest, msgValidationErr, h.formatValidationErrors(err))
		return
	}

	user, err := h.svc.SignUp(c.Request.Context(), service.SignUpParams{
		Email:     string(req.Email),
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyRegistered) {
			h.sendError(c, http.StatusConflict, msgUserExists, nil)
			return
		}

		log.Printf("внутренняя ошибка (метод=SignUp): %v", err)
		h.sendError(c, http.StatusInternalServerError, msgInternalError, nil)
		return
	}

	c.JSON(http.StatusCreated, h.mapUserToResponse(user))
}

func (h *AuthHandler) SignIn(c *gin.Context) {
	var req authapi.SignInRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, msgInvalidFormat, nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.sendError(c, http.StatusBadRequest, msgValidationErr, h.formatValidationErrors(err))
		return
	}

	res, err := h.svc.SignIn(c.Request.Context(), string(req.Email), req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			h.sendError(c, http.StatusUnauthorized, msgInvalidCredentials, nil)
			return
		}

		log.Printf("внутренняя ошибка (метод=SignIn): %v", err)
		h.sendError(c, http.StatusInternalServerError, msgInternalError, nil)
		return
	}

	c.JSON(http.StatusOK, authapi.SignInResponse{
		AccessToken: res.AccessToken,
		User:        h.mapUserToResponse(res.User),
	})
}

func (h *AuthHandler) mapUserToResponse(user *domain.User) authapi.UserResponse {
	return authapi.UserResponse{
		ID:        user.ID,
		Email:     types.Email(user.Email),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}
}
