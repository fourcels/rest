package rest

// ParamIn defines parameter location.
type ParamIn string

const (
	// ParamInPath indicates path parameters, such as `/users/{id}`.
	ParamInPath = ParamIn("path")

	// ParamInQuery indicates query parameters, such as `/users?page=10`.
	ParamInQuery = ParamIn("query")

	// ParamInBody indicates body value, such as `{"id": 10}`.
	ParamInBody = ParamIn("json")

	// ParamInFormData indicates body form parameters.
	ParamInForm = ParamIn("form")

	// ParamInFormData indicates body form parameters.
	ParamInFormData = ParamIn("formData")

	// ParamInCookie indicates cookie parameters, which are passed ParamIn the `Cookie` header,
	// such as `Cookie: debug=0; gdpr=2`.
	ParamInCookie = ParamIn("cookie")

	// ParamInHeader indicates header parameters, such as `X-Header: value`.
	ParamInHeader = ParamIn("header")
)
