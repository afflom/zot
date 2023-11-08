package graphql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GenerateResolver(tableName string, db *sql.DB) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		columnName := p.Info.FieldName
		var queryStr string
		if p.Args[columnName] == nil {
			queryStr = fmt.Sprintf("SELECT %s FROM %s", columnName, tableName)
		} else {
			queryStr = fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", columnName, tableName, p.Args[columnName])

		}
		rows, err := db.Query(queryStr, fmt.Sprintf("%%%s%%", p.Args[columnName]))
		if err != nil {
			return nil, err
		}
		return rows, nil
	}

}

// printType recursively prints the details of a GraphQL type
func printType(t graphql.Type, indent string) {
	switch typ := t.(type) {
	case *graphql.Object:
		fmt.Printf("%sObject: %s\n", indent, typ.Name())
		PrintFieldDefinitions(typ.Fields(), indent+"  ")
	case *graphql.List:
		fmt.Printf("%sList: \n", indent)
		printType(typ.OfType, indent+"  ")
	case *graphql.NonNull:
		fmt.Printf("%sNonNull: \n", indent)
		printType(typ.OfType, indent+"  ")
	default:
		fmt.Printf("%sType: %v\n", indent, typ)
	}
}

// printFieldDefinitions iterates through the field definitions of an object, printing each one and
// recursively printing the fields of any nested objects
func PrintFieldDefinitions(fields graphql.FieldDefinitionMap, indent string) {
	for fieldName, field := range fields {
		fmt.Printf("%sField: %s, Type: %v\n", indent, fieldName, field.Type)
		printType(field.Type, indent+"  ")
	}
}

// FindNestedObject is a receiver to a graphql-go.Object and
// takes a json path reference to a target graphql-go.Object
// FindNestedObject returns a graphql-go.Object and an error.
func FindNestedObject(obj *graphql.Object, jsonPath string) (*graphql.Object, error) {
	// Split the JSON path into its components
	components := strings.Split(jsonPath, ".")

	// Begin the recursive search
	return searchNestedObject(obj, components)
}

func searchNestedObject(obj *graphql.Object, components []string) (*graphql.Object, error) {
	// Base case: if there are no more components, we've reached the target object
	if len(components) == 0 {
		return obj, nil
	}

	// Get the field corresponding to the next component in the path
	field := obj.Fields()[components[0]]
	if field == nil {
		return nil, errors.New("field not found")
	}

	// Check if the field is an object type
	nestedObj, ok := field.Type.(*graphql.Object)
	if !ok {
		return nil, errors.New("field is not an object type")
	}

	// Recurse with the remaining components of the path
	return searchNestedObject(nestedObj, components[1:])
}

func ConvertToGraphQL(jsonSchema map[string]interface{}, name string, database mongo.Database) (*graphql.Object, error) {

	// Navigate to the desired definition within $defs
	defs, ok := jsonSchema["$defs"].(map[string]interface{})
	if !ok {
		return nil, errors.New("$defs is not a map[string]interface{}")
	}

	statementRecord, ok := defs["StatementRecord"].(map[string]interface{})
	if !ok {
		return nil, errors.New("StatementRecord not found in $defs")
	}

	properties := statementRecord["properties"]
	fields, err := generateFields(properties, name, defs, database)
	if err != nil {
		log.Printf("Error generating fields: %v\n", err)
		return nil, err
	}

	// Check if fields is empty and add a dummy field if necessary
	if len(fields) == 0 {
		fields["_empty"] = &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "no data", nil
			},
		}
	}

	return graphql.NewObject(graphql.ObjectConfig{
		Name:   name,
		Fields: fields,
	}), nil
}

