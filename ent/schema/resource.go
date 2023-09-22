package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Resource struct {
	ent.Schema
}

func (Resource) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate()),
	}
}

func (Resource) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("message", map[string]interface{}{}),
	}
}

func (Resource) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("elements", Element.Type),
		edge.From("elements", Element.Type).Ref("resources"),
	}
}
