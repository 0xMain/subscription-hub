package req

import (
	"strconv"

	"github.com/0xMain/subscription-hub/internal/pkg/pagination"
	"github.com/gin-gonic/gin"
)

func Int(c *gin.Context, key string) (int, bool) {
	val := c.Query(key)
	if val == "" {
		return 0, false
	}
	i, err := strconv.Atoi(val)
	return i, err == nil
}

func Pagination(limitPtr, offsetPtr *int) (int, int) {
	var l, o int
	if limitPtr != nil {
		l = *limitPtr
	}
	if offsetPtr != nil {
		o = *offsetPtr
	}
	return pagination.Normalize(l, o)
}
