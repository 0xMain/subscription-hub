package req

import (
	"github.com/0xMain/subscription-hub/internal/http/res"
	"github.com/gin-gonic/gin"
)

func Body(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		res.Validation(c, err)
		return false
	}
	return true
}
