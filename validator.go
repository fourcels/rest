package rest

import (
	"net/http"
	"reflect"

	"github.com/swaggest/jsonschema-go"
	"github.com/xeipuuv/gojsonschema"
)

type CustomValidator struct{}

type ValidatorError struct {
	HTTPCodeAsError
	paramIn      ParamIn
	ResultErrors []gojsonschema.ResultError
}

func (ve *ValidatorError) Fields() map[string]any {
	fields := make(map[string]any)
	for _, re := range ve.ResultErrors {
		fieldName := string(ve.paramIn) + ":" + re.Field()
		if val, ok := fields[fieldName]; ok {
			switch val := val.(type) {
			case string:
				fields[fieldName] = []string{val, re.Description()}
			case []string:
				fields[fieldName] = append(val, re.Description())
			}
		} else {
			fields[fieldName] = re.Description()
		}
	}
	return fields
}

func (cv *CustomValidator) Validate(i any) error {

	params := []ParamIn{
		ParamInPath, ParamInQuery, ParamInHeader,
		ParamInCookie, ParamInBody, ParamInForm,
		ParamInFormData,
	}
	for _, param := range params {
		result, err := validate(i, string(param))
		if err != nil {
			return err
		}
		if !result.Valid() {
			return &ValidatorError{
				http.StatusBadRequest,
				param,
				result.Errors(),
			}
		}
	}

	return nil
}

func validate(i any, param string) (*gojsonschema.Result, error) {
	reflector := jsonschema.Reflector{}
	schema, _ := reflector.Reflect(i, jsonschema.PropertyNameTag(param))
	j, _ := schema.JSONSchemaBytes()
	schemaLoader := gojsonschema.NewStringLoader(string(j))
	documentLoader := gojsonschema.NewGoLoader(structToMap(i, param))
	return gojsonschema.Validate(schemaLoader, documentLoader)
}

func structToMap(item any, tagName string) map[string]any {
	res := make(map[string]any)
	if item == nil {
		return res
	}
	typ := reflect.TypeOf(item)
	val := reflect.ValueOf(item)
	val = reflect.Indirect(val)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Name == "_" {
			continue
		}
		tag := field.Tag.Get(tagName)
		fieldVal := val.Field(i).Interface()
		if field.Anonymous && tag == "" {
			res = mergeMaps(res, structToMap(fieldVal, tagName))
			continue
		}
		if tag != "" && tag != "-" {
			if field.Type.Kind() == reflect.Struct {
				res[tag] = structToMap(fieldVal, tagName)
			} else {
				res[tag] = fieldVal
			}
		}
	}
	return res
}

func mergeMaps(m1 map[string]any, m2 map[string]any) map[string]any {
	merged := make(map[string]any)
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}
