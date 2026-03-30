package req

import (
	"net/http"

	"github.com/0xMain/subscription-hub/internal/http/errs"
	"github.com/0xMain/subscription-hub/internal/http/res"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func Body(c *gin.Context, obj any) bool {
	if err := c.ShouldBindWith(obj, binding.JSON); err != nil {
		res.Error(c, http.StatusBadRequest, errs.MsgInvalidFormatErr, nil)

		return false
	}

	return true
}
