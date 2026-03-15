package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/0xMain/subscription-hub/internal/domain"
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
		userIDVal, exists := c.Get(string(ctxutil.ContextUserID))
		userID, ok := userIDVal.(int64)
		if !exists || !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "необходима повторная авторизация"})
			return
		}

		tenantIDVal, exists := c.Get(string(ctxutil.ContextTenantID))
		tenantID, ok := tenantIDVal.(int64)
		if !exists || !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "идентификатор компании (X-Tenant-ID) отсутствует или некорректен"})
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
	case errors.Is(err, service.ErrUserNotInTenant), errors.Is(err, service.ErrAccessDenied):
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "ошибка проверки прав доступа"})
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
