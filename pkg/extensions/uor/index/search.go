package index

import (
	"context"
	"encoding/json"
	"fmt"

	ggql "github.com/graphql-go/graphql"
	_ "github.com/mattn/go-sqlite3"
	ispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.mongodb.org/mongo-driver/mongo"
	zgql "zotregistry.io/zot/pkg/extensions/uor/graphql"
	"zotregistry.io/zot/pkg/extensions/uor/schema"
	sschema "zotregistry.io/zot/pkg/extensions/uor/schema"

	"github.com/invopop/jsonschema"
)

func AddStatement(statement sschema.Statement, repo string, descriptor ispec.Descriptor, eclient *mongo.Database) error {
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
		Resource:     schema.Resource{UORStatement: statement},
		LocatorType:  "oci_descriptor",
		Location:     schema.Location{OCIDescriptor: descriptor},
	}
	statementRecord.Location.OCIDescriptor.URLs = append(statementRecord.Location.OCIDescriptor.URLs, repo)

	// Convert statementRecord to map[string]interface{}
	statementRecordMap := make(map[string]interface{})

	sb, err := json.Marshal(statementRecord)
	fmt.Printf("statement record: %v\n", string(sb))
	if err != nil {
		return fmt.Errorf("error marshalling statement record: %v", err)
	}
	if err := json.Unmarshal(sb, &statementRecordMap); err != nil {
		return fmt.Errorf("error unmarshalling statement record: %v", err)
	}

	fmt.Printf("statement record map: %v\n", statementRecordMap)
	// write the statement record to mongodb
	var result *mongo.InsertOneResult
	result, err = eclient.Collection("statements").InsertOne(context.Background(), statementRecordMap)
	if err != nil {
		return fmt.Errorf("error inserting statement record: %v", err)
	}
	fmt.Printf("inserted statement record: %v\n", result.InsertedID)
	return nil
}

func Manifest2Statement(manifest ispec.Manifest) (sschema.Statement, error) {
	var statement sschema.Statement
	fmt.Println("Manifest2Statement called")
	// marshal manifest to a byte slice
	bmanifest, err := json.Marshal(manifest)
	if err != nil {
		return statement, fmt.Errorf("error marshalling manifest: %v", err)
	}
	var mmanifest map[string]interface{}
	json.Unmarshal(bmanifest, mmanifest)

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
			Resource:     mmanifest,
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
		Resource:     mmanifest,
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
		Resource:     mmanifest,
		ResourceType: manifest.MediaType,
	}

	fmt.Printf("statement: %+v\n", statement)

	return statement, nil
}

func CreateStatementRecordSchema(database mongo.Database) ggql.Schema {
	fmt.Println("CreateStatementRecordTables called")
	// Convert the statement to a JSON schema
	//statementSchema := sschema.GenerateJSONSchema(reflect.TypeOf(sschema.Statement{}))
	JSONstatementRecordSchema := jsonschema.Reflect(&sschema.StatementRecord{})

	data, err := json.MarshalIndent(JSONstatementRecordSchema, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(string(data))

	schemaRecord, err := json.Marshal(JSONstatementRecordSchema)
	if err != nil {
		panic(err.Error())

	}

	var schema map[string]interface{}
	json.Unmarshal([]byte(schemaRecord), &schema)

	// Convert to GraphQL Schema
	graphqlSchema, err := zgql.ConvertToGraphQL(schema, "uor_statementrecord", database)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("generated schema: %v\n", graphqlSchema)
	zgql.PrintFieldDefinitions(graphqlSchema.Fields(), "  ")

	gqlSchema, err := ggql.NewSchema(ggql.SchemaConfig{
		Query: graphqlSchema,
	})
	if err != nil {
		panic(err)
	}

	return gqlSchema
}
