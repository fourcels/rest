package rest

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mcuadros/go-defaults"
	"github.com/swaggest/openapi-go/openapi3"
)

type NoContent struct{}

type Group struct {
	*echo.Group
	prefix    string
	ops       []option
	reflector *openapi3.Reflector
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

func parenthesesToColon(pattern string) string {
	re := regexp.MustCompile(`\{(\w+)\}`)
	return re.ReplaceAllString(pattern, ":$1")
}

// Method adds routes for `basePattern` that matches the `method` HTTP method.
func (g *Group) add(method, pattern string, h Interactor, middleware ...echo.MiddlewareFunc) *echo.Route {
	path := g.prefix + pattern
	oc, err := g.reflector.NewOperationContext(method, path)
	if err != nil {
		log.Println(method, path, err)
	}

	for _, op := range append(g.ops, h.Options()...) {
		op(oc)
	}

	oc.SetSummary(h.Summary())
	oc.AddReqStructure(h.Input())
	oc.AddRespStructure(h.Output())

	if err := g.reflector.AddOperation(oc); err != nil {
		log.Println(method, path, err)
	}

	return g.Add(method, parenthesesToColon(pattern), func(c echo.Context) error {
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
		if _, ok := out.(*NoContent); ok {
			return c.NoContent(http.StatusNoContent)
		}
		return c.JSON(http.StatusOK, out)
	}, middleware...)
}

func (g *Group) GET(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) *echo.Route {
	return g.add(http.MethodGet, pattern, h, middleware...)
}

func (g *Group) POST(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) *echo.Route {
	return g.add(http.MethodPost, pattern, h, middleware...)
}
func (g *Group) PATCH(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) *echo.Route {
	return g.add(http.MethodPatch, pattern, h, middleware...)
}
func (g *Group) PUT(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) *echo.Route {
	return g.add(http.MethodPut, pattern, h, middleware...)
}
func (g *Group) DELETE(pattern string, h Interactor, middleware ...echo.MiddlewareFunc) *echo.Route {
	return g.add(http.MethodDelete, pattern, h, middleware...)
}
