# REST with Clean Architecture for Go

[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/fourcels/rest)

Inspired by [`swaggest/rest`](https://github.com/swaggest/rest)

## Features

- Built with [`echo`](https://github.com/labstack/echo).
- Automatic OpenAPI 3 documentation with
  [`openapi-go`](https://github.com/swaggest/openapi-go).
- Automatic request JSON schema validation with
  [`jsonschema-go`](https://github.com/swaggest/jsonschema-go),
  [`gojsonschema`](https://github.com/xeipuuv/gojsonschema).
- Embedded [Swagger UI](https://swagger.io/tools/swagger-ui/).

## Usage

### Request

Go struct with field tags defines input.

```go
// Declare input port type.
type helloInput struct {
    Locale string `query:"locale" default:"en-US" pattern:"^[a-z]{2}-[A-Z]{2}$" enum:"zh-CN,en-US"`
    Name   string `path:"name" minLength:"3"` // Field tags define parameter location and JSON schema constraints.

    _      struct{} `title:"My Struct" description:"Holds my data."`
}
```

Input data can be located in:

- `path` parameter in request URI, e.g. `/users/:name`,
- `query` parameter in request URI, e.g. `/users?locale=en-US`,
- `form` parameter in request body with `application/x-www-form-urlencoded`
  content,
- `formData` parameter in request body with `multipart/form-data` content,
- `json` parameter in request body with `application/json` content,
- `cookie` parameter in request cookie,
- `header` parameter in request header.

Field tags

- number `maximum`, `exclusiveMaximum`, `minimum`, `exclusiveMinimum`,
  `multipleOf`
- string `minLength`, `maxLength`, `pattern`, `format`
- array `minItems`, `maxItems`, `uniqueItems`
- all `title`, `description`, `default`, `const`, `enum`

Additional field tags describe JSON schema constraints, please check
[documentation](https://github.com/swaggest/jsonschema-go#field-tags).

## Response

```go
// Declare output port type.
type helloOutput struct {
    Now     time.Time `header:"X-Now" json:"-"`
    Message string    `json:"message"`
    Sess    string    `cookie:"sess,httponly,secure,max-age=86400,samesite=lax"`
}
```

Output data can be located in:

- `json` for response body with `application/json` content,
- `header` for values in response header,
- `cookie` for cookie values, cookie fields can have configuration in field tag
  (same as in actual cookie, but with comma separation).

## Example

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fourcels/rest"
	"github.com/labstack/echo/v4"
	"github.com/swaggest/swgui"
)

func main() {
	s := rest.DefaultService()
	s.OpenAPI.Info.WithTitle("Basic Example")
	s.GET("/hello/:name", hello())

	// Swagger UI endpoint at /docs.
	s.Docs("/docs", swgui.Config{})

	// Start server.
	log.Println("http://localhost:1323/docs")
	s.Start(":1323")
}

func hello() rest.Interactor {
	// Declare input port type.
	type input struct {
		Name   string   `path:"name" minLength:"3"` // Field tags define parameter
		Locale string   `query:"locale" default:"en-US" pattern:"^[a-z]{2}-[A-Z]{2}$" enum:"zh-CN,en-US"`
		_      struct{} `title:"My Struct" description:"Holds my data."`
	}

	// Declare output port type.
	type output struct {
		Now     time.Time `header:"X-Now" json:"-"`
		Message string    `json:"message"`
	}

	messages := map[string]string{
		"en-US": "Hello, %s!",
		"zh-CN": "你好, %s!",
	}
	return rest.NewHandler(func(c echo.Context, in input, out *output) error {
		msg := messages[in.Locale]
		out.Now = time.Now()
		out.Message = fmt.Sprintf(msg, in.Name)
		return nil
	})
}
```

## Use Case

- Route Group

```go
admin := s.Group("/admin")
admin.GET("/hello", hello())
```
