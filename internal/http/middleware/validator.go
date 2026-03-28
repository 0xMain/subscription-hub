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
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"
)

type OpenApiValidator struct {
	r routers.Router
}

func NewOpenApiValidator(swagger *openapi3.T) (*OpenApiValidator, error) {
	swagger.Servers = nil

	r, err := gorillamux.NewRouter(swagger)
	if err != nil {
		return nil, err
	}

	return &OpenApiValidator{r: r}, nil
}

func (v *OpenApiValidator) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if v.r == nil {
			c.Next()
			return
		}

		route, pathParams, err := v.r.FindRoute(c.Request)
		if err != nil {
			c.Next()
			return
		}

		input := &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				MultiError:         true,
				AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			},
		}

		if err = openapi3filter.ValidateRequest(c.Request.Context(), input); err != nil {
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
			return
		}

		c.Next()
	}
}

// unwrap вытаскивает все ошибки из общей кучи
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

// resolve находит конкретное сломанное поле
func (v *OpenApiValidator) resolve(err error, details map[string][]string, loc string) {
	if err == nil {
		return
	}

	var reqErr *openapi3filter.RequestError
	if errors.As(err, &reqErr) {
		if p := reqErr.Parameter; p != nil {
			loc += p.In + "." + p.Name + "."
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

// fallback достает ошибки из текста, если структура не ясна («костыль» на крайний случай)
func (v *OpenApiValidator) fallback(errMsg string, details map[string][]string, loc string) {
	for _, block := range strings.Split(errMsg, " | ") {
		first, _, _ := strings.Cut(block, "\n")
		first = strings.TrimSpace(first)
		if first == "" {
			continue
		}

		currLoc := loc
		for _, p := range [...]struct{ k, v string }{
			{"in query", "query."}, {"in path", "path."}, {"in header", "header."}, {"in cookie", "cookie."},
		} {
			if strings.Contains(first, p.k) {
				currLoc = p.v
				break
			}
		}

		msg := v.formatMsg(first, nil)
		v.report(details, currLoc, v.parsePath(first), msg, errors.New(first))
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

func (v *OpenApiValidator) formatPath(loc, field string) string {
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

	for _, c := range errorCheckers {
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

	if prefix, tail, found := strings.Cut(msg, ": "); found {
		lowPrefix := strings.ToLower(prefix)
		if strings.Contains(lowPrefix, "error") || strings.Contains(lowPrefix, "at \"") {
			if tail = strings.TrimSpace(tail); tail != "" {
				msg = tail
			}
		}
	}

	if strings.Contains(msg, "property \"") {
		if start, after, ok := strings.Cut(msg, "\""); ok {
			if _, end, ok := strings.Cut(after, "\""); ok {
				msg = strings.Join(strings.Fields(start+end), " ")
			}
		}
	}

	return strings.ToLower(strings.TrimSpace(msg))
}

type errChecker interface {
	check(msg string, schema *openapi3.Schema) (string, bool)
}

var errorCheckers = []errChecker{
	&requiredChecker{}, &typeChecker{}, &enumChecker{}, &limitChecker{}, &formatChecker{},
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
