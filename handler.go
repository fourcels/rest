package rest

import (
	"github.com/labstack/echo/v4"
)

type Interactor interface {
	Input() any
	Output() any
	Interact(c echo.Context, in, out any) error
	Options() []option
}

type Handler[i, o any] struct {
	handler interact[i, o]
	options []option
}

type interact[i, o any] func(c echo.Context, in i, out *o) error

func NewHandler[i, o any](handler interact[i, o], ops ...option) Interactor {
	return &Handler[i, o]{
		handler: handler,
		options: ops,
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
