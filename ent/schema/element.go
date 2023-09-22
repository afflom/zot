package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Element struct {
	ent.Schema
}

func (Statement) Element() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate()),
	}
}

func (Element) Fields() []ent.Field {
	return []ent.Field{
		field.String("resourceType"),
		field.String("locatorType"),
	}
}

func (Element) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("statements", Statement.Type),
		edge.To("resources", Resource.Type),
		edge.To("locations", Resource.Type),
		edge.From("statements", Statement.Type).Ref("objects"),
		edge.From("statements", Statement.Type).Ref("predicates"),
		edge.From("statements", Statement.Type).Ref("subjects"),
		edge.From("statements", Statement.Type).Ref("statements"),
	}
}
