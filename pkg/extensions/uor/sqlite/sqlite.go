package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Element represents a SQLite schema element
type Table struct {
	TableName  string
	ColumnName string
	ColumnType string
}

func JSONSchemaToSQLiteSchema(jsonSchema map[string]interface{}, parent *Table, db *sql.DB, resType string, attach bool) {
	var createdTables []string
	properties, hasProperties := jsonSchema["properties"].(map[string]interface{})

	if !hasProperties {
		fmt.Println("Debug: No properties found")
		return
	}

	resType = urlToUnderscore(resType)
	var e Table

	if parent != nil {
		fmt.Printf("Debug: found parent: %v\n", parent.TableName)
		e.TableName = parent.TableName

	} else {
		fmt.Printf("Debug: no parent found, setting tablename: %v\n", resType)
		e.TableName = resType
		execSQL(db, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY);", e.TableName))
	}

	for key, value := range properties {
		fmt.Printf("Debug: key: %v\n", StripPrefix(key))
		propMap, isMap := value.(map[string]interface{})
		fmt.Printf("Debug: propMap: %v\n", propMap)
		var propsOnly map[string]interface{}

		if isMap {
			typeVal, _ := propMap["type"].(string)
			newTable := Table{TableName: e.TableName + "_" + StripPrefix(key)}

			// Special handling for 'Resource' and 'Location'
			// if the propMap has a 'properties' key, then save the value of that key
			// to propsOnly
			var schemas map[string]interface{}
			var exists bool
			propsOnly, exists = propMap["properties"].(map[string]interface{})
			fmt.Printf("propsOnly: %v\n", propsOnly)
			if exists {
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
				JSONSchemaToSQLiteSchema(propMap, &newTable, db, "", false)

			} else if key == "Resource" {
				if parent == nil && resType == "uor_statementrecord" {
					newTable.TableName = "uor_statementrecord_resource"
				} else {
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
				JSONSchemaToSQLiteSchema(propMap, &newTable, db, "", false)

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
				JSONSchemaToSQLiteSchema(propMap, &newTable, db, "", false)

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
								JSONSchemaToSQLiteSchema(itemMap, &arrayTable, db, "", false)
							} else {
								// Handle primitive types in the array
								primitiveType := typeMap(itemType)
								execSQL(db, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s_%d %s;", newTable.TableName, StripPrefix(key), i, primitiveType))
							}
						}
					}
				}
			default:
				e.ColumnName = StripPrefix(key)
				e.ColumnType = typeMap(typeVal)
				execSQL(db, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", e.TableName, e.ColumnName, e.ColumnType))
			}
		} else {
			e.ColumnName = StripPrefix(key)
			e.ColumnType = typeMap(fmt.Sprintf("%v", value))
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

func urlToUnderscore(url string) string {
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
