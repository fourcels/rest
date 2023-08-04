package rest

import (
	"net/http"

	"github.com/swaggest/openapi-go"
)

type option func(oc openapi.OperationContext)

func WithSummary(val string) option {
	return func(oc openapi.OperationContext) {
		oc.SetSummary(val)
	}
}

func WithDescription(val string) option {
	return func(oc openapi.OperationContext) {
		oc.SetDescription(val)
	}
}
func WithTags(val ...string) option {
	return func(oc openapi.OperationContext) {
		oc.SetTags(val...)
	}
}
func WithSecurity(key string) option {
	return func(oc openapi.OperationContext) {
		oc.AddSecurity(key)
		oc.AddRespStructure(new(ErrResponse), openapi.WithHTTPStatus(http.StatusUnauthorized))
	}
}
