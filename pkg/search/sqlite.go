package search

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	sschema "zotregistry.io/zot/pkg/search/schema"
)

// Element represents a SQLite schema element
type Table struct {
	TableName  string
	ColumnName string
	ColumnType string
}

// JSONSchemaToSQLiteSchema converts a JSON schema to SQLite schema
func JSONSchemaToSQLiteSchema(jsonSchema map[string]interface{}, parent *Table, db *sql.DB, resType string) {
	var createdTables []string
	properties, hasProperties := jsonSchema["properties"].(map[string]interface{})

	if !hasProperties {
		fmt.Println("Debug: No properties found")
		return
	}

	resType = urlToUnderscore(resType)
	var e Table

	if parent != nil {
		e.TableName = parent.TableName
	} else {
		e.TableName = resType
		execSQL(db, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY);", e.TableName))
	}

	for key, value := range properties {
		propMap, isMap := value.(map[string]interface{})
		if isMap {
			typeVal, _ := propMap["type"].(string)
			switch typeVal {
			case "object":
				newTable := Table{TableName: e.TableName + "_" + key}
				execSQL(db, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY, fk_%s INTEGER);", newTable.TableName, e.TableName))
				createdTables = append(createdTables, newTable.TableName)
				fmt.Printf("Debug: Handling nested object for table %s\n", newTable.TableName)

				// Special handling for locatorType and resourceType
				if key == "location" || key == "resource" {
					if locator, exists := propMap["locatorType"]; exists {
						newTable.TableName = locator.(string)
					} else if resource, exists := propMap["resourceType"]; exists {
						newTable.TableName = resource.(string)
					}
				}

				JSONSchemaToSQLiteSchema(propMap, &newTable, db, "")
			default:
				e.ColumnName = key
				e.ColumnType = typeMap(typeVal)
				fmt.Printf("Debug: Adding column %s of type %s to table %s\n", e.ColumnName, e.ColumnType, e.TableName)
				execSQL(db, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", e.TableName, e.ColumnName, e.ColumnType))
			}
		} else {
			e.ColumnName = key
			e.ColumnType = typeMap(fmt.Sprintf("%v", value))
			fmt.Printf("Debug: Adding column %s of type %s to table %s\n", e.ColumnName, e.ColumnType, e.TableName)
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
	case "string":
		return "TEXT"
	case "boolean":
		return "BOOLEAN"
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

func InsertJSONToSQLite(jsonData map[string]interface{}, parentTable string, parentID int64, db *sql.DB, resType string) {
	for key, value := range jsonData {
		var table string
		if parentTable == "uor_statementrecord_statement_predicate" || parentTable == "uor_statementrecord_statement_subject" || parentTable == "uor_statementrecord_statement_object" || parentTable == "uor_statementrecord_reference" {
			table = resType + "_" + key
		} else {
			if parentTable != "" {
				table = parentTable + "_" + key
			} else {
				table = "uor_statementrecord"
			}
		}

		switch v := value.(type) {
		case map[string]interface{}:
			fmt.Println("map detected")
			fmt.Printf("key: %v\nvalue: %v\n", key, v)
			if parentTable != "" {
				res, err := db.Exec(fmt.Sprintf("INSERT INTO %s (fk_%s) VALUES (?)", table, parentTable), parentID)
				if err != nil {
					log.Fatal(err)
				}
				newID, err := res.LastInsertId()
				if err != nil {
					log.Fatal(err)
				}
				// Recursively insert nested JSON
				InsertJSONToSQLite(v, table, newID, db, resType)
			} else {
				// Handle the root table differently
				res, err := db.Exec(fmt.Sprintf("INSERT INTO %s DEFAULT VALUES", table))
				if err != nil {
					log.Fatal(err)
				}
				newID, err := res.LastInsertId()
				if err != nil {
					log.Fatal(err)
				}
				// Recursively insert nested JSON
				InsertJSONToSQLite(v, table, newID, db, resType)
			}
		case float64, string, bool:
			fmt.Println("primitive detected")
			fmt.Printf("key: %v\nvalue: %v\n", key, v)
			column := key
			_, err := db.Exec(fmt.Sprintf("UPDATE %s SET %s = ? WHERE id = ?", table, column), value, parentID)
			if err != nil {
				log.Fatal(err)
			}
			if key == "resourceType" {
				resType = value.(string)
				resType = urlToUnderscore(resType)
			}
		default:
			fmt.Println("Unsupported type")
		}
	}
}

// QueryExtendedDatabase queries an SQLite database based on a JSON-like map.
func QueryDatabase(db *sql.DB, jsonDoc []byte) ([]map[string]interface{}, error) {
	var queryJSON map[string]interface{}
	err := json.Unmarshal(jsonDoc, &queryJSON)
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	// Start the recursive query generation
	query, args := generateQuery(queryJSON, "root", "")
	fmt.Printf("Debug: Generated SQL query: %s\n", query)

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("Debug: Query execution failed: %v", err)
	}
	defer rows.Close()

	// Fetch the results
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("Debug: Fetching columns failed: %v", err)
	}

	for rows.Next() {
		row := make([]interface{}, len(columns))
		rowPtrs := make([]interface{}, len(columns))
		for i := 0; i < len(columns); i++ {
			rowPtrs[i] = &row[i]
		}

		if err := rows.Scan(rowPtrs...); err != nil {
			return nil, fmt.Errorf("Debug: Scanning row failed: %v", err)
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
			nestedQuery, nestedArgs := generateQuery(nestedMap, tableName+"_"+key, key)
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

func queryPrimitiveType(db *sql.DB, tableName string, value interface{}, parentID int64) (interface{}, error) {
	queryStr := fmt.Sprintf("SELECT * FROM %s WHERE id = ? AND fk_%s = ?", tableName, parentID)
	row := db.QueryRow(queryStr, value, parentID)
	var result interface{}
	err := row.Scan(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func checkTableExists(db *sql.DB, tableName string) bool {
	queryStr := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s';", tableName)
	row := db.QueryRow(queryStr)
	var result string
	err := row.Scan(&result)
	return err == nil && result == tableName
}

func sanitizeKey(key string) string {
	return regexp.MustCompile(`\W`).ReplaceAllString(key, "_")
}

func singleTableQuery(db *sql.DB, tableName string, queryJSON map[string]interface{}, parentTable string, locatorType string, resourceType string) (map[string]interface{}, error) {
	fmt.Println("singleTableQuery called")
	var placeholders []string
	var values []interface{}
	for k, v := range queryJSON {
		if k == "resourceType" {
			resourceType = sanitizeKey(v.(string))
			if resourceType == "oci" {
				resourceType = "oci_descriptor"
			}
			fmt.Printf("found resourceType: %v\n", resourceType)
		}
		if k == "locatorType" {
			locatorType = sanitizeKey(v.(string))
			if locatorType == "oci" {
				locatorType = "oci_namespace"
			}
			fmt.Printf("found locatorType: %v\n", locatorType)
		}
		fmt.Printf("parentTable: %v\n", parentTable)
		sanitizedKey := sanitizeKey(k)
		fullKey := sanitizedKey
		newTableName := tableName
		if parentTable != "" {
			fullKey = parentTable + "_" + sanitizedKey
		}
		switch v := v.(type) {
		case string, int, int64, float64, bool:
			placeholders = append(placeholders, fmt.Sprintf("%s = ?", fullKey))
			values = append(values, v)
		case map[string]interface{}:
			if k == "location" {
				newTableName = locatorType

			} else if k == "resource" {
				newTableName = resourceType
			}
			rowData, err := singleTableQuery(db, newTableName, v, k, locatorType, resourceType)
			if err != nil {
				return nil, err
			}
			for k, v := range rowData {
				placeholders = append(placeholders, fmt.Sprintf("%s = ?", k))
				values = append(values, v)
			}
		default:
			continue
		}
	}
	if len(placeholders) == 0 {
		return nil, fmt.Errorf("no valid query fields")
	}

	queryStr := fmt.Sprintf(
		"SELECT * FROM %s WHERE %s",
		tableName,
		strings.Join(placeholders, " AND "),
	)
	rows, err := db.Query(queryStr, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		ptrs[i] = &vals[i]
	}

	rowData := make(map[string]interface{})
	for rows.Next() {
		err = rows.Scan(ptrs...)
		if err != nil {
			return nil, err
		}
		for i, col := range cols {
			rowData[col] = vals[i]
		}
	}
	return rowData, nil
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

func CreateStatementRecordTables(db *sql.DB) {
	fmt.Println("CreateStatementRecordTables called")
	// Convert the statement to a JSON schema
	statementSchema := generateJSONSchema(reflect.TypeOf(sschema.Statement{}), true)
	fmt.Printf("statementSchema: %v\n", statementSchema)
	// Initialize the database with the statementSchema
	JSONSchemaToSQLiteSchema(statementSchema, nil, db, "uor_statementrecord")
	ociSchema := generateJSONSchema(reflect.TypeOf(sschema.UORDescriptor{}), true)
	fmt.Printf("ociSchema: %v\n", ociSchema)
	uorSchema := Table{
		TableName: "uor_statementrecord",
	}
	JSONSchemaToSQLiteSchema(ociSchema, &uorSchema, db, "uor_descriptor")
	oci_namespace := generateJSONSchema(reflect.TypeOf(sschema.Location{}), true)
	fmt.Printf("oci_namespace: %v\n", oci_namespace)
	JSONSchemaToSQLiteSchema(oci_namespace, &uorSchema, db, "oci_namespace")
}

func createTableFromStruct(db *sql.DB, tableName string, t reflect.Type, parentTableName string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fmt.Printf("createTableFromStruct called with tableName: %v, t: %v, parentTableName: %v\n", tableName, t, parentTableName)

	sqlStmt := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY);", tableName)
	fmt.Printf("sqlStmt: %v\n", sqlStmt)
	execSQL(db, sqlStmt)

	if parentTableName != "" {
		execSQL(db, fmt.Sprintf("ALTER TABLE %s ADD COLUMN fk_%s INTEGER;", tableName, parentTableName))
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			// Handle nested struct
			newTableName := fmt.Sprintf("%s_%s", tableName, strings.ToLower(field.Name))
			newTableName = sanitizeKey(newTableName)
			createTableFromStruct(db, newTableName, fieldType, tableName)
		} else {
			sqlType := mapGoTypeToSQL(fieldType.Name())
			if sqlType != "" {
				field.Name = sanitizeKey(field.Name)
				execSQL(db, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", tableName, field.Name, sqlType))
			}
		}
	}
}

func mapGoTypeToSQL(goType string) string {
	switch goType {
	case "string":
		return "TEXT"
	case "int":
		return "INTEGER"
	case "bool":
		return "BOOLEAN"
	default:
		return ""
	}
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
		fmt.Println("-- End of schema\n")
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Row scanning error: %v", err)
	}
}
