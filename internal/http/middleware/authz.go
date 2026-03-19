package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/0xMain/subscription-hub/internal/domain"
	"github.com/0xMain/subscription-hub/internal/http/errs"
	"github.com/0xMain/subscription-hub/internal/http/res"
	"github.com/0xMain/subscription-hub/internal/pkg/ctxutil"
	"github.com/0xMain/subscription-hub/internal/service"

	"github.com/gin-gonic/gin"
)

type accessChecker interface {
	CheckAccess(ctx context.Context, userID, tenantID int64, allowedRoles ...domain.UserRole) error
}

type AuthzMiddleware struct {
	svc accessChecker
}

func NewAuthzMiddleware(svc accessChecker) *AuthzMiddleware {
	return &AuthzMiddleware{svc: svc}
}

func (m *AuthzMiddleware) RequireRoles(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, okUser := ctxutil.GetUserID(c.Request.Context())
		if !okUser {
			res.Error(c, http.StatusUnauthorized, errs.MsgUnauthorizedErr, nil)
			return
		}

		tenantID, okTenant := ctxutil.GetTenantID(c.Request.Context())
		if !okTenant {
			res.Error(c, http.StatusBadRequest, errs.MsgMissingTenantErr, nil)
			return
		}

		err := m.svc.CheckAccess(c.Request.Context(), userID, tenantID, allowedRoles...)
		if err != nil {
			m.handleAccessError(c, err)
			return
		}

		c.Next()
	}
}

func (m *AuthzMiddleware) handleAccessError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotInTenant):
		res.Error(c, http.StatusForbidden, errs.MsgUserNotInTenantErr, nil)
	case errors.Is(err, service.ErrAccessDenied):
		res.Error(c, http.StatusForbidden, errs.MsgForbiddenErr, nil)
	default:
		res.Error(c, http.StatusInternalServerError, errs.MsgInternalErr, nil)
	}
}

func (m *AuthzMiddleware) RequireOwner() gin.HandlerFunc {
	return m.RequireRoles(domain.RoleOwner)
}

func (m *AuthzMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRoles(domain.RoleOwner, domain.RoleAdmin)
}

func (m *AuthzMiddleware) RequireManager() gin.HandlerFunc {
	return m.RequireRoles(domain.RoleOwner, domain.RoleAdmin, domain.RoleManager)
}

func (m *AuthzMiddleware) RequireViewer() gin.HandlerFunc {
	return m.RequireRoles(domain.RoleOwner, domain.RoleAdmin, domain.RoleManager, domain.RoleViewer)
}
