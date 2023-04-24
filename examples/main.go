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
