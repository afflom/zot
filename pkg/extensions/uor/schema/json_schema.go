package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func GenerateJSONSchema(t reflect.Type, root bool) map[string]interface{} {
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
			fieldSchema = GenerateJSONSchema(fieldType, false)
		case reflect.Map:
			fieldSchema["type"] = "object"
		default:
			fieldSchema["type"] = "null"
		}

		properties[field.Name] = fieldSchema
	}

	schema["type"] = "object"
	schema["properties"] = properties

	// marshal schema as json and print to the terminal
	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("schema: %s\n", b)

	return schema
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
