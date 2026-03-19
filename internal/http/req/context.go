package req

import (
	"net/http"

	"github.com/0xMain/subscription-hub/internal/http/errs"
	"github.com/0xMain/subscription-hub/internal/http/res"
	"github.com/0xMain/subscription-hub/internal/pkg/ctxutil"
	"github.com/gin-gonic/gin"
)

func UserID(c *gin.Context) (int64, bool) {
	id, ok := ctxutil.GetUserID(c.Request.Context())
	if !ok {
		res.Error(c, http.StatusUnauthorized, errs.MsgUnauthorizedErr, nil)
		return 0, false
	}
	return id, true
}

func TenantID(c *gin.Context) (int64, bool) {
	id, ok := ctxutil.GetTenantID(c.Request.Context())
	if !ok {
		res.Error(c, http.StatusBadRequest, errs.MsgMissingTenantErr, nil)
		return 0, false
	}
	return id, true
}
