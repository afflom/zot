package schema

import (
	"fmt"
	"reflect"
)

func GenerateJSONSchema(t reflect.Type) map[string]interface{} {
	schema := map[string]interface{}{}

	switch t.Kind() {
	case reflect.String:
		schema["type"] = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema["type"] = "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"
	case reflect.Bool:
		schema["type"] = "boolean"
	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"
	case reflect.Array, reflect.Slice:
		schema["type"] = "array"
		schema["items"] = GenerateJSONSchema(t.Elem())
	case reflect.Map:
		schema["type"] = "object"
		schema["additionalProperties"] = GenerateJSONSchema(t.Elem())
	case reflect.Struct:
		schema["type"] = "object"
		properties := map[string]interface{}{}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			properties[field.Name] = GenerateJSONSchema(field.Type)
		}
		schema["properties"] = properties
	case reflect.Ptr:
		// Dereference pointer to get the underlying type
		return GenerateJSONSchema(t.Elem())
	default:
		panic(fmt.Sprintf("unsupported type: %v", t))
	}

	return schema
}

// Helper function to map Go types to JSON Schema types
func typeToJSONSchemaType(kind reflect.Kind) string {
	switch kind {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int64:
		return "integer"
	case reflect.Bool:
		return "boolean"
	case reflect.Float64:
		return "number"
	default:
		return "null"
	}
}

func JSONLDToJSONSchema(jsonld map[string]interface{}) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}

	var processNested func(map[string]interface{}) map[string]interface{}

	processNested = func(jsonld map[string]interface{}) map[string]interface{} {
		props := make(map[string]interface{})

		for key, value := range jsonld {
			if key == "@context" || key == "@type" || key == "@id" {
				continue
			}

			fieldType := "string"

			switch v := value.(type) {
			case map[string]interface{}:
				fieldType = "object"
				nestedProps := processNested(v)
				props[key] = map[string]interface{}{
					"type":       fieldType,
					"properties": nestedProps,
				}

			case string:
				fieldType = v
				props[key] = map[string]interface{}{
					"type": fieldType,
				}

			default:
				props[key] = map[string]interface{}{
					"type": "string",
				}
			}
		}

		return props
	}

	schema["properties"] = processNested(jsonld)
	return schema
}
