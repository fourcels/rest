package rest

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mcuadros/go-defaults"
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/swgui"
	"github.com/swaggest/swgui/v4cdn"
)

type Scheme string

const (
	SchemeBearer = Scheme("bearer")
	SchemeBasic  = Scheme("basic")
)

type Service struct {
	*echo.Echo
	prefix    string
	group     *echo.Group
	ops       []option
	reflector *openapi3.Reflector
	OpenAPI   *openapi3.Spec
}

func customHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	code, errRes := Err(err)

	// Send response
	if c.Request().Method == http.MethodHead { // Issue #608
		c.NoContent(code)
	} else {
		c.JSON(code, errRes)
	}
}

func DefaultService() *Service {
	s := Service{}

	s.reflector = &openapi3.Reflector{}
	s.OpenAPI = &openapi3.Spec{Openapi: "3.0.3"}
	s.OpenAPI.Info.
		WithTitle("Things API").
		WithVersion("1.2.3").
		WithDescription("API description")
	s.reflector.Spec = s.OpenAPI
	e := echo.New()
	e.HideBanner = true
	e.Binder = &CustomBinder{}
	e.Validator = &CustomValidator{}
	e.HTTPErrorHandler = customHTTPErrorHandler

	// Root level middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	s.Echo = e
	s.group = e.Group("")

	return &s
}

func setHeader(c echo.Context, name, value string) {
	c.Response().Header().Set(name, value)
}
func setCookie(c echo.Context, cookie, value string) {
	options := strings.Split(cookie, ",")
	options[0] = fmt.Sprintf("%s=%s", options[0], value)
	c.Response().Header().Add("Set-Cookie", strings.Join(options, ";"))
}

func setupOutput(c echo.Context, out any) error {
	if out == nil {
		return nil
	}
	typ := reflect.TypeOf(out).Elem()
	val := reflect.ValueOf(out).Elem()
	// !struct
	if typ.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if typeField.Anonymous {
			continue
		}
		if header := typeField.Tag.Get("header"); header != "" {
			setHeader(c, header, fmt.Sprintf("%v", structField))
		}
		if cookie := typeField.Tag.Get("cookie"); cookie != "" {
			setCookie(c, cookie, fmt.Sprintf("%v", structField))
		}
	}
	return nil
}

func pathColonToParentheses(pattern string) string {
	re := regexp.MustCompile(`:(\w+)`)
	return re.ReplaceAllString(pattern, "{$1}")
}

// Method adds routes for `basePattern` that matches the `method` HTTP method.
func (s *Service) add(method, pattern string, h Interactor, middleware ...echo.MiddlewareFunc) {
	operation := &openapi3.Operation{}
	s.reflector.SetRequest(operation, h.Input(), method)
	s.reflector.SetJSONResponse(operation, h.Output(), http.StatusOK)
	for _, op := range append(s.ops, h.Options()...) {
		op(operation)
	}
	path := pathColonToParentheses(s.prefix + pattern)
	if err := s.OpenAPI.AddOperation(method, path, *operation); err != nil {
		log.Println(method, path, err)
	}

	s.group.Add(method, pattern, func(c echo.Context) error {
		in := h.Input()
		defaults.SetDefaults(in)
		if err := c.Bind(in); err != nil {
			return err
		}
		if err := c.Validate(in); err != nil {
			return err
		}
		out := h.Output()
		if err := h.Interact(c, in, out); err != nil {
			return err
		}
		setupOutput(c, out)
		return c.JSON(http.StatusOK, out)
	}, middleware...)
}

func (s *Service) GET(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) {
	s.add(http.MethodGet, pattern, h, middleware...)
}

func (s *Service) POST(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) {
	s.add(http.MethodPost, pattern, h, middleware...)
}
func (s *Service) PATCH(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) {
	s.add(http.MethodPatch, pattern, h, middleware...)
}
func (s *Service) PUT(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) {
	s.add(http.MethodPut, pattern, h, middleware...)
}
func (s *Service) DELETE(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) {
	s.add(http.MethodDelete, pattern, h, middleware...)
}

func (s *Service) Docs(pattern string, config swgui.Config) {
	pattern = strings.TrimRight(pattern, "/")
	s.Echo.GET(pattern+"/openapi.json", func(c echo.Context) error {
		schema, err := s.OpenAPI.MarshalJSON()
		if err != nil {
			return err
		}
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		return c.String(http.StatusOK, string(schema))
	})
	h := v4cdn.NewWithConfig(config)
	s.Any(pattern+"*", echo.WrapHandler(
		h(s.OpenAPI.Info.Title, pattern+"/openapi.json", pattern),
	))
}

func (s *Service) WithSecurity(key string, securityScheme *openapi3.SecurityScheme) {
	s.OpenAPI.ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(key,
		openapi3.SecuritySchemeOrRef{
			SecurityScheme: securityScheme,
		})
}

func (s *Service) WithHttpSecurity(key string, scheme Scheme) {
	s.WithSecurity(key, &openapi3.SecurityScheme{
		HTTPSecurityScheme: &openapi3.HTTPSecurityScheme{
			Scheme: string(scheme),
		},
	})
}
func (s *Service) WithAPIKeySecurity(key, name string, in openapi3.APIKeySecuritySchemeIn) {
	s.WithSecurity(key, &openapi3.SecurityScheme{
		APIKeySecurityScheme: &openapi3.APIKeySecurityScheme{
			In:   in,
			Name: name,
		},
	})
}

func (s *Service) Group(prefix string, ops ...option) *Service {
	group := &Service{}
	group.group = s.group.Group(prefix)
	group.reflector = s.reflector
	group.OpenAPI = s.OpenAPI
	group.prefix = s.prefix + prefix
	group.ops = append(s.ops, ops...)
	return group
}

func (s *Service) Use(middleware ...echo.MiddlewareFunc) {
	s.group.Use(middleware...)
}
