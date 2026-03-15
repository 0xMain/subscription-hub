package ctxutil

import "context"

type contextKey string

const (
	ContextUserID   contextKey = "user_id"
	ContextTenantID contextKey = "tenant_id"
)

func GetUserID(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(ContextUserID).(int64)
	return v, ok
}

func GetTenantID(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(ContextTenantID).(int64)
	return v, ok
}
