package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/0xMain/subscription-hub/internal/domain"
	"github.com/0xMain/subscription-hub/internal/http/errs"
	"github.com/0xMain/subscription-hub/internal/http/gen/authapi"
	usergen "github.com/0xMain/subscription-hub/internal/http/gen/user"
	"github.com/0xMain/subscription-hub/internal/http/req"
	"github.com/0xMain/subscription-hub/internal/http/res"
	"github.com/0xMain/subscription-hub/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime/types"
)

type authService interface {
	SignUp(ctx context.Context, p service.SignUpParams) (*domain.User, error)
	SignIn(ctx context.Context, email, password string) (*service.SignInResult, error)
}

type AuthHandler struct {
	svc authService
}

func NewAuthHandler(svc authService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) SignUp(c *gin.Context) {
	var r authapi.SignUpRequest
	if !req.Body(c, &r) {
		return
	}

	user, err := h.svc.SignUp(c.Request.Context(), service.SignUpParams{
		Email:     string(r.Email),
		Password:  r.Password,
		FirstName: r.FirstName,
		LastName:  r.LastName,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyRegistered) {
			res.Error(c, http.StatusConflict, errs.MsgUserExistsErr, nil)
			return
		}

		log.Printf("внутренняя ошибка (метод=SignUp): %v", err)
		res.Error(c, http.StatusInternalServerError, errs.MsgInternalErr, nil)
		return
	}

	res.Created(c, h.toUserResponse(user))
}

func (h *AuthHandler) SignIn(c *gin.Context) {
	var r authapi.SignInRequest
	if !req.Body(c, &r) {
		return
	}

	result, err := h.svc.SignIn(c.Request.Context(), string(r.Email), r.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			res.Error(c, http.StatusUnauthorized, errs.MsgInvalidCredentialsErr, nil)
			return
		}

		log.Printf("внутренняя ошибка (метод=SignIn): %v", err)
		res.Error(c, http.StatusInternalServerError, errs.MsgInternalErr, nil)
		return
	}

	res.OK(c, authapi.SignInResponse{
		AccessToken: result.AccessToken,
		User:        h.toUserResponse(result.User),
	})
}

func (h *AuthHandler) toUserResponse(user *domain.User) usergen.UserResponse {
	return usergen.UserResponse{
		ID:        user.ID,
		Email:     types.Email(user.Email),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}
}
