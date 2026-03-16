package ctxutil

import (
	"context"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	ContextUserID   contextKey = "user_id"
	ContextTenantID contextKey = "tenant_id"
)

func GetUserID(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(ContextUserID).(int64)
	return v, ok
}

func SetUserID(c *gin.Context, userID int64) {
	c.Set(string(ContextUserID), userID)
	ctx := context.WithValue(c.Request.Context(), ContextUserID, userID)
	c.Request = c.Request.WithContext(ctx)
}

func GetTenantID(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(ContextTenantID).(int64)
	return v, ok
}

func SetTenantID(c *gin.Context, tenantID int64) {
	c.Set(string(ContextTenantID), tenantID)
	ctx := context.WithValue(c.Request.Context(), ContextTenantID, tenantID)
	c.Request = c.Request.WithContext(ctx)
}