func convertJsonToGraphQLType(jsonSchema interface{}, fieldName string, defs map[string]interface{}, database mongo.Database) (graphql.Output, error) {
	propSchema, ok := jsonSchema.(map[string]interface{})
	if !ok {
		return nil, errors.New("property is not a json schema")
	}

	if ref, ok := propSchema["$ref"].(string); ok {
		refName := strings.TrimPrefix(ref, "#/$defs/")
		def, ok := defs[refName]
		if !ok {
			return nil, fmt.Errorf("definition %s not found in $defs", refName)
		}
		return convertJsonToGraphQLType(def, fieldName, defs, database)
	}

	typeOfSchema, ok := propSchema["type"].(string)
	if !ok {
		return nil, errors.New("json schema does not have 'type' field")
	}

	switch typeOfSchema {
	case "string":
		return graphql.String, nil
	case "integer":
		return graphql.Int, nil
	case "float":
		return graphql.Float, nil
	case "boolean":
		return graphql.Boolean, nil
	case "object":
		additionalProps, ok := propSchema["additionalProperties"].(map[string]interface{})
		if ok {
			additionalType, err := convertJsonToGraphQLType(additionalProps, fieldName+"_additional", defs, database)
			if err != nil {
				return nil, err
			}
			mapEntryType := graphql.NewObject(graphql.ObjectConfig{
				Name: fieldName + "_MapEntry",
				Fields: graphql.Fields{
					"key":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
					"value": &graphql.Field{Type: additionalType},
				},
			})
			return graphql.NewList(mapEntryType), nil
		} else if properties, ok := propSchema["properties"].(map[string]interface{}); ok {
			fields, err := generateFields(properties, fieldName, defs, database) // recursively generate fields for nested object
			if err != nil {
				return nil, err
			}
			return graphql.NewObject(graphql.ObjectConfig{
				Name:   fieldName,
				Fields: fields,
			}), nil
		} else {
			// Default handling for objects without properties or additionalProperties
			return graphql.NewObject(graphql.ObjectConfig{
				Name: fieldName,
				Fields: graphql.Fields{
					"_empty": &graphql.Field{Type: graphql.String},
				},
			}), nil
		}

	case "array":
		items, ok := propSchema["items"].(map[string]interface{})
		if !ok {
			return nil, errors.New("json schema for array does not have 'items' field")
		}
		itemType, err := convertJsonToGraphQLType(items, fieldName+"_item", defs, database)
		if err != nil {
			return nil, err
		}
		return graphql.NewList(itemType), nil
	default:
		return nil, fmt.Errorf("unsupported json schema type %s", typeOfSchema)
	}
}

func sanitizeName(name string) string {
	re := regexp.MustCompile(`[^_a-zA-Z0-9]`)
	return re.ReplaceAllString(name, "_")
}

// Helper function to determine if a field is a nested field
func isNestedField(propValue interface{}, defs map[string]interface{}) bool {
	if _, isMap := propValue.(map[string]interface{}); isMap {
		// This is a nested object, should be handled by GraphQL resolver
		return true
	}
	if ref, isRef := propValue.(map[string]interface{})["$ref"]; isRef {
		// If it's a reference, check if the referenced definition is an object
		defKey := strings.TrimPrefix(ref.(string), "#/$defs/")
		if def, exists := defs[defKey]; exists {
			_, isObject := def.(map[string]interface{})
			return isObject
		}
	}
	// If it's an array, we need to check if its items are objects that require nested resolution
	if array, isArray := propValue.([]interface{}); isArray {
		for _, item := range array {
			if isNestedField(item, defs) {
				return true
			}
		}
	}
	return false
}

