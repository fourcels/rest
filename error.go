package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPCodeAsError exposes HTTP status code as use case error that can be translated to response status.
type HTTPCodeAsError int

// Error return HTTP status text.
func (c HTTPCodeAsError) Error() string {
	return http.StatusText(int(c))
}

// HTTPStatus returns HTTP status code.
func (c HTTPCodeAsError) HTTPStatus() int {
	return int(c)
}

// ErrWithHTTPStatus exposes HTTP status code.
type ErrWithHTTPStatus interface {
	error
	HTTPStatus() int
}

// ErrWithFields exposes structured context of error.
type ErrWithFields interface {
	error
	Fields() map[string]interface{}
}

// ErrWithAppCode exposes application error code.
type ErrWithAppCode interface {
	error
	AppErrCode() int
}

// Err creates HTTP status code and ErrResponse for error.

func Err(err error) (int, ErrResponse) {
	if err == nil {
		panic("nil error received")
	}

	er := ErrResponse{}

	var (
		withHTTPStatus ErrWithHTTPStatus
		withAppCode    ErrWithAppCode
		withFields     ErrWithFields
	)

	er.err = err
	er.ErrorText = err.Error()
	er.httpStatusCode = http.StatusInternalServerError

	if he, ok := err.(*echo.HTTPError); ok {
		if he.Internal != nil {
			if herr, ok := he.Internal.(*echo.HTTPError); ok {
				he = herr
			}
		}
		er.httpStatusCode = he.Code
		if m, ok := he.Message.(string); ok {
			er.ErrorText = m
		}
	}
	if errors.As(err, &withHTTPStatus) {
		er.httpStatusCode = withHTTPStatus.HTTPStatus()
	}

	if errors.As(err, &withAppCode) {
		er.AppCode = withAppCode.AppErrCode()
	}

	if errors.As(err, &withFields) {
		er.Context = withFields.Fields()
	}

	if er.ErrorText == er.StatusText {
		er.ErrorText = ""
	}

	return er.httpStatusCode, er
}

// ErrResponse is HTTP error response body.
type ErrResponse struct {
	StatusText string                 `json:"status,omitempty" description:"Status text."`
	AppCode    int                    `json:"code,omitempty" description:"Application-specific error code."`
	ErrorText  string                 `json:"error,omitempty" description:"Error message."`
	Context    map[string]interface{} `json:"context,omitempty" description:"Application context."`

	err            error // Original error.
	httpStatusCode int   // HTTP response status code.
}

// Error implements error.
func (e ErrResponse) Error() string {
	if e.ErrorText != "" {
		return e.ErrorText
	}

	return e.StatusText
}

// Unwrap returns parent error.
func (e ErrResponse) Unwrap() error {
	return e.err
}
