package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/0xMain/subscription-hub/internal/http/errs"
	"github.com/0xMain/subscription-hub/internal/http/res"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	oapiMiddleware "github.com/oapi-codegen/gin-middleware"
)

var (
	reNumbers     = regexp.MustCompile(`\d+`)
	reLimitType   = regexp.MustCompile(`(?i)(min|least|short|max|most|long)`)
	reErrorTarget = regexp.MustCompile(`"([^"]+)"`)
)

func OAPIValidator(swagger *openapi3.T) gin.HandlerFunc {
	swagger.Servers = nil

	return oapiMiddleware.OapiRequestValidatorWithOptions(swagger, &oapiMiddleware.Options{
		Options: openapi3filter.Options{
			MultiError:         true,
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
		ErrorHandler: func(c *gin.Context, message string, statusCode int) {

			fmt.Println(message)
			details := make(map[string]string)

			for _, ginErr := range c.Errors {
				var multiErr openapi3.MultiError

				if errors.As(ginErr.Err, &multiErr) {
					for _, err := range multiErr {
						addError(err, details)
					}
				}
			}

			if len(details) == 0 {
				details = extractErrors(message)
			}

			if len(details) == 0 {
				res.Error(c, http.StatusBadRequest, errs.MsgInvalidFormatErr, nil)
				return
			}

			res.Error(c, http.StatusUnprocessableEntity, errs.MsgValidationErr, details)
		},
	})
}

func formatSchemaError(reason string, schema *openapi3.Schema) string {
	msg := strings.ToLower(reason)

	if out, ok := checkRequired(msg); ok {
		return out
	}

	if out, ok := checkLimit(msg, schema); ok {
		return out
	}

	if out, ok := checkType(msg); ok {
		return out
	}

	if out, ok := checkFormat(msg); ok {
		return out
	}

	if out, ok := checkEnum(msg, schema); ok {
		return out
	}

	return "некорректные данные"
}

func checkRequired(msg string) (string, bool) {
	if strings.Contains(msg, "missing") || strings.Contains(msg, "required") {
		return "обязательное поле отсутствует", true
	}

	return "", false
}

func checkLimit(msg string, schema *openapi3.Schema) (string, bool) {
	limitMatch := strings.ToLower(reLimitType.FindString(msg))

	if limitMatch == "" && !strings.Contains(strings.ToLower(msg), "length") {
		return "", false
	}

	isMax := limitMatch == "max" || limitMatch == "most" || limitMatch == "long"
	isMin := !isMax

	plural := func(n int) string {
		v := n % 100
		if v >= 11 && v <= 19 {
			return "символов"
		}
		switch v % 10 {
		case 1:
			return "символ"
		case 2, 3, 4:
			return "символа"
		default:
			return "символов"
		}
	}

	limitVal := ""
	if schema != nil {
		switch {
		case isMin && schema.Min != nil:
			limitVal = fmt.Sprintf("%g", *schema.Min)
		case isMin && schema.MinLength > 0:
			limitVal = strconv.Itoa(int(schema.MinLength))
		case isMax && schema.Max != nil:
			limitVal = fmt.Sprintf("%g", *schema.Max)
		case isMax && schema.MaxLength != nil:
			limitVal = strconv.Itoa(int(*schema.MaxLength))
		}
	}

	if limitVal == "" {
		limitVal = reNumbers.FindString(msg)
	}

	if limitVal != "" {
		prefix := "минимум "
		if isMax {
			prefix = "максимум "
		}

		isString := schema != nil && schema.Type.Is("string")
		if !isString {
			isString = strings.Contains(msg, "string") || strings.Contains(msg, "length")
		}

		if isString {
			n, _ := strconv.Atoi(limitVal)
			return prefix + limitVal + " " + plural(n), true
		}

		return prefix + limitVal, true
	}

	if isMin {
		return "слишком короткое значение", true
	}

	return "слишком длинное значение", true
}

func checkType(msg string) (string, bool) {
	if strings.Contains(msg, "null") {
		return "поле не может быть пустым", true
	}

	if !strings.Contains(msg, "must be") && !strings.Contains(msg, "expected") && !strings.Contains(msg, "type") {
		return "", false
	}

	switch {
	case strings.Contains(msg, "integer"):
		return "ожидалось целое число", true
	case strings.Contains(msg, "number"), strings.Contains(msg, "float"), strings.Contains(msg, "decimal"):
		return "ожидалось число", true
	case strings.Contains(msg, "string"):
		return "ожидалась строка", true
	case strings.Contains(msg, "boolean"), strings.Contains(msg, "bool"):
		return "ожидалось true/false", true
	case strings.Contains(msg, "array"):
		return "ожидался список", true
	case strings.Contains(msg, "object"):
		return "ожидался объект", true
	}

	return "неверный тип данных", true
}

func checkFormat(msg string) (string, bool) {
	if strings.Contains(msg, "format") || strings.Contains(msg, "regular expression") {
		return "некорректный формат данных", true
	}

	return "", false
}

func checkEnum(msg string, schema *openapi3.Schema) (string, bool) {
	if !strings.Contains(msg, "enum") {
		return "", false
	}

	out := "выбрано недопустимое значение"

	if schema != nil && len(schema.Enum) > 0 {
		var values []string
		for _, v := range schema.Enum {
			values = append(values, fmt.Sprintf("%v", v))
		}

		out += ". допустимые: [" + strings.Join(values, ", ") + "]"
	}

	return out, true
}

func addError(err error, details map[string]string) {
	var schemaErr *openapi3.SchemaError

	if errors.As(err, &schemaErr) {
		path := strings.Join(schemaErr.JSONPointer(), ".")
		if path == "" {
			path = "body"
		}

		details[path] = formatSchemaError(schemaErr.Error(), schemaErr.Schema)
	}
}

func extractErrors(msg string) map[string]string {
	details := make(map[string]string)

	for _, block := range strings.Split(msg, " | ") {
		if block = strings.TrimSpace(block); block == "" {
			continue
		}

		if match := reErrorTarget.FindStringSubmatch(block); len(match) > 1 {
			field := strings.ReplaceAll(strings.TrimPrefix(match[1], "/"), "/", ".")

			prefix := ""
			if strings.Contains(block, "in query") {
				prefix = "query."
			} else if strings.Contains(block, "in path") {
				prefix = "path."
			}

			details[prefix+field] = formatSchemaError(block, nil)
		}
	}

	return details
}