func generateNestedFields(database mongo.Database, propName string, nestedProperties map[string]interface{}, defs map[string]interface{}) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		// Recursively process nested fields
		var processNestedFields func(string, interface{}) (interface{}, error)
		processNestedFields = func(parentPropName string, propValue interface{}) (interface{}, error) {
			// Handle arrays of primitives or objects
			if propArray, isArray := propValue.([]interface{}); isArray {
				var arrayResponse []interface{}
				for _, item := range propArray {
					if nestedProps, isNested := item.(map[string]interface{}); isNested && isNestedField(item, defs) {
						itemResponse, err := processNestedFields(parentPropName, nestedProps)
						if err != nil {
							return nil, err
						}
						arrayResponse = append(arrayResponse, itemResponse)
					} else {
						arrayResponse = append(arrayResponse, item)
					}
				}
				return arrayResponse, nil
			}

			// Process nested objects
			if propMap, isMap := propValue.(map[string]interface{}); isMap {
				var nestedResponse map[string]interface{} = make(map[string]interface{})
				for nestedPropName, nestedPropValue := range propMap {
					fullPropName := parentPropName + "." + nestedPropName
					sanitizedPropName := sanitizeName(fullPropName)

					if isNestedField(nestedPropValue, defs) {
						// Recursively process nested fields
						deeperNestedResponse, err := processNestedFields(sanitizedPropName, nestedPropValue)
						if err != nil {
							return nil, err
						}
						nestedResponse[sanitizedPropName] = deeperNestedResponse
					} else {
						// For non-nested fields, perform database operations
						filter, err := generateFilter(p.Args, sanitizedPropName)
						if err != nil {
							return nil, err
						}

						projection := bson.D{{Key: sanitizedPropName, Value: 1}}

						cursor, err := findDocuments(database, "statements", filter, projection, p)
						if err != nil {
							return nil, err
						}

						var results []map[string]interface{}
						if err = cursor.All(p.Context, &results); err != nil {
							return nil, err
						}

						extractedValues, err := extractFieldValues(results, sanitizedPropName)
						if err != nil {
							return nil, err
						}

						nestedResponse[sanitizedPropName] = extractedValues
					}
				}
				return nestedResponse, nil
			}

			// Non-object, non-array properties can be returned as is
			return propValue, nil
		}

		// Begin processing with the top-level properties
		topLevelResponse, err := processNestedFields(propName, nestedProperties)
		if err != nil {
			return nil, err
		}

		return topLevelResponse, nil
	}
}

// Helper function to retrieve collection and execute the Find operation
func findDocuments(database mongo.Database, collectionName string, filter, projection bson.D, p graphql.ResolveParams) (*mongo.Cursor, error) {
	// Obtain a reference to the collection
	collection := database.Collection(collectionName)

	// Set up find options with projection
	findOptions := options.Find().SetProjection(projection)
	//fmt.Printf("Find options: %+v\n", findOptions)

	// Execute the Find operation with projection
	cursor, err := collection.Find(p.Context, filter, findOptions)
	if err != nil {
		log.Printf("Error executing find operation: %v\n", err)
		return nil, err
	}
	return cursor, nil
}

// Helper function to extract field values from documents
func extractFieldValues(results []map[string]interface{}, propName string) ([]interface{}, error) {
	var response []interface{} // Change type to interface{} to handle different data types
	for _, result := range results {
		// Handle nested fields using dot notation
		value, err := getNestedFieldValue(result, propName)
		if err != nil {
			log.Printf("Error getting nested field value for field: %s, error: %v\n", propName, err)
			return nil, err // Consider returning an error if the field should always exist
		}

		// Append the value to response
		response = append(response, value)
	}
	return response, nil
}

func getNestedFieldValue(doc map[string]interface{}, fieldPath string) (interface{}, error) {
	current := doc
	keys := strings.Split(fieldPath, ".")
	pathSoFar := ""

	for i, key := range keys {
		pathSoFar = appendPath(pathSoFar, key) // Helper function to build the error path

		if value, exists := current[key]; exists {
			if i == len(keys)-1 {
				return value, nil
			}
			if next, ok := value.(map[string]interface{}); ok {
				current = next
			} else {
				return nil, fmt.Errorf("expected a nested object at path '%s', but got a different type", pathSoFar)
			}
		} else {
			return nil, fmt.Errorf("field does not exist at path '%s'", pathSoFar)
		}
	}

	return nil, fmt.Errorf("field does not exist at path '%s'", fieldPath)
}

// appendPath helps build the error message with the full path traversed so far
func appendPath(basePath, addition string) string {
	if basePath == "" {
		return addition
	}
	return basePath + "." + addition
}

