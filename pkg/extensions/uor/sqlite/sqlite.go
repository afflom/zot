package sqlite

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/graphql-go/graphql"

	_ "github.com/mattn/go-sqlite3"
)

// Element represents a SQLite schema element
type Table struct {
	TableName  string
	ColumnName string
	ColumnType string
}

// JSONSchemaToSQLiteSchema converts a JSON schema to a SQLite schema while
// building a GraphQL schema from the JSON schema. GraphQL query resolvers are
// generated for each table and column in the SQLite schema and inserted into
// the GraphQL schema.
func JSONSchemaToSQLiteSchema(jsonSchema map[string]interface{}, parent *Table, db *sql.DB, resType string, attach bool, currentObject *graphql.Object) {
	var createdTables []string
	properties, hasProperties := jsonSchema["properties"].(map[string]interface{})

	if !hasProperties {
		fmt.Println("Debug: No properties found")
		return
	}
	// The resType is the name of the logical schema being handled.
	// It is the prefix for the table name. Any resource type that
	// is not uor_statementrecord will be a child of the uor_statement
	// resource and location anchors (see README.md)
	resType = URLToUnderscore(resType)
	var e Table

	// if there is a parent table, then these columns are the added
	// to that table
	if parent != nil {
		fmt.Printf("Debug: found parent: %v\n", parent.TableName)
		e.TableName = parent.TableName

		// if there is no parent, then this is a database initialization or
		// a new logical schema. In either case, the create a new table
		// with the name of the logical schema (resType).
	} else {
		fmt.Printf("Debug: no parent found, setting tablename: %v\n", resType)
		e.TableName = resType

		execSQL(db, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY);", e.TableName))

		if currentObject.Name() != "uor_statementrecord" {
			// Create the corresponding GraphQL object, whose name is the resType
			//src := rand.NewSource(time.Now().UnixNano())
			//r := rand.New(src)

			// Generate a random number between 10000 and 99999
			//randomNumber := r.Intn(90000) + 10000

			//saltedKey := fmt.Sprintf("%s_%d", resType, randomNumber)

			//randomNumberBytes := []byte(fmt.Sprintf("%d", saltedKey))

			// Hash the byte slice using SHA-256
			//hash := sha256.Sum256(randomNumberBytes)

			// Convert the hash to a hexadecimal string
			//hashHex := hex.EncodeToString(hash[:])
			//hashHex = "_" + hashHex

			// Create the root graphql object

			currentObject = CreateObject(resType, graphql.Fields{})
		}

	}

	// iterate over the properties of the json schema to discover the
	// columns of the table
	for key, value := range properties {
		src := rand.NewSource(time.Now().UnixNano())
		r := rand.New(src)

		// Generate a random number between 10000 and 99999
		randomNumber := r.Intn(90000) + 10000

		saltedKey := fmt.Sprintf("%s_%d", key, randomNumber)

		randomNumberBytes := []byte(fmt.Sprintf("%d", saltedKey))

		// Hash the byte slice using SHA-256
		hash := sha256.Sum256(randomNumberBytes)

		// Convert the hash to a hexadecimal string
		hashHex := hex.EncodeToString(hash[:])
		hashHex = "_" + hashHex

		skip := false
		if key == "Location" || key == "Resource" {
			skip = true
		}

		fmt.Printf("Debug: key: %v\n", StripPrefix(key))
		// The propMap is a forecast of the current key's nested object's
		// properties if it has them. It is used to determine if the
		// current object is a UOR Element. If it
		// is, then that child's table will be the start of a new
		// logical schema.
		propMap, isMap := value.(map[string]interface{})
		fmt.Printf("Debug: propMap: %v\n", propMap)
		var propsOnly map[string]interface{}

		// if the current key is a map, then start constructing a new table
		// and continue recursing into the schema.
		if isMap {
			typeVal, _ := propMap["type"].(string)
			fmt.Printf("Debug: typeVal: %v\n", typeVal)
			newTable := Table{TableName: e.TableName + "_" + StripPrefix(key)}

			// Special handling for 'Resource' and 'Location'
			// if the propMap has a 'properties' key, then save the value of that key
			// to propsOnly.
			// TODO: Needs better validation to determine that it is
			// descending into a UOR Element.
			var schemas map[string]interface{}
			var exists bool
			propsOnly, exists = propMap["properties"].(map[string]interface{})
			fmt.Printf("propsOnly: %v\n", propsOnly)
			if exists {
				// if the uor_statement anchors are found, then prepare to
				// handle their logical schemas. These if statements
				// idenfity the element's type information and use it
				// as the prefix of the element's child table name
				// (the root of a new logical schema).
				schemas := make(map[string]interface{})
				if subject, _ := propsOnly["Subject"].(map[string]interface{}); exists {
					fmt.Printf("found subject locator: %v\n", subject["LocatorType"])
					subjectProps, _ := subject["properties"].(map[string]interface{})
					schemas["subjectLocator"], _ = subjectProps["LocatorType"].(string)
					schemas["subjectResource"], _ = subjectProps["ResourceType"].(string)
				}
				if predicate, _ := propsOnly["Predicate"].(map[string]interface{}); exists {
					predicateProps, _ := predicate["properties"].(map[string]interface{})
					fmt.Printf("found predicate locator: %v\n", predicateProps["LocatorType"])
					schemas["predicateLocator"], _ = predicateProps["LocatorType"].(string)
					schemas["predicateResource"], _ = predicateProps["ResourceType"].(string)
				}
				if object, _ := propsOnly["Object"].(map[string]interface{}); exists {
					objectProps, _ := object["properties"].(map[string]interface{})

					fmt.Printf("found object locator: %v\n", objectProps["LocatorType"])
					schemas["objectLocator"], _ = objectProps["LocatorType"].(string)
					schemas["objectResource"], _ = objectProps["ResourceType"].(string)
				}
				fmt.Printf("schemas: %v\n", schemas)

			}

			if key == "Location" {
				// in case the parent column name is "Subject", "Predicate", or "Object"
				// then the new table name should be the value of the locatorType
				if parent == nil && resType == "uor_statementrecord" {
					newTable.TableName = "uor_statementrecord_location"
					// Set the name of the new logical schema
				} else {
					switch parent.ColumnName {
					case "Subject":
						fmt.Printf("processing subject locator: %v\n", schemas["subjectLocator"])
						newTable.TableName = schemas["subjectLocator"].(string)
					case "Predicate":
						fmt.Printf("processing predicate locator: %v\n", schemas["predicateLocator"])
						newTable.TableName = schemas["predicateLocator"].(string)
					case "Object":
						fmt.Printf("processing object locator: %v\n", schemas["objectLocator"])
						newTable.TableName = schemas["objectLocator"].(string)
					}
				}
				// Initialize new logical schema with foreign key to parent
				execSQL(db, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY, fk_%s INTEGER);", newTable.TableName, e.TableName))

				newObject := CreateObject(hashHex, graphql.Fields{})

				currentObject.AddFieldConfig(key, &graphql.Field{
					Name: key,
					Type: newObject,
				})
				JSONSchemaToSQLiteSchema(propMap, &newTable, db, "", false, newObject)

			} else if key == "Resource" {
				if parent == nil && resType == "uor_statementrecord" {
					newTable.TableName = "uor_statementrecord_resource"
				} else {
					// Set the name of the new logical schema
					switch parent.ColumnName {
					case "Subject":
						fmt.Printf("processing subject resource: %v\n", schemas["subjectResource"])
						newTable.TableName = schemas["subjectResource"].(string)
					case "Predicate":
						fmt.Printf("processing predicate resource: %v\n", schemas["predicateResource"])
						newTable.TableName = schemas["predicateResource"].(string)
					case "Object":
						fmt.Printf("processing object resource: %v\n", schemas["objectResource"])
						newTable.TableName = schemas["objectResource"].(string)
					}
				}
				// Initialize new logical schema with foreign key to parent
				execSQL(db, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY, fk_%s INTEGER);", newTable.TableName, e.TableName))

				newObject := CreateObject(hashHex, graphql.Fields{})

				currentObject.AddFieldConfig(key, &graphql.Field{
					Name: key,
					Type: newObject,
				})
				JSONSchemaToSQLiteSchema(propMap, &newTable, db, "", false, newObject)

			}
			var fieldConfig graphql.Field
			if skip {
				continue
			} else {
				fieldConfig = graphql.Field{Name: key, Type: SqliteTypeToGraphqlType(typeVal)}
				fmt.Printf("graphql key: %s, is type: %v\n", key, SqliteTypeToGraphqlType(typeVal))
				currentObject.AddFieldConfig(key, &fieldConfig)

			}

			switch typeVal {
			case "object":
				if attach {
					fmt.Printf("Debug: attaching table: %v\n", newTable.TableName)
					execSQL(db, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY, fk_%s INTEGER, fk_uor_statementrecord_resource_Predicate_Location INTEGER, fk_uor_statementrecord_resource_Predicate_Resource INTEGER, fk_uor_statementrecord_resource_Subject_Location INTEGER, fk_uor_statementrecord_resource_Subject_Resource INTEGER, fk_uor_statementrecord_resource_Object_Location INTEGER, fk_uor_statementrecord_resource_Object_Resource INTEGER);", newTable.TableName, e.TableName))
				} else {
					execSQL(db, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY, fk_%s INTEGER);", newTable.TableName, e.TableName))
				}
				createdTables = append(createdTables, newTable.TableName)

				if skip {
					fmt.Printf("skipping field: %v\n", key)
					continue
				} else {

					newObject := CreateObject(hashHex, graphql.Fields{})

					currentObject.AddFieldConfig(key, &graphql.Field{
						Name: key,
						Type: newObject,
					})
					JSONSchemaToSQLiteSchema(propMap, &newTable, db, "", false, newObject)
				}
			case "array":
				// Handle arrays containing objects or primitives
				items, hasItems := propMap["items"].([]interface{})
				if hasItems {
					for i, item := range items {
						itemMap, isMap := item.(map[string]interface{})
						if isMap {
							itemType, _ := itemMap["type"].(string)
							if itemType == "object" {
								// Create a table for the array items
								arrayTable := Table{TableName: fmt.Sprintf("%s_array_%d", newTable.TableName, i)}
								execSQL(db, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY, fk_%s INTEGER);", arrayTable.TableName, newTable.TableName))
								createdTables = append(createdTables, arrayTable.TableName)

								newObject := CreateObject(hashHex, graphql.Fields{})

								currentObject.AddFieldConfig(key, &graphql.Field{
									Name: key,
									Type: newObject,
								})

								JSONSchemaToSQLiteSchema(itemMap, &arrayTable, db, "", false, newObject)
							} else {
								// Handle primitive types in the array
								primitiveType := typeMap(itemType)

								currentObject.AddFieldConfig(key, &graphql.Field{
									Name: key,
									Type: SqliteTypeToGraphqlType(typeMap(typeVal)),
								})
								fmt.Printf("graphql key: %s, is type: %v\n", key, SqliteTypeToGraphqlType(typeMap(typeVal)))
								execSQL(db, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s_%d %s;", newTable.TableName, StripPrefix(key), i, primitiveType))
							}
						}
					}
				}
			default:
				e.ColumnName = StripPrefix(key)
				e.ColumnType = typeMap(typeVal)
				currentObject.AddFieldConfig(key, &graphql.Field{
					Name: key,
					Type: SqliteTypeToGraphqlType(typeMap(typeVal)),
				})
				fmt.Printf("graphql key: %s, is type: %v\n", key, SqliteTypeToGraphqlType(typeMap(typeVal)))

				execSQL(db, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", e.TableName, e.ColumnName, e.ColumnType))
			}
		} else {
			e.ColumnName = StripPrefix(key)
			e.ColumnType = typeMap(fmt.Sprintf("%v", value))
			if skip {
				fmt.Printf("skipping field: %v\n", key)
				continue
			} else {
				currentObject.AddFieldConfig(key, &graphql.Field{
					Name: key,
					Type: SqliteTypeToGraphqlType(typeMap(fmt.Sprintf("%v", value))),
				})
				fmt.Printf("graphql key: %s, is type: %v\n", key, SqliteTypeToGraphqlType(typeMap(fmt.Sprintf("%v", value))))
			}
			execSQL(db, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", e.TableName, e.ColumnName, e.ColumnType))
		}
	}

	fmt.Println("Debug: Created tables:")
	for _, tableName := range createdTables {
		fmt.Println("  - " + tableName)
	}

}

func typeMap(jsonType string) string {
	switch jsonType {
	case "integer":
		return "INTEGER"
	case "number":
		return "REAL"
	case "string":
		return "TEXT"
	case "boolean":
		return "BOOLEAN"
	case "null":
		return "NULL"
	default:
		return "BLOB"
	}
}

// SqliteTypeToGraphqlType maps SQLite types to GraphQL types
func SqliteTypeToGraphqlType(sqliteType string) graphql.Output {
	switch sqliteType {
	case "INTEGER":
		return graphql.Int
	case "REAL":
		return graphql.Float
	case "TEXT":
		return graphql.String
	case "BOOLEAN":
		return graphql.Boolean
	case "NULL":
		// GraphQL doesn't have a specific type for NULL, any type can be null
		// TODO: handle null correctly
		return graphql.String
	default:
		// GraphQL doesn't have a direct equivalent to SQLite's BLOB type
		// TODO: handle BLOB correctly
		return graphql.String
	}
}
func ExecSQL(db *sql.DB, sqlStmt string) {
	execSQL(db, sqlStmt)
}

func execSQL(db *sql.DB, sqlStmt string) {

	result, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
	lastinsert, errr := result.LastInsertId()
	if errr != nil {
		fmt.Printf("last insert id: %v\n", lastinsert)
	}
	fmt.Printf("result: %v\n", lastinsert)

}

/*
func ExecQuery(db *sql.DB, sqlStmt string) (result sql.Result) {
	db.Query()
	result, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

}*/

// stripPrefix removes the prefix from a key and returns the key without the prefix.
func StripPrefix(key string) string {
	idx := strings.Index(key, ":")
	if idx != -1 {
		return key[idx+1:]
	}
	return key
}

func QueryDynamicSchema(db *sql.DB, jsonDoc map[string]interface{}, parent *Table, schemas map[string]interface{}) (map[string]interface{}, error) {
	var result map[string]interface{} = make(map[string]interface{})

	// Base case: if jsonDoc is empty, return
	if len(jsonDoc) == 0 {
		return result, nil
	}

	// Determine the table to query
	tableName := parent.TableName
	fmt.Printf("Debug: tableName: %v\n", tableName)

	// Handle nested objects
	for key, value := range jsonDoc {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Determine the new table name based on parent and key
			newTable := Table{TableName: tableName + "_" + key}
			nestedResult, err := QueryDynamicSchema(db, nestedMap, &newTable, schemas)
			if err != nil {
				return nil, err
			}
			result[key] = nestedResult
		} else {
			// Construct and execute the query for non-nested fields
			queryStr := fmt.Sprintf("SELECT * FROM %s WHERE %s LIKE ?", tableName, key)
			rows, err := db.Query(queryStr, fmt.Sprintf("%%%s%%", value))
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			// Fetch the rows and construct the result set
			columns, err := rows.Columns()
			if err != nil {
				return nil, err
			}
			for rows.Next() {
				values := make([]interface{}, len(columns))
				scanArgs := make([]interface{}, len(values))
				for i := range values {
					scanArgs[i] = &values[i]
				}
				if err := rows.Scan(scanArgs...); err != nil {
					return nil, err
				}
				rowMap := make(map[string]interface{})
				for i, col := range columns {
					rowMap[col] = values[i]
				}
				result[key] = rowMap
			}
		}
	}
	return result, nil
}

func WriteToDynamicSchema(db *sql.DB, jsonDoc map[string]interface{}, parent *Table, schemas map[string]interface{}) error {
	// Base case: if jsonDoc is empty, return
	if len(jsonDoc) == 0 {
		return nil
	}

	// Determine the table to insert into
	tableName := parent.TableName
	fmt.Printf("Debug: tableName: %v\n", tableName)

	// Prepare the SQL statement for insertion
	columns := []string{}
	placeholders := []string{}
	values := []interface{}{}
	for key, value := range jsonDoc {
		if _, ok := value.(map[string]interface{}); !ok {
			columns = append(columns, key)
			placeholders = append(placeholders, "?")
			values = append(values, value)
		}
	}
	if len(columns) > 0 {
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
		_, err := db.Exec(stmt, values...)
		if err != nil {
			return err
		}
	}

	// Handle nested objects
	for key, value := range jsonDoc {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Determine the new table name based on parent and key
			newTable := Table{TableName: tableName + "_" + key}
			err := WriteToDynamicSchema(db, nestedMap, &newTable, schemas)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// genericResolver is a resolver function that can handle any field and query argument.
func genericResolver(params graphql.ResolveParams) (interface{}, error) {
	// The name of the field being resolved.
	fieldName := params.Info.FieldName

	// The query argument, if provided.
	queryArg, isOk := params.Args["query"]
	if !isOk {
		queryArg = ""
	}

	// For simplicity, we'll just return the field name and query argument as a string.
	// In a real application, you'd use these values to construct and execute a database query.
	return fieldName + ": " + queryArg.(string), nil
}

// QueryDatabase queries an SQLite database based on a JSON-like map.
func QueryDatabase(db *sql.DB, jsonDoc []byte) ([]map[string]interface{}, error) {
	var queryJSON map[string]interface{}
	err := json.Unmarshal(jsonDoc, &queryJSON)
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	// Start the recursive query generation
	query, args := generateQuery(queryJSON, "uor_statementrecord", "")
	fullQuery := "SELECT * FROM uor_statementrecord WHERE " + query
	fmt.Printf("Debug: Generated Full SQL query: %s\n", fullQuery)

	// Execute the query
	rows, err := db.Query(fullQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("debug: Query execution failed: %v", err)
	}
	defer rows.Close()

	// Fetch the results
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("debug: Fetching columns failed: %v", err)
	}

	for rows.Next() {
		row := make([]interface{}, len(columns))
		rowPtrs := make([]interface{}, len(columns))
		for i := 0; i < len(columns); i++ {
			rowPtrs[i] = &row[i]
		}

		if err := rows.Scan(rowPtrs...); err != nil {
			return nil, fmt.Errorf("debug: Scanning row failed: %v", err)
		}

		result := make(map[string]interface{})
		for i, col := range columns {
			result[col] = row[i]
		}
		results = append(results, result)
	}

	return results, nil
}

// generateQuery recursively generates SQL query based on JSON document.
func generateQuery(jsonMap map[string]interface{}, tableName string, parentKey string) (string, []interface{}) {
	var queryParts []string
	var args []interface{}

	for key, value := range jsonMap {
		// Handle nested objects
		if nestedMap, ok := value.(map[string]interface{}); ok {
			newTableName := tableName + "_" + key // Default nested table name

			// Special handling for locatorType and resourceType
			if key == "location" || key == "resource" {
				if locator, exists := nestedMap["locatorType"]; exists {
					newTableName = locator.(string)
				} else if resource, exists := nestedMap["resourceType"]; exists {
					newTableName = resource.(string)
				}
			}

			nestedQuery, nestedArgs := generateQuery(nestedMap, newTableName, key)
			queryParts = append(queryParts, nestedQuery)
			args = append(args, nestedArgs...)
			continue
		}

		// Handle primitive types
		queryParts = append(queryParts, fmt.Sprintf("%s.%s = ?", tableName, key))
		args = append(args, value)
	}

	// Combine all query parts
	query := strings.Join(queryParts, " AND ")
	if parentKey != "" {
		query = fmt.Sprintf("(%s)", query)
	}

	return query, args
}

func URLToUnderscore(url string) string {
	// Replace each character that is not allowed in a SQL table name with an underscore
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, url)
	return result
}

func DumpSchema(db *sql.DB) {
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table';`)
	if err != nil {
		log.Fatalf("Failed to query for table names: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("Failed to scan table name: %v", err)
		}

		fmt.Printf("-- Schema for table: %s\n", tableName)

		schemaRows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%s);`, tableName))
		if err != nil {
			log.Fatalf("Failed to get table schema: %v", err)
		}

		for schemaRows.Next() {
			var cid, name, fieldType string
			var notnull, pk int
			var dfltValue interface{}

			err := schemaRows.Scan(&cid, &name, &fieldType, &notnull, &dfltValue, &pk)
			if err != nil {
				log.Fatalf("Failed to scan schema info: %v", err)
			}

			fmt.Printf("Column ID: %s, Name: %s, Type: %s, NotNull: %d, DefaultValue: %v, IsPrimaryKey: %d\n", cid, name, fieldType, notnull, dfltValue, pk)
		}

		schemaRows.Close()
		fmt.Println("-- End of schema")
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Row scanning error: %v", err)
	}
}

func CreateObject(name string, fields graphql.Fields) *graphql.Object {
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   name,
			Fields: fields,
		},
	)
}

// CreateField creates a new graphql.Field with the specified name, type, and resolver.
func CreateField(name string, gType graphql.Output, resolve graphql.FieldResolveFn) *graphql.Field {
	return &graphql.Field{
		Type:        gType,
		Description: name,
		Resolve:     resolve,
	}
}
