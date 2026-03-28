package res

import (
	"github.com/0xMain/subscription-hub/internal/http/gen/common"
	"github.com/gin-gonic/gin"
)

func Error(c *gin.Context, status int, msg string, details map[string][]string) {
	resp := common.ErrorResponse{Error: msg}
	if len(details) > 0 {
		resp.Details = &details
	}
	c.AbortWithStatusJSON(status, resp)
}
