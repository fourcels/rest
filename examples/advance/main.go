package main

import (
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/fourcels/rest"
	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/swgui"
)

var signingKey = []byte("secret")

func main() {
	s := rest.DefaultService()
	s.OpenAPI.Info.WithTitle("Advance Example")
	s.WithHttpSecurity("bearerAuth", rest.SchemeBearer)
	s.POST("/login", login())
	s.POST("/upload", upload())

	admin := s.Group("/admin", func(op *openapi3.Operation) {
		op.WithTags("Admin")
		op.WithSecurity(map[string][]string{
			"bearerAuth": {},
		})
	})
	admin.Use(echojwt.JWT(signingKey))
	admin.GET("/hello", hello())

	// Swagger UI endpoint at /docs.
	s.Docs("/docs", swgui.Config{
		SettingsUI: map[string]string{
			"persistAuthorization": "true",
		},
	})

	// Start server.
	log.Println("http://localhost:1323/docs")
	s.Start(":1323")
}

type jwtCustomClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

func login() rest.Interactor {
	// Declare input port type.
	type input struct {
		Username string `json:"username" minLenght:"3" default:"admin"`
		Password string `json:"password" minLength:"3" default:"a12345"`
	}

	// Declare output port type.
	type output struct {
		Token string `json:"token"`
	}
	// jwtCustomClaims are custom claims extending default ones.
	// See https://github.com/golang-jwt/jwt for more examples

	return rest.NewHandler(func(c echo.Context, in input, out *output) error {
		username := in.Username
		password := in.Password

		// Throws unauthorized error
		if username != "admin" || password != "a12345" {
			return echo.NewHTTPError(http.StatusBadRequest, "Incorrect username or password")
		}

		// Set custom claims
		claims := &jwtCustomClaims{
			"Jon Snow",
			true,
			jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
			},
		}

		// Create token with claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Generate encoded token and send it as response.
		t, err := token.SignedString(signingKey)
		if err != nil {
			return err
		}
		out.Token = t
		return nil
	}, func(op *openapi3.Operation) {
		op.WithSummary("Login")
		op.WithTags("Auth")
	})
}

func hello() rest.Interactor {
	return rest.NewHandler(func(c echo.Context, in struct{}, out *string) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		name := claims["name"].(string)
		*out = "Welcome " + name + "!"
		return nil
	}, func(op *openapi3.Operation) {
		op.WithSummary("Hello")
	})
}

func upload() rest.Interactor {
	// Declare input port type.
	type input struct {
		File *multipart.FileHeader `formData:"file"`
	}

	// Declare output port type.
	type output struct {
		Filename string `json:"filename"`
	}

	return rest.NewHandler(func(c echo.Context, in input, out *output) error {
		out.Filename = in.File.Filename
		return nil
	}, func(op *openapi3.Operation) {
		op.WithTags("Upload")
	})
}
