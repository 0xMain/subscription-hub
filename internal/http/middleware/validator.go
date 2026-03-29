package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/0xMain/subscription-hub/internal/http/errs"
	"github.com/0xMain/subscription-hub/internal/http/res"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/gin-gonic/gin"
)

// defCheckers базовый набор правил для обработки ошибок валидации
var defCheckers = []checker{
	&requiredChecker{}, &typeChecker{}, &enumChecker{}, &limitChecker{}, &formatChecker{},
}

type (
	OpenApiValidator struct {
		cache    map[string]*route
		checkers []checker
	}

	route struct {
		path      string
		pathItem  *openapi3.PathItem
		operation *openapi3.Operation
		method    string
	}

	checker interface {
		check(msg string, schema *openapi3.Schema) (string, bool)
	}
)

func NewOpenApiValidator(swagger *openapi3.T) (*OpenApiValidator, error) {
	swagger.Servers = nil

	v := &OpenApiValidator{
		cache:    make(map[string]*route),
		checkers: defCheckers,
	}

	for path, pathItem := range swagger.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			key := v.routeKey(method, path)

			v.cache[key] = &route{
				path:      path,
				pathItem:  pathItem,
				operation: operation,
				method:    method,
			}
		}
	}

	return v, nil
}

func (v *OpenApiValidator) routeKey(method, path string) string {
	var b strings.Builder

	b.Grow(len(method) + 1 + len(path))

	b.WriteString(method)
	b.WriteByte(' ')

	for i := 0; i < len(path); i++ {
		switch char := path[i]; char {
		case '{':
			b.WriteByte(':')
		case '}':
			continue
		default:
			b.WriteByte(char)
		}
	}

	return b.String()
}

func (v *OpenApiValidator) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		entry, ok := v.cache[v.routeKey(c.Request.Method, c.FullPath())]
		if !ok {
			c.Next()
			return
		}

		params := make(map[string]string, len(c.Params))
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}

		input := &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: params,

			Route: &routers.Route{
				Path:      entry.path,
				PathItem:  entry.pathItem,
				Method:    entry.method,
				Operation: entry.operation,
			},

			Options: &openapi3filter.Options{
				MultiError:         true,
				AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			},
		}

		if err := openapi3filter.ValidateRequest(c.Request.Context(), input); err != nil {
			v.handleError(c, err)

			c.Abort()
			return
		}

		c.Next()
	}
}

func (v *OpenApiValidator) handleError(c *gin.Context, err error) {
	details := make(map[string][]string)

	v.unwrap(err, details, "")

	if len(details) == 0 {
		v.fallback(err.Error(), details, "")
	}

	if _, isBadBody := details["body"]; len(details) == 0 || isBadBody {
		res.Error(c, http.StatusBadRequest, errs.MsgInvalidFormatErr, nil)
		return
	}

	res.Error(c, http.StatusUnprocessableEntity, errs.MsgValidationErr, details)
}

// unwrap разбирает пачку ошибок по одной
func (v *OpenApiValidator) unwrap(err error, details map[string][]string, loc string) {
	if err == nil {
		return
	}

	var multiErr openapi3.MultiError
	if errors.As(err, &multiErr) {
		for _, e := range multiErr {
			v.resolve(e, details, loc)
		}

		return
	}

	v.resolve(err, details, loc)
}

// resolve вытаскивает из ошибки название поля и текст проблемы
func (v *OpenApiValidator) resolve(err error, details map[string][]string, loc string) {
	if err == nil {
		return
	}

	var reqErr *openapi3filter.RequestError
	if errors.As(err, &reqErr) {
		if p := reqErr.Parameter; p != nil {
			loc = v.formatParamLoc(loc, p.In, p.Name)
		}

		v.unwrap(reqErr.Err, details, loc)
		return
	}

	var schErr *openapi3.SchemaError
	if errors.As(err, &schErr) {
		path := strings.Join(schErr.JSONPointer(), ".")
		msg := v.formatMsg(schErr.Reason, schErr.Schema)

		v.report(details, loc, path, msg, errors.New(schErr.Reason))
		return
	}

	msg := v.formatMsg(err.Error(), nil)
	v.report(details, loc, "", msg, err)
}

// fallback ищет ошибки в тексте, если не удалось разобрать их структуру («костыль» на крайний случай)
func (v *OpenApiValidator) fallback(errMsg string, details map[string][]string, loc string) {
	start := 0

	for {
		idx := strings.Index(errMsg[start:], " | ")

		var block string
		if idx == -1 {
			block = errMsg[start:]
		} else {
			block = errMsg[start : start+idx]
		}

		v.parseBlock(block, details, loc)

		if idx == -1 {
			break
		}

		start += idx + 3
	}
}

