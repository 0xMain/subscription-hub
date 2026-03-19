package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/0xMain/subscription-hub/internal/http/errs"
	"github.com/0xMain/subscription-hub/internal/http/res"
	"github.com/0xMain/subscription-hub/internal/pkg/ctxutil"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type tokenValidator interface {
	ValidateToken(token string) (jwt.MapClaims, error)
}

type AuthnMiddleware struct {
	svc tokenValidator
}

func NewAuthnMiddleware(svc tokenValidator) *AuthnMiddleware {
	return &AuthnMiddleware{svc: svc}
}

func (m *AuthnMiddleware) Verify() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			res.Error(c, http.StatusUnauthorized, errs.MsgUnauthorizedErr, nil)
			return
		}

		claims, err := m.svc.ValidateToken(tokenString)
		if err != nil {
			res.Error(c, http.StatusUnauthorized, errs.MsgInvalidTokenErr, nil)
			return
		}

		sub, err := claims.GetSubject()
		if err != nil || sub == "" {
			res.Error(c, http.StatusUnauthorized, errs.MsgInvalidUserIDErr, nil)
			return
		}

		userID, err := strconv.ParseInt(sub, 10, 64)
		if err != nil {
			res.Error(c, http.StatusUnauthorized, errs.MsgInvalidUserIDErr, nil)
			return
		}

		ctxutil.SetUserID(c, userID)

		if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
			if tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64); err == nil && tenantID > 0 {
				ctxutil.SetTenantID(c, tenantID)
			}
		}

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}
