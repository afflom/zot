package search

import (
	"reflect"
)

func generateJSONSchema(t reflect.Type, root bool) map[string]interface{} {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := make(map[string]interface{})

	if root {
		schema["$schema"] = "http://json-schema.org/draft-07/schema#"
	}

	properties := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		fieldSchema := make(map[string]interface{})

		switch fieldType.Kind() {
		case reflect.String:
			fieldSchema["type"] = "string"
		case reflect.Int, reflect.Int64:
			fieldSchema["type"] = "integer"
		case reflect.Bool:
			fieldSchema["type"] = "boolean"
		case reflect.Float64:
			fieldSchema["type"] = "number"
		case reflect.Struct:
			fieldSchema = generateJSONSchema(fieldType, false)
		case reflect.Map:
			fieldSchema["type"] = "object"
		default:
			fieldSchema["type"] = "null"
		}

		properties[field.Name] = fieldSchema
	}

	schema["type"] = "object"
	schema["properties"] = properties

	return schema
}
