// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"zotregistry.io/zot/ent/object"
	"zotregistry.io/zot/ent/spredicate"
	"zotregistry.io/zot/ent/statement"
	"zotregistry.io/zot/ent/subject"
)

// CollectFields tells the query-builder to eagerly load connected nodes by resolver context.
func (o *ObjectQuery) CollectFields(ctx context.Context, satisfies ...string) (*ObjectQuery, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return o, nil
	}
	if err := o.collectField(ctx, graphql.GetOperationContext(ctx), fc.Field, nil, satisfies...); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *ObjectQuery) collectField(ctx context.Context, opCtx *graphql.OperationContext, collected graphql.CollectedField, path []string, satisfies ...string) error {
	path = append([]string(nil), path...)
	var (
		unknownSeen    bool
		fieldSeen      = make(map[string]struct{}, len(object.Columns))
		selectedFields = []string{object.FieldID}
	)
	for _, field := range graphql.CollectFields(opCtx, collected.Selections, satisfies) {
		switch field.Name {
		case "statement":
			var (
				alias = field.Alias
				path  = append(path, alias)
				query = (&StatementClient{config: o.config}).Query()
			)
			if err := query.collectField(ctx, opCtx, field, path, mayAddCondition(satisfies, statementImplementors)...); err != nil {
				return err
			}
			o.WithNamedStatement(alias, func(wq *StatementQuery) {
				*wq = *query
			})
		case "objecttype":
			if _, ok := fieldSeen[object.FieldObjectType]; !ok {
				selectedFields = append(selectedFields, object.FieldObjectType)
				fieldSeen[object.FieldObjectType] = struct{}{}
			}
		case "object":
			if _, ok := fieldSeen[object.FieldObject]; !ok {
				selectedFields = append(selectedFields, object.FieldObject)
				fieldSeen[object.FieldObject] = struct{}{}
			}
		case "id":
		case "__typename":
		default:
			unknownSeen = true
		}
	}
	if !unknownSeen {
		o.Select(selectedFields...)
	}
	return nil
}

type objectPaginateArgs struct {
	first, last   *int
	after, before *Cursor
	opts          []ObjectPaginateOption
}

func newObjectPaginateArgs(rv map[string]any) *objectPaginateArgs {
	args := &objectPaginateArgs{}
	if rv == nil {
		return args
	}
	if v := rv[firstField]; v != nil {
		args.first = v.(*int)
	}
	if v := rv[lastField]; v != nil {
		args.last = v.(*int)
	}
	if v := rv[afterField]; v != nil {
		args.after = v.(*Cursor)
	}
	if v := rv[beforeField]; v != nil {
		args.before = v.(*Cursor)
	}
	if v, ok := rv[whereField].(*ObjectWhereInput); ok {
		args.opts = append(args.opts, WithObjectFilter(v.Filter))
	}
	return args
}

// CollectFields tells the query-builder to eagerly load connected nodes by resolver context.
func (s *SpredicateQuery) CollectFields(ctx context.Context, satisfies ...string) (*SpredicateQuery, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return s, nil
	}
	if err := s.collectField(ctx, graphql.GetOperationContext(ctx), fc.Field, nil, satisfies...); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SpredicateQuery) collectField(ctx context.Context, opCtx *graphql.OperationContext, collected graphql.CollectedField, path []string, satisfies ...string) error {
	path = append([]string(nil), path...)
	var (
		unknownSeen    bool
		fieldSeen      = make(map[string]struct{}, len(spredicate.Columns))
		selectedFields = []string{spredicate.FieldID}
	)
	for _, field := range graphql.CollectFields(opCtx, collected.Selections, satisfies) {
		switch field.Name {
		case "statement":
			var (
				alias = field.Alias
				path  = append(path, alias)
				query = (&StatementClient{config: s.config}).Query()
			)
			if err := query.collectField(ctx, opCtx, field, path, mayAddCondition(satisfies, statementImplementors)...); err != nil {
				return err
			}
			s.WithNamedStatement(alias, func(wq *StatementQuery) {
				*wq = *query
			})
		case "predicatetype":
			if _, ok := fieldSeen[spredicate.FieldPredicateType]; !ok {
				selectedFields = append(selectedFields, spredicate.FieldPredicateType)
				fieldSeen[spredicate.FieldPredicateType] = struct{}{}
			}
		case "predicate":
			if _, ok := fieldSeen[spredicate.FieldPredicate]; !ok {
				selectedFields = append(selectedFields, spredicate.FieldPredicate)
				fieldSeen[spredicate.FieldPredicate] = struct{}{}
			}
		case "id":
		case "__typename":
		default:
			unknownSeen = true
		}
	}
	if !unknownSeen {
		s.Select(selectedFields...)
	}
	return nil
}

