package search

/*
import (
	"log"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"log"
	"net/http"
	"sync"
	"github.com/graphql-go/graphql"
)

func initGraphqlSchema() *graphql.Schema {
	rootQuery = graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		// Fields to be populated dynamically
		Fields: graphql.Fields{},
	})

	rootMutation = graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		// Fields to be populated dynamically
		Fields: graphql.Fields{},
	})

	schemaConfig := graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	return &schema
}

// todo: integrate the below code.

var schemaLock sync.Mutex
var activeSchema graphql.Schema

func updateActiveSchema(newSchema graphql.Schema) {
	schemaLock.Lock()
	defer schemaLock.Unlock()
	activeSchema = newSchema
}

func getActiveSchema() graphql.Schema {
	schemaLock.Lock()
	defer schemaLock.Unlock()
	return activeSchema
}

func main() {
	// Initial empty schema
	initialRootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name:   "RootQuery",
		Fields: graphql.Fields{},
	})

	schemaConfig := graphql.SchemaConfig{Query: initialRootQuery}
	var err error
	activeSchema, err = graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		currentSchema := getActiveSchema()
		h := handler.New(&handler.Config{
			Schema: &currentSchema,
		})
		h.ServeHTTP(w, r)
	})

	go http.ListenAndServe(":8080", nil)

	// Somewhere else in your code when you want to extend the schema
	// This should happen after some event or trigger

	// ... Generate your new extended root query object here
	extendedRootQuery :=  your new root query object

	// Create a new schema with the extended root query
	newSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: extendedRootQuery,
	})
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	// Update the active schema
	updateActiveSchema(newSchema)
}
*/