func generateFields(properties interface{}, parentName string, defs interface{}, database mongo.Database) (graphql.Fields, error) {
	// Attempt to cast properties to a map[string]interface{}
	propMap, ok := properties.(map[string]interface{})
	if !ok {
		return nil, errors.New("properties is not a map[string]interface{}")
	}

	// Attempt to cast defs to a map[string]interface{}
	defMap, ok := defs.(map[string]interface{})
	if !ok {
		return nil, errors.New("$defs is not a map[string]interface{}")
	}

	// Initialize an empty fields map
	fields := graphql.Fields{}

	// Iterate through each property in propMap
	for propName, propValue := range propMap {
		sanitizedPropName := sanitizeName(propName)
		fieldName := fmt.Sprintf("%s_%s", parentName, sanitizedPropName)

		// Convert JSON to GraphQL Type
		baseType, err := convertJsonToGraphQLType(propValue, fieldName, defMap, database)
		if err != nil {
			log.Printf("Error converting JSON to GraphQL Type: %v\n", err)
			return nil, err
		}

		// Wrap the baseType with graphql.NewList to ensure this field is a list of baseType
		fieldType := graphql.NewList(baseType)

		fields[sanitizedPropName] = &graphql.Field{
			Type: fieldType,
			Args: graphql.FieldConfigArgument{
				"search": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			// Resolve function tailored to handle both nested fields and list types
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				// If the field is a nested object or a list of nested objects, use a specialized handler
				if isNestedField(propValue, defMap) {
					nestedProperties, ok := propValue.(map[string]interface{})
					if !ok {
						// If the nested field is not a map, then it might be an array of objects or primitives
						if propArray, isArray := propValue.([]interface{}); isArray {
							// Process each item in the array
							var arrayResponse []interface{}
							for _, item := range propArray {
								// If the item is a nested object, process it accordingly
								if nestedProps, isNested := item.(map[string]interface{}); isNested {
									nestedResponse, err := generateNestedFields(database, sanitizedPropName, nestedProps, defMap)(p)
									if err != nil {
										return nil, err
									}
									arrayResponse = append(arrayResponse, nestedResponse)
								} else {
									// If the item is not an object, append it directly (e.g., string in a list of strings)
									arrayResponse = append(arrayResponse, item)
								}
							}
							return arrayResponse, nil
						} else {
							// If it's not an array or map, return an error as the structure is unexpected
							return nil, fmt.Errorf("expected a nested object or array for field '%s', but got %T", sanitizedPropName, propValue)
						}
					}
					// If it's a nested map, process it with the generateNestedFields function
					return generateNestedFields(database, sanitizedPropName, nestedProperties, defMap)(p)
				}

				// For non-nested fields, generate a filter and query the database
				filter, err := generateFilter(p.Args, sanitizedPropName)
				if err != nil {
					log.Printf("Error generating filter: %v\n", err)
					return nil, err
				}

				projection := bson.D{{Key: sanitizedPropName, Value: 1}}
				cursor, err := findDocuments(database, "statements", filter, projection, p)
				if err != nil {
					return nil, err
				}

				var results []map[string]interface{}
				if err = cursor.All(p.Context, &results); err != nil {
					log.Printf("Error retrieving matching documents: %v\n", err)
					return nil, err
				}

				if len(results) == 0 {
					log.Printf("No matching documents found for field: %s\n", sanitizedPropName)
					return nil, errors.New("no matching documents found")
				}

				response, err := extractFieldValues(results, sanitizedPropName)
				if err != nil {
					return nil, err
				}

				// The response is already []interface{}, so it can be returned directly.
				return response, nil
			}}
	}
	return fields, nil
}

func generateFilter(args map[string]interface{}, propName string) (bson.D, error) {
	// Initialize an array of filters for `$and` operation
	var andFilters []bson.E

	// Check if the field exists
	andFilters = append(andFilters, bson.E{Key: propName, Value: bson.D{{Key: "$exists", Value: true}}})

	// Process each argument
	for argKey, argValue := range args {
		// Assert that the argument value is a string
		regexValue, ok := argValue.(string)
		if !ok {
			// Handle non-string argument, could be continue, return error, etc.
			log.Printf("Non-string argument provided for %s, value: %+v\n", argKey, argValue)
			continue
		}

		// Create a regex filter element for the propName field
		regexFilter := bson.D{{Key: "$regex", Value: regexValue}, {Key: "$options", Value: "i"}} // "i" for case-insensitive
		filterElement := bson.E{Key: propName, Value: regexFilter}

		// Append the filter element to the `$and` array
		andFilters = append(andFilters, filterElement)
	}

	// Combine filters with `$and` if there are multiple conditions
	if len(andFilters) > 1 {
		return bson.D{{Key: "$and", Value: andFilters}}, nil
	} else if len(andFilters) == 1 {
		// If there's only one filter, return it directly
		return bson.D{andFilters[0]}, nil
	}

	// Default to an empty filter if there are no conditions
	return bson.D{}, nil
}