type spredicatePaginateArgs struct {
	first, last   *int
	after, before *Cursor
	opts          []SpredicatePaginateOption
}

func newSpredicatePaginateArgs(rv map[string]any) *spredicatePaginateArgs {
	args := &spredicatePaginateArgs{}
	if rv == nil {
		return args
	}
	if v := rv[firstField]; v != nil {
		args.first = v.(*int)
	}
	if v := rv[lastField]; v != nil {
		args.last = v.(*int)
	}
	if v := rv[afterField]; v != nil {
		args.after = v.(*Cursor)
	}
	if v := rv[beforeField]; v != nil {
		args.before = v.(*Cursor)
	}
	if v, ok := rv[whereField].(*SpredicateWhereInput); ok {
		args.opts = append(args.opts, WithSpredicateFilter(v.Filter))
	}
	return args
}

// CollectFields tells the query-builder to eagerly load connected nodes by resolver context.
func (s *StatementQuery) CollectFields(ctx context.Context, satisfies ...string) (*StatementQuery, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return s, nil
	}
	if err := s.collectField(ctx, graphql.GetOperationContext(ctx), fc.Field, nil, satisfies...); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *StatementQuery) collectField(ctx context.Context, opCtx *graphql.OperationContext, collected graphql.CollectedField, path []string, satisfies ...string) error {
	path = append([]string(nil), path...)
	var (
		unknownSeen    bool
		fieldSeen      = make(map[string]struct{}, len(statement.Columns))
		selectedFields = []string{statement.FieldID}
	)
	for _, field := range graphql.CollectFields(opCtx, collected.Selections, satisfies) {
		switch field.Name {
		case "objects":
			var (
				alias = field.Alias
				path  = append(path, alias)
				query = (&ObjectClient{config: s.config}).Query()
			)
			if err := query.collectField(ctx, opCtx, field, path, mayAddCondition(satisfies, objectImplementors)...); err != nil {
				return err
			}
			s.WithNamedObjects(alias, func(wq *ObjectQuery) {
				*wq = *query
			})
		case "predicates":
			var (
				alias = field.Alias
				path  = append(path, alias)
				query = (&SpredicateClient{config: s.config}).Query()
			)
			if err := query.collectField(ctx, opCtx, field, path, mayAddCondition(satisfies, spredicateImplementors)...); err != nil {
				return err
			}
			s.WithNamedPredicates(alias, func(wq *SpredicateQuery) {
				*wq = *query
			})
		case "subjects":
			var (
				alias = field.Alias
				path  = append(path, alias)
				query = (&SubjectClient{config: s.config}).Query()
			)
			if err := query.collectField(ctx, opCtx, field, path, mayAddCondition(satisfies, subjectImplementors)...); err != nil {
				return err
			}
			s.WithNamedSubjects(alias, func(wq *SubjectQuery) {
				*wq = *query
			})
		case "namespace":
			if _, ok := fieldSeen[statement.FieldNamespace]; !ok {
				selectedFields = append(selectedFields, statement.FieldNamespace)
				fieldSeen[statement.FieldNamespace] = struct{}{}
			}
		case "statement":
			if _, ok := fieldSeen[statement.FieldStatement]; !ok {
				selectedFields = append(selectedFields, statement.FieldStatement)
				fieldSeen[statement.FieldStatement] = struct{}{}
			}
		case "id":
		case "__typename":
		default:
			unknownSeen = true
		}
	}
	if !unknownSeen {
		s.Select(selectedFields...)
	}
	return nil
}

type statementPaginateArgs struct {
	first, last   *int
	after, before *Cursor
	opts          []StatementPaginateOption
}

func newStatementPaginateArgs(rv map[string]any) *statementPaginateArgs {
	args := &statementPaginateArgs{}
	if rv == nil {
		return args
	}
	if v := rv[firstField]; v != nil {
		args.first = v.(*int)
	}
	if v := rv[lastField]; v != nil {
		args.last = v.(*int)
	}
	if v := rv[afterField]; v != nil {
		args.after = v.(*Cursor)
	}
	if v := rv[beforeField]; v != nil {
		args.before = v.(*Cursor)
	}
	if v, ok := rv[whereField].(*StatementWhereInput); ok {
		args.opts = append(args.opts, WithStatementFilter(v.Filter))
	}
	return args
}

