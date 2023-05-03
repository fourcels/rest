package rest

import (
	"net/http"

	"github.com/swaggest/openapi-go/openapi3"
)

type option func(o *openapi3.Operation, reflect *openapi3.Reflector)

func WithSummary(val string) option {
	return func(op *openapi3.Operation, ref *openapi3.Reflector) {
		op.WithSummary(val)
	}
}

func WithDescription(val string) option {
	return func(op *openapi3.Operation, ref *openapi3.Reflector) {
		op.WithDescription(val)
	}
}
func WithTags(val ...string) option {
	return func(op *openapi3.Operation, ref *openapi3.Reflector) {
		op.WithTags(val...)
	}
}
func WithSecurity(key string) option {
	return func(op *openapi3.Operation, ref *openapi3.Reflector) {
		op.WithSecurity(map[string][]string{
			key: {},
		})
		ref.SetJSONResponse(op, new(ErrResponse), http.StatusUnauthorized)
	}
}
