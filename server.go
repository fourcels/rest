package rest

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	group     *Group
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

func DefaultService(ops ...option) *Service {
	s := &Service{}

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
	s.group = s.Group("", ops...)

	return s
}

func (s *Service) GET(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) *echo.Route {
	return s.group.GET(pattern, h, middleware...)
}
func (s *Service) POST(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) *echo.Route {
	return s.group.POST(pattern, h, middleware...)
}
func (s *Service) PATCH(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) *echo.Route {
	return s.group.PATCH(pattern, h, middleware...)
}
func (s *Service) PUT(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) *echo.Route {
	return s.group.PUT(pattern, h, middleware...)
}
func (s *Service) DELETE(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) *echo.Route {
	return s.group.DELETE(pattern, h, middleware...)
}

func (s *Service) Group(prefix string, ops ...option) *Group {
	group := &Group{}
	group.Group = s.Echo.Group(prefix)
	group.reflector = s.reflector
	group.prefix = prefix
	group.ops = ops
	return group
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
	s.Echo.Any(pattern+"*", echo.WrapHandler(
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
