package res

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/0xMain/subscription-hub/internal/http/errs"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func Validation(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	details := make(map[string]string)
	if errors.As(err, &ve) {
		for _, fe := range ve {
			details[fe.Field()] = fmt.Sprintf("некорректное значение: %s", fe.Tag())
		}
	}
	Error(c, http.StatusUnprocessableEntity, errs.MsgValidationErr, nil)
}
