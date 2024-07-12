package rest

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swaggest/openapi-go/openapi3"
)

//go:embed swagger.tmpl
var swagger string

type Scheme string

const (
	SchemeBearer = Scheme("bearer")
	SchemeBasic  = Scheme("basic")
)

type Service struct {
	*echo.Echo
	baseUrl   string
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

func NewService(baseUrl ...string) *Service {
	s := &Service{}

	s.reflector = &openapi3.Reflector{}
	s.OpenAPI = &openapi3.Spec{Openapi: "3.0.3"}
	if len(baseUrl) > 0 {
		s.baseUrl = baseUrl[0]
		s.OpenAPI.WithServers(openapi3.Server{
			URL: s.baseUrl,
		})
	}

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
	s.group = s.Group("")

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
	group.Group = s.Echo.Group(s.baseUrl + parenthesesToColon(prefix))
	group.reflector = s.reflector
	group.prefix = prefix
	group.ops = ops
	return group
}

func (s *Service) Docs(pattern string, config ...map[string]any) {
	pattern = strings.TrimRight(pattern, "/")
	s.Echo.GET(s.baseUrl+pattern+"/openapi.json", func(c echo.Context) error {
		schema, err := s.OpenAPI.MarshalJSON()
		if err != nil {
			return err
		}
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		return c.String(http.StatusOK, string(schema))
	})
	s.Echo.Any(s.baseUrl+pattern+"*", func(c echo.Context) error {
		t := template.Must(template.New("swagger").Parse(swagger))
		var html bytes.Buffer
		var setting map[string]any
		if len(config) > 0 {
			setting = config[0]
		}
		t.Execute(&html, map[string]any{
			"AssetBase":   "https://unpkg.com/swagger-ui-dist",
			"SwaggerJson": s.baseUrl + pattern + "/openapi.json",
			"Setting":     setting,
		})
		return c.HTML(http.StatusOK, html.String())
	})
}

func (s *Service) WithSecurity(key string, securityScheme *openapi3.SecurityScheme) {
	s.OpenAPI.ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(key,
		openapi3.SecuritySchemeOrRef{
			SecurityScheme: securityScheme,
		})
}

func (s *Service) WithHttpBearerSecurity(key string) {
	s.WithHttpSecurity(key, SchemeBearer)
}
func (s *Service) WithHttpBasicSecurity(key string) {
	s.WithHttpSecurity(key, SchemeBasic)
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

func (s *Service) WithTags(val ...string) {
	tags := make([]openapi3.Tag, 0)
	for _, v := range val {
		tags = append(tags, openapi3.Tag{Name: v})
	}
	s.OpenAPI.WithTags(tags...)
}
