package main

import (
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/fourcels/rest"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

var signingKey = []byte("secret")

func main() {
	s := rest.NewService()
	s.OpenAPI.Info.WithTitle("Advance Example")
	s.WithHttpBearerSecurity("bearerAuth")

	s.POST("/login", login())
	s.POST("/upload", upload())

	admin := s.Group("/admin", rest.WithTags("Admin"), rest.WithSecurity("bearerAuth"))

	admin.Use(echojwt.JWT(signingKey))
	admin.GET("/hello", hello())

	// Swagger UI endpoint at /docs.
	s.Docs("/docs", map[string]any{
		"persistAuthorization": true,
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
		Username string `json:"username" minLength:"3" default:"admin"`
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
	}, rest.WithSummary("login"), rest.WithTags("Login"))
}

func hello() rest.Interactor {
	return rest.NewHandler(func(c echo.Context, in struct{}, out *string) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		name := claims["name"].(string)
		*out = "Welcome " + name + "!"
		return nil
	})
}

func upload() rest.Interactor {
	// Declare input port type.
	type input struct {
		Type uint                  `query:"type"`
		File *multipart.FileHeader `formData:"file"`
	}

	// Declare output port type.
	type output struct {
		Filename string `json:"filename"`
	}

	return rest.NewHandler(func(c echo.Context, in input, out *output) error {
		out.Filename = in.File.Filename
		return nil
	}, rest.WithTags("Upload"))
}
