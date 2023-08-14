package zot

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"zotregistry.io/zot/ent"
	"zotregistry.io/zot/ent/predicate"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the resolver root.
type Resolver struct{ client *ent.Client }

// NewSchema creates a graphql executable schema.
func NewSchema(client *ent.Client) graphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers: &Resolver{client},
	})
}

const jsonPathKey = "jsonPath"

func (r *statementWhereInputResolver) JsonPath(ctx context.Context, obj *ent.StatementWhereInput, data *string) error {
	cctx := ctx.(*CustomContext)
	cctx.JsonPath = data
	return nil
}

func (r *statementWhereInputResolver) JsonValue(ctx context.Context, obj *ent.StatementWhereInput, data *string) error {
	cctx := ctx.(*CustomContext)
	cctx.JsonValue = data
	return nil
}

type CustomContext struct {
	context.Context
	JsonPath  *string
	JsonValue *string
}

// JSONExtractEquals generates a SQL predicate for querying JSON fields in SQLite.
func JSONExtractEquals(column, path, value string) sql.Predicate {
	return *sql.P(func(b *sql.Builder) {
		// Construct the SQL expression for JSON extraction.
		jsonQuery := fmt.Sprintf(`json_extract(%s, "$.%s") = ?`, column, path)
		b.WriteString(jsonQuery)
		b.Args(value)
	})
}

type statementHasJSONValue struct {
	column string
	path   string
	value  string
}

func HasJSONValue(column, path, value string) *statementHasJSONValue {
	return &statementHasJSONValue{
		column: column,
		path:   path,
		value:  value,
	}
}

// This is where we implement the sql.Predicate interface
func (s *statementHasJSONValue) Eval(builder *sql.Builder) {
	jsonQuery := fmt.Sprintf(`json_extract(%s, "$.%s") = ?`, s.column, s.path)
	builder.WriteString(jsonQuery)
	builder.Args(s.value)
}

func (s *statementHasJSONValue) P() *sql.Predicate {
	return sql.P(s.Eval)
}

func (s *statementHasJSONValue) SQL() (string, []interface{}) {
	jsonQuery := fmt.Sprintf(`json_extract(%s, "$.%s") = ?`, s.column, s.path)
	return jsonQuery, []interface{}{s.value}
}

func StatementHasJSONValue(column, path, value string) predicate.Statement {
	return func(selector *sql.Selector) {
		// Construct the SQL expression for JSON extraction.
		jsonQuery := fmt.Sprintf(`json_extract(%s, "$.%s") = ?`, column, path)
		fmt.Printf("jsonquery: %s\n", jsonQuery)
		predicate := sql.P(func(b *sql.Builder) {
			b.WriteString(jsonQuery)
			b.Args(value)
		})
		selector.Where(predicate)
		fmt.Printf("selector: %v\n", selector)
	}
}
