package httputil

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/0xMain/subscription-hub/internal/http/gen/profileapi"
	"github.com/0xMain/subscription-hub/internal/http/httperrs"
	"github.com/0xMain/subscription-hub/internal/pkg/ctxutil"
	"github.com/0xMain/subscription-hub/internal/pkg/pagination"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BaseHelper struct{}

func (h *BaseHelper) RequireUserID(c *gin.Context) (int64, bool) {
	userID, ok := ctxutil.GetUserID(c.Request.Context())
	if !ok {
		h.SendError(c, http.StatusUnauthorized, httperrs.MsgUnauthorizedErr, nil)
		return 0, false
	}
	return userID, true
}

func (h *BaseHelper) RequireTenantID(c *gin.Context) (int64, bool) {
	tenantID, ok := ctxutil.GetTenantID(c.Request.Context())
	if !ok {
		h.SendError(c, http.StatusBadRequest, httperrs.MsgMissingTenantErr, nil)
		return 0, false
	}
	return tenantID, true
}

func (h *BaseHelper) GetPaginationParams(limitPtr, offsetPtr *int) (int, int) {
	var l, o int
	if limitPtr != nil {
		l = *limitPtr
	}
	if offsetPtr != nil {
		o = *offsetPtr
	}

	return pagination.Normalize(l, o)
}

func (h *BaseHelper) SendError(c *gin.Context, status int, msg string, details map[string]string) {
	resp := profileapi.ErrorResponse{
		Error: msg,
	}
	if len(details) > 0 {
		resp.Details = &details
	}
	c.AbortWithStatusJSON(status, resp)
}

func (h *BaseHelper) FormatValidationErrors(err error) map[string]string {
	var ve validator.ValidationErrors
	details := make(map[string]string)
	if errors.As(err, &ve) {
		for _, fe := range ve {
			details[fe.Field()] = fmt.Sprintf("некорректное значение: %s", fe.Tag())
		}
	}
	return details
}

func (h *BaseHelper) GetQueryInt(c *gin.Context, key string) (int, bool) {
	val := c.Query(key)
	if val == "" {
		return 0, false
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}
	return i, true
}
