package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "требуется авторизация"})
			return
		}

		claims, err := m.svc.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "невалидный токен"})
			return
		}

		sub, err := claims.GetSubject()
		if err != nil || sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "некорректный идентификатор пользователя"})
			return
		}

		userID, err := strconv.ParseInt(sub, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "ошибка идентификации"})
			return
		}

		c.Set(string(ctxutil.ContextUserID), userID)
		ctx := context.WithValue(c.Request.Context(), ctxutil.ContextUserID, userID)

		if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
			if tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64); err == nil && tenantID > 0 {
				c.Set(string(ctxutil.ContextTenantID), tenantID)
				ctx = context.WithValue(ctx, ctxutil.ContextTenantID, tenantID)
			}
		}

		c.Request = c.Request.WithContext(ctx)
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
