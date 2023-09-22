// statement.go
package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Statement struct {
	ent.Schema
}

func (Statement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate()),
	}
}

func (Statement) Fields() []ent.Field {
	return []ent.Field{
		field.String("mediaType"),
	}
}

func (Statement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("objects", Element.Type),
		edge.To("predicates", Element.Type),
		edge.To("subjects", Element.Type),
		edge.To("statements", Element.Type),
		edge.From("statements", Element.Type).Ref("objects"),
		edge.From("statements", Element.Type).Ref("predicates"),
		edge.From("statements", Element.Type).Ref("subjects"),
		edge.From("statements", Element.Type).Ref("statements"),
	}
}