// CollectFields tells the query-builder to eagerly load connected nodes by resolver context.
func (s *SubjectQuery) CollectFields(ctx context.Context, satisfies ...string) (*SubjectQuery, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return s, nil
	}
	if err := s.collectField(ctx, graphql.GetOperationContext(ctx), fc.Field, nil, satisfies...); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SubjectQuery) collectField(ctx context.Context, opCtx *graphql.OperationContext, collected graphql.CollectedField, path []string, satisfies ...string) error {
	path = append([]string(nil), path...)
	var (
		unknownSeen    bool
		fieldSeen      = make(map[string]struct{}, len(subject.Columns))
		selectedFields = []string{subject.FieldID}
	)
	for _, field := range graphql.CollectFields(opCtx, collected.Selections, satisfies) {
		switch field.Name {
		case "statement":
			var (
				alias = field.Alias
				path  = append(path, alias)
				query = (&StatementClient{config: s.config}).Query()
			)
			if err := query.collectField(ctx, opCtx, field, path, mayAddCondition(satisfies, statementImplementors)...); err != nil {
				return err
			}
			s.WithNamedStatement(alias, func(wq *StatementQuery) {
				*wq = *query
			})
		case "subjecttype":
			if _, ok := fieldSeen[subject.FieldSubjectType]; !ok {
				selectedFields = append(selectedFields, subject.FieldSubjectType)
				fieldSeen[subject.FieldSubjectType] = struct{}{}
			}
		case "subject":
			if _, ok := fieldSeen[subject.FieldSubject]; !ok {
				selectedFields = append(selectedFields, subject.FieldSubject)
				fieldSeen[subject.FieldSubject] = struct{}{}
			}
		case "id":
		case "__typename":
		default:
			unknownSeen = true
		}
	}
	if !unknownSeen {
		s.Select(selectedFields...)
	}
	return nil
}

type subjectPaginateArgs struct {
	first, last   *int
	after, before *Cursor
	opts          []SubjectPaginateOption
}

func newSubjectPaginateArgs(rv map[string]any) *subjectPaginateArgs {
	args := &subjectPaginateArgs{}
	if rv == nil {
		return args
	}
	if v := rv[firstField]; v != nil {
		args.first = v.(*int)
	}
	if v := rv[lastField]; v != nil {
		args.last = v.(*int)
	}
	if v := rv[afterField]; v != nil {
		args.after = v.(*Cursor)
	}
	if v := rv[beforeField]; v != nil {
		args.before = v.(*Cursor)
	}
	if v, ok := rv[whereField].(*SubjectWhereInput); ok {
		args.opts = append(args.opts, WithSubjectFilter(v.Filter))
	}
	return args
}

const (
	afterField     = "after"
	firstField     = "first"
	beforeField    = "before"
	lastField      = "last"
	orderByField   = "orderBy"
	directionField = "direction"
	fieldField     = "field"
	whereField     = "where"
)

func fieldArgs(ctx context.Context, whereInput any, path ...string) map[string]any {
	field := collectedField(ctx, path...)
	if field == nil || field.Arguments == nil {
		return nil
	}
	oc := graphql.GetOperationContext(ctx)
	args := field.ArgumentMap(oc.Variables)
	return unmarshalArgs(ctx, whereInput, args)
}

// unmarshalArgs allows extracting the field arguments from their raw representation.
func unmarshalArgs(ctx context.Context, whereInput any, args map[string]any) map[string]any {
	for _, k := range []string{firstField, lastField} {
		v, ok := args[k]
		if !ok {
			continue
		}
		i, err := graphql.UnmarshalInt(v)
		if err == nil {
			args[k] = &i
		}
	}
	for _, k := range []string{beforeField, afterField} {
		v, ok := args[k]
		if !ok {
			continue
		}
		c := &Cursor{}
		if c.UnmarshalGQL(v) == nil {
			args[k] = c
		}
	}
	if v, ok := args[whereField]; ok && whereInput != nil {
		if err := graphql.UnmarshalInputFromContext(ctx, v, whereInput); err == nil {
			args[whereField] = whereInput
		}
	}

	return args
}

func limitRows(partitionBy string, limit int, orderBy ...sql.Querier) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		d := sql.Dialect(s.Dialect())
		s.SetDistinct(false)
		with := d.With("src_query").
			As(s.Clone()).
			With("limited_query").
			As(
				d.Select("*").
					AppendSelectExprAs(
						sql.RowNumber().PartitionBy(partitionBy).OrderExpr(orderBy...),
						"row_number",
					).
					From(d.Table("src_query")),
			)
		t := d.Table("limited_query").As(s.TableName())
		*s = *d.Select(s.UnqualifiedColumns()...).
			From(t).
			Where(sql.LTE(t.C("row_number"), limit)).
			Prefix(with)
	}
}

// mayAddCondition appends another type condition to the satisfies list
// if it does not exist in the list.
func mayAddCondition(satisfies []string, typeCond []string) []string {
Cond:
	for _, c := range typeCond {
		for _, s := range satisfies {
			if c == s {
				continue Cond
			}
		}
		satisfies = append(satisfies, c)
	}
	return satisfies
}
