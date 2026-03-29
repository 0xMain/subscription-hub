package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/0xMain/subscription-hub/internal/http/errs"
	"github.com/0xMain/subscription-hub/internal/pkg/strutil"
	"github.com/getkin/kin-openapi/openapi3"
)

// Ключи для поиска кастомных текстов ошибок в расширении x-errors схемы OpenAPI
const (
	specKeyType = "type"

	specKeyEnum = "enum"

	specKeyPattern = "pattern"

	specKeyEmpty    = "empty"
	specKeyRequired = "required"
	specKeyNullable = "nullable"

	specKeyMinimum = "minimum"
	specKeyMaximum = "maximum"

	specKeyMinItems = "minItems"
	specKeyMaxItems = "maxItems"

	specKeyMinLength = "minLength"
	specKeyMaxLength = "maxLength"
)

type requiredChecker struct{}

func (r *requiredChecker) check(msg string, sch *openapi3.Schema) (string, bool) {
	if strings.Contains(msg, "empty value") {
		return withSpec(sch, specKeyEmpty, errs.MsgEmptyNotAllowedDetailErr)
	}

	if strings.Contains(msg, "is ") {
		if strings.Contains(msg, "missing") || strings.Contains(msg, "required") {
			return withSpec(sch, specKeyRequired, errs.MsgRequiredDetailErr)
		}
	}

	return "", false
}

type typeChecker struct{}

func (t *typeChecker) check(msg string, sch *openapi3.Schema) (string, bool) {
	if strings.Contains(msg, "not nullable") {
		return withSpec(sch, specKeyNullable, errs.MsgNotNullableDetailErr)
	}

	if idx := strings.Index(msg, "must be a"); idx != -1 {
		tail := msg[idx:]

		switch {
		case strings.Contains(tail, "string"):
			return withSpec(sch, specKeyType, errs.MsgExpectedStringDetailErr)
		case strings.Contains(tail, "integer"):
			return withSpec(sch, specKeyType, errs.MsgExpectedIntegerDetailErr)
		case strings.Contains(tail, "number"):
			return withSpec(sch, specKeyType, errs.MsgExpectedNumberDetailErr)
		case strings.Contains(tail, "boolean"):
			return withSpec(sch, specKeyType, errs.MsgExpectedBooleanDetailErr)
		case strings.Contains(tail, "array"):
			return withSpec(sch, specKeyType, errs.MsgExpectedArrayDetailErr)
		case strings.Contains(tail, "object"):
			return withSpec(sch, specKeyType, errs.MsgExpectedObjectDetailErr)
		}
	}

	return "", false
}

type formatChecker struct{}

func (*formatChecker) check(msg string, sch *openapi3.Schema) (string, bool) {
	if strings.Contains(msg, "regular expression") {
		return withSpec(sch, specKeyPattern, errs.MsgInvalidFormatDetailErr)
	}

	return "", false
}

type enumChecker struct{}

func (e *enumChecker) check(msg string, sch *openapi3.Schema) (string, bool) {
	if !strings.Contains(msg, "allowed values") && !strings.Contains(msg, "enum") {
		return "", false
	}

	resMsg, _ := withSpec(sch, specKeyEnum, errs.MsgInvalidValueDetailErr)
	allowed := e.allowedValues(msg, sch)
	if allowed == "" {
		return resMsg, true
	}

	var b strings.Builder
	b.Grow(len(resMsg) + len(allowed) + 15)
	b.WriteString(resMsg)
	b.WriteString("; допустимые: ")
	b.WriteString(allowed)

	return b.String(), true
}

func (e *enumChecker) allowedValues(msg string, sch *openapi3.Schema) string {
	if sch != nil && len(sch.Enum) > 0 {
		var b strings.Builder
		b.Grow(len(sch.Enum) * 12)

		b.WriteByte('[')

		for i, v := range sch.Enum {
			if i > 0 {
				b.WriteString(", ")
			}

			switch val := v.(type) {
			case string:
				b.WriteString(val)
			case int:
				b.WriteString(strconv.Itoa(val))
			case float64:
				b.WriteString(strconv.FormatFloat(val, 'g', -1, 64))
			case bool:
				b.WriteString(strconv.FormatBool(val))
			default:
				_, _ = fmt.Fprint(&b, v)
			}
		}

		b.WriteByte(']')
		return b.String()
	}

	_, tail, found := strings.Cut(msg, "values ")
	if !found {
		return ""
	}

	clean := strings.ReplaceAll(tail, `\"`, "")
	clean = strings.ReplaceAll(clean, `"`, "")
	clean = strings.ReplaceAll(clean, " ", "")

	return strings.ReplaceAll(clean, ",", ", ")
}

