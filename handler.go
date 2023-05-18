package rest

import (
	"runtime"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/labstack/echo/v4"
)

type Interactor interface {
	Input() any
	Output() any
	Interact(c echo.Context, in, out any) error
	Options() []option
	Summary() string
}

type Handler[i, o any] struct {
	handler interact[i, o]
	options []option
	summary string
}

type interact[i, o any] func(c echo.Context, in i, out *o) error

func getSummary() string {
	counter, _, _, success := runtime.Caller(2)

	if !success {
		return ""
	}
	name := strings.Split(runtime.FuncForPC(counter).Name(), ".")

	return strings.Join(camelcase.Split(name[len(name)-1]), " ")
}

func NewHandler[i, o any](handler interact[i, o], ops ...option) Interactor {
	return &Handler[i, o]{
		handler: handler,
		options: ops,
		summary: getSummary(),
	}
}

func (h *Handler[i, o]) Interact(c echo.Context, in, out any) error {
	return h.handler(c, *in.(*i), out.(*o))
}

func (h *Handler[i, o]) Input() any {
	return new(i)
}

func (h *Handler[i, o]) Output() any {
	return new(o)
}
func (h *Handler[i, o]) Options() []option {
	return h.options
}
func (h *Handler[i, o]) Summary() string {
	return h.summary
}
