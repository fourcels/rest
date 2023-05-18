package rest

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	gojsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/swaggest/jsonschema-go"
)

type CustomValidator struct{}

type ValidatorError struct {
	HTTPCodeAsError
	paramIn         ParamIn
	ValidationError *gojsonschema.ValidationError
}

func (*ValidatorError) Error() string {
	return "Validation Error"
}

func (ve *ValidatorError) Fields() map[string]any {
	fields := make(map[string]any)
	for _, re := range ve.ValidationError.Causes {
		fieldName := string(ve.paramIn) + ":" + strings.TrimLeft(re.InstanceLocation, "/")
		if val, ok := fields[fieldName]; ok {
			switch val := val.(type) {
			case string:
				fields[fieldName] = []string{val, re.Message}
			case []string:
				fields[fieldName] = append(val, re.Message)
			}
		} else {
			fields[fieldName] = re.Message
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
		if err := validate(i, string(param)); err != nil {
			return &ValidatorError{
				http.StatusBadRequest,
				param,
				err.(*gojsonschema.ValidationError),
			}
		}
	}

	return nil
}

func validate(i any, param string) error {
	reflector := jsonschema.Reflector{}
	schema, _ := reflector.Reflect(i, jsonschema.PropertyNameTag(param))
	j, _ := schema.JSONSchemaBytes()
	schemaLoader := gojsonschema.MustCompileString("schema.json", string(j))
	v, err := decodeMap(structToMap(i, param))
	if err != nil {
		return err
	}
	return schemaLoader.Validate(v)
}

func decodeMap(i map[string]any) (map[string]any, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	var v map[string]any
	json.Unmarshal(b, &v)
	return v, nil
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
		structField := val.Field(i)
		if field.Name == "_" {
			continue
		}
		if !structField.CanInterface() {
			continue
		}
		tag := field.Tag.Get(tagName)
		fieldVal := structField.Interface()
		if field.Anonymous && tag == "" {
			res = mergeMaps(res, structToMap(fieldVal, tagName))
			continue
		}
		if tag != "" && tag != "-" {
			res[tag] = fieldVal
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
