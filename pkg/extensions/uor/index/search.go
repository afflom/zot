package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"

	_ "github.com/mattn/go-sqlite3"
	ispec "github.com/opencontainers/image-spec/specs-go/v1"
	sschema "zotregistry.io/zot/pkg/extensions/uor/schema"
	"zotregistry.io/zot/pkg/extensions/uor/sqlite"
)

func AddStatement(statement sschema.Statement, repo string, descriptor ispec.Descriptor, eclient *sql.DB) error {
	fmt.Printf("preparing to write statement: %v\n", statement)
	repoMap := make(map[string]interface{})
	repoMap["namespace"] = repo

	//ctx := context.Background()
	descriptor.Annotations = nil
	descriptor.Platform = nil
	descriptor.URLs = nil
	bytes, err := json.Marshal(descriptor)
	if err != nil {
		return fmt.Errorf("marshalling error: %v", err)
	}
	var mdescriptor map[string]interface{}
	if err := json.Unmarshal(bytes, &mdescriptor); err != nil {
		return fmt.Errorf("unmarshalling error: %v", err)
	}
	// Construct a StatementRecord from the statement and descriptor
	statementRecord := sschema.StatementRecord{
		ResourceType: "uor_statement",
		Resource:     statement,
		LocatorType:  "oci_descriptor",
		Location:     descriptor,
	}
	statementRecord.Location.URLs = append(statementRecord.Location.URLs, repo)

	// Convert statementRecord to map[string]interface{}
	statementRecordMap := make(map[string]interface{})

	sb, err := json.Marshal(statementRecord)
	if err != nil {
		return fmt.Errorf("error marshalling statement record: %v", err)
	}
	if err := json.Unmarshal(sb, &statementRecordMap); err != nil {
		return fmt.Errorf("error unmarshalling statement record: %v", err)
	}

	fmt.Printf("statement record map: %v\n", statementRecordMap)

	initialTable := sqlite.Table{TableName: "uor_statementrecord"}
	schemas := make(map[string]interface{})

	result, err := sqlite.QueryDynamicSchema(eclient, statementRecordMap, &initialTable, schemas)
	fmt.Printf("result: %v\n", result)

	if err != nil {
		return fmt.Errorf("error querying extended database: %v", err)
	}
	if reflect.DeepEqual(statementRecordMap, result) {
		fmt.Printf("existing statement: %v\n", result)
		fmt.Printf("new statement: %v\n", mdescriptor)
		fmt.Printf("duplicate statement found for namespace: %s", repo)
		return nil
	}
	//var i int64
	sqlite.WriteToDynamicSchema(eclient, statementRecordMap, &initialTable, schemas)

	return nil
}

func Manifest2Statement(manifest ispec.Manifest) (sschema.Statement, error) {
	var statement sschema.Statement
	fmt.Println("Manifest2Statement called")

	// Handle the config object
	bConfig, err := json.Marshal(manifest.Config)
	if err != nil {
		return statement, fmt.Errorf("error marshalling config: %v", err)
	}
	fmt.Println("config marshalled")
	mConfig := make(map[string]interface{})
	if err := json.Unmarshal(bConfig, &mConfig); err != nil {
		return statement, fmt.Errorf("error unmarshalling config: %v", err)
	}
	fmt.Println("config unmarshalled")
	if len(mConfig) != 0 {
		statement.Object = &sschema.Element{
			ResourceType: manifest.Config.MediaType,
			Resource:     mConfig,
		}

		fmt.Printf("config is: %v\n", statement.Object)
	} else {
		statement.Object = nil
		fmt.Println("config is nil")
	}

	mLayers := make(map[string]interface{})
	for i, layer := range manifest.Layers {
		bLayer, err := json.Marshal(layer)
		if err != nil {
			return statement, fmt.Errorf("error marshalling layer: %v", err)
		}
		var layerMap map[string]interface{}
		if err := json.Unmarshal(bLayer, &layerMap); err != nil {
			return statement, fmt.Errorf("error unmarshalling layer: %v", err)
		}
		mLayers[fmt.Sprintf("layer%d", i)] = layerMap
	}
	statement.Subject = &sschema.Element{
		ResourceType: manifest.MediaType,
		Resource:     mLayers,
	}

	fmt.Printf("layers are: %+v\n", statement.Subject)

	cManifest := ispec.Manifest{}
	cManifest = manifest
	cManifest.Layers = nil
	cManifest.Config = ispec.Descriptor{}
	bManifest, err := json.Marshal(cManifest)
	if err != nil {
		return statement, fmt.Errorf("error marshalling manifest: %v", err)
	}
	mManifest := make(map[string]interface{})
	if err := json.Unmarshal(bManifest, &mManifest); err != nil {
		return statement, fmt.Errorf("error unmarshalling manifest: %v", err)
	}
	statement.Predicate = &sschema.Element{
		Resource:     mManifest,
		ResourceType: manifest.MediaType,
	}

	fmt.Printf("statement: %+v\n", statement)

	return statement, nil
}

func CreateStatementRecordSchema(db *sql.DB) {
	fmt.Println("CreateStatementRecordTables called")
	// Convert the statement to a JSON schema
	statementSchema := sschema.GenerateJSONSchema(reflect.TypeOf(sschema.StatementRecord{}), true)
	fmt.Printf("statementSchema: %v\n", statementSchema)
	// Initialize the database with the statementSchema
	sqlite.JSONSchemaToSQLiteSchema(statementSchema, nil, db, "uor_statementrecord", false)
	fmt.Println("sqlite schema initialized")
	// Convert the JSON schema to a GraphQL schema
}