func (v *OpenApiValidator) report(details map[string][]string, loc, field, msg string, err error) {
	path := v.formatPath(loc, field)

	if msg == "" || msg == errs.MsgGenericInvalidDetailErr {
		log.Printf("неизвестная ошибка валидации OAPI (источник=%s, причина=%v)", path, v.clean(err.Error()))
		msg = v.fallbackMsg(loc)
	}

	v.add(details, path, msg)
}

func (v *OpenApiValidator) fallbackMsg(loc string) string {
	switch root, _, _ := strings.Cut(loc, "."); root {
	case "query", "path", "header", "cookie":
		return errs.MsgInvalidStructureDetailErr
	default:
		return errs.MsgGenericInvalidDetailErr
	}
}

func (v *OpenApiValidator) add(details map[string][]string, key, newMsg string) {
	for _, msg := range details[key] {
		if msg == newMsg {
			return
		}
	}

	details[key] = append(details[key], newMsg)
}

func (v *OpenApiValidator) parseBlock(block string, details map[string][]string, loc string) {
	if idx := strings.IndexByte(block, '\n'); idx != -1 {
		block = block[:idx]
	}
	block = strings.TrimSpace(block)
	if block == "" {
		return
	}

	currLoc := loc
	for _, p := range [...]struct{ k, v string }{
		{"in query", "query."}, {"in path", "path."},
		{"in header", "header."}, {"in cookie", "cookie."},
	} {
		if strings.Contains(block, p.k) {
			currLoc = p.v
			break
		}
	}

	msg := v.formatMsg(block, nil)
	field := v.parsePath(block)

	v.report(details, currLoc, field, msg, errors.New(block))
}

func (v *OpenApiValidator) parsePath(block string) string {
	_, content, found := strings.Cut(block, "\"")
	if !found {
		return ""
	}

	path, _, found := strings.Cut(content, "\"")
	if !found {
		return ""
	}

	return strings.ReplaceAll(strings.TrimPrefix(path, "/"), "/", ".")
}

func (*OpenApiValidator) formatParamLoc(base, in, name string) string {
	var b strings.Builder

	b.Grow(len(base) + len(in) + len(name) + 2)

	b.WriteString(base)
	b.WriteString(in)
	b.WriteByte('.')
	b.WriteString(name)
	b.WriteByte('.')

	return b.String()
}

func (*OpenApiValidator) formatPath(loc, field string) string {
	if path := strings.TrimSuffix(loc+field, "."); path != "" {
		return path
	}

	return "body"
}

func (v *OpenApiValidator) formatMsg(reason string, sch *openapi3.Schema) string {
	msg := v.clean(reason)

	if msg == "" {
		return errs.MsgGenericInvalidDetailErr
	}

	for _, c := range v.checkers {
		if out, ok := c.check(msg, sch); ok {
			return out
		}
	}

	return errs.MsgGenericInvalidDetailErr
}

func (v *OpenApiValidator) clean(msg string) string {
	if idx := strings.IndexByte(msg, '\n'); idx != -1 {
		msg = msg[:idx]
	}

	for {
		idx := strings.IndexByte(msg, ':')
		if idx == -1 {
			break
		}

		prefix := msg[:idx]

		switch {
		case
			strings.Contains(prefix, "failed"),
			strings.Contains(prefix, "error"),
			strings.Contains(prefix, "at \""):

			msg = strings.TrimSpace(msg[idx+1:])
			continue
		}
		break
	}

	if idx := strings.Index(msg, "property \""); idx != -1 {
		msg = v.hideProperty(msg, idx)
	}

	return strings.ToLower(strings.TrimSpace(msg))
}

func (v *OpenApiValidator) hideProperty(msg string, idx int) string {
	start := strings.IndexByte(msg[idx:], '"') + idx
	if start == -1 {
		return msg
	}

	end := strings.IndexByte(msg[start+1:], '"') + start + 1
	if end == -1 {
		return msg
	}

	size := len(msg) - (end - start + 1) + 1

	var b strings.Builder
	b.Grow(size)

	b.WriteString(strings.TrimSpace(msg[:start]))

	tail := strings.TrimSpace(msg[end+1:])
	if tail != "" {
		b.WriteByte(' ')
		b.WriteString(tail)
	}

	return b.String()
}

func spec(sch *openapi3.Schema, key string) (string, bool) {
	if sch == nil || sch.Extensions == nil || key == "" {
		return "", false
	}

	xExt, ok := sch.Extensions["x-errors"].(map[string]interface{})
	if !ok {
		return "", false
	}

	if custom, ok := xExt[key].(string); ok && custom != "" {
		return custom, true
	}

	return "", false
}

func withSpec(sch *openapi3.Schema, key string, defaultMsg string) (string, bool) {
	if custom, ok := spec(sch, key); ok && custom != "" {
		return custom, true
	}

	return defaultMsg, defaultMsg != ""
}