type limitChecker struct{}

func (l *limitChecker) check(msg string, sch *openapi3.Schema) (string, bool) {
	if !l.isLimit(msg) {
		return "", false
	}

	isMin, ok := l.parseBound(msg)
	if !ok {
		return "", false
	}

	schType := l.schType(msg, sch)

	if sk := l.specKey(schType, isMin); sk != "" {
		if custom, ok := withSpec(sch, sk, ""); ok && custom != "" {
			return custom, true
		}
	}

	limitStr := l.limitVal(msg, sch, isMin, schType)

	count, err := strconv.Atoi(limitStr)
	if limitStr == "" || err != nil {
		if isMin {
			return errs.MsgTooSmallDetailErr, true
		}
		return errs.MsgTooLargeDetailErr, true
	}

	var b strings.Builder
	b.Grow(48)

	if isMin {
		b.WriteString("минимум ")
	} else {
		b.WriteString("максимум ")
	}
	b.WriteString(limitStr)

	switch schType {
	case "string":
		b.WriteByte(' ')
		b.WriteString(strutil.Plural(count, "символ", "символа", "символов"))
	case "array":
		b.WriteByte(' ')
		b.WriteString(strutil.Plural(count, "элемент", "элемента", "элементов"))
	}

	return b.String(), true
}

func (l *limitChecker) isLimit(msg string) bool {
	keywords := [...]string{"min", "max", "length", "item", "least", "most"}
	for _, k := range keywords {
		if strings.Contains(msg, k) {
			return true
		}
	}

	return false
}

func (l *limitChecker) parseBound(msg string) (isMin bool, ok bool) {
	switch {
	case strings.Contains(msg, "min"),
		strings.Contains(msg, "least"),
		strings.Contains(msg, "short"):
		return true, true

	case strings.Contains(msg, "max"),
		strings.Contains(msg, "most"),
		strings.Contains(msg, "long"):
		return false, true

	default:
		return false, false
	}
}

func (l *limitChecker) limitVal(msg string, sch *openapi3.Schema, isMin bool, schType string) string {
	if sch != nil {
		switch schType {
		case "string":
			if isMin {
				return strconv.FormatUint(sch.MinLength, 10)
			}
			if sch.MaxLength != nil {
				return strconv.FormatUint(*sch.MaxLength, 10)
			}
		case "array":
			if isMin {
				return strconv.FormatUint(sch.MinItems, 10)
			}
			if sch.MaxItems != nil {
				return strconv.FormatUint(*sch.MaxItems, 10)
			}
		case "number":
			if isMin {
				if sch.Min != nil {
					return strconv.FormatFloat(*sch.Min, 'g', -1, 64)
				}
			} else {
				if sch.Max != nil {
					return strconv.FormatFloat(*sch.Max, 'g', -1, 64)
				}
			}
		}
	}

	return l.parseVal(msg)
}

func (*limitChecker) parseVal(msg string) string {
	end := -1
	for i := len(msg) - 1; i >= 0; i-- {
		if msg[i] >= '0' && msg[i] <= '9' {
			if end == -1 {
				end = i + 1
			}
		} else if end != -1 {
			return msg[i+1 : end]
		}
	}

	if end != -1 {
		return msg[:end]
	}

	return ""
}

func (*limitChecker) schType(msg string, sch *openapi3.Schema) string {
	if sch != nil && sch.Type != nil {

		types := sch.Type.Slice()

		if len(types) > 0 {
			switch types[0] {
			case "string":
				return "string"
			case "array":
				return "array"
			case "number", "integer":
				return "number"
			}
		}
	}

	switch {
	case strings.Contains(msg, "item"), strings.Contains(msg, "array"):
		return "array"
	case strings.Contains(msg, "length"), strings.Contains(msg, "string"):
		return "string"
	default:
		return "number"
	}
}

func (*limitChecker) specKey(schType string, isMin bool) string {
	switch schType {
	case "string":
		if isMin {
			return specKeyMinLength
		}
		return specKeyMaxLength
	case "array":
		if isMin {
			return specKeyMinItems
		}
		return specKeyMaxItems
	case "number":
		if isMin {
			return specKeyMinimum
		}
		return specKeyMaximum
	default:
		return ""
	}
}
