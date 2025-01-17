// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/errcode"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"zotregistry.io/zot/ent/object"
	"zotregistry.io/zot/ent/spredicate"
	"zotregistry.io/zot/ent/statement"
	"zotregistry.io/zot/ent/subject"
)

// Common entgql types.
type (
	Cursor         = entgql.Cursor[int]
	PageInfo       = entgql.PageInfo[int]
	OrderDirection = entgql.OrderDirection
)

func orderFunc(o OrderDirection, field string) func(*sql.Selector) {
	if o == entgql.OrderDirectionDesc {
		return Desc(field)
	}
	return Asc(field)
}

const errInvalidPagination = "INVALID_PAGINATION"

func validateFirstLast(first, last *int) (err *gqlerror.Error) {
	switch {
	case first != nil && last != nil:
		err = &gqlerror.Error{
			Message: "Passing both `first` and `last` to paginate a connection is not supported.",
		}
	case first != nil && *first < 0:
		err = &gqlerror.Error{
			Message: "`first` on a connection cannot be less than zero.",
		}
		errcode.Set(err, errInvalidPagination)
	case last != nil && *last < 0:
		err = &gqlerror.Error{
			Message: "`last` on a connection cannot be less than zero.",
		}
		errcode.Set(err, errInvalidPagination)
	}
	return err
}

func collectedField(ctx context.Context, path ...string) *graphql.CollectedField {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return nil
	}
	field := fc.Field
	oc := graphql.GetOperationContext(ctx)
walk:
	for _, name := range path {
		for _, f := range graphql.CollectFields(oc, field.Selections, nil) {
			if f.Alias == name {
				field = f
				continue walk
			}
		}
		return nil
	}
	return &field
}

func hasCollectedField(ctx context.Context, path ...string) bool {
	if graphql.GetFieldContext(ctx) == nil {
		return true
	}
	return collectedField(ctx, path...) != nil
}

const (
	edgesField      = "edges"
	nodeField       = "node"
	pageInfoField   = "pageInfo"
	totalCountField = "totalCount"
)

func paginateLimit(first, last *int) int {
	var limit int
	if first != nil {
		limit = *first + 1
	} else if last != nil {
		limit = *last + 1
	}
	return limit
}

// ObjectEdge is the edge representation of Object.
type ObjectEdge struct {
	Node   *Object `json:"node"`
	Cursor Cursor  `json:"cursor"`
}

// ObjectConnection is the connection containing edges to Object.
type ObjectConnection struct {
	Edges      []*ObjectEdge `json:"edges"`
	PageInfo   PageInfo      `json:"pageInfo"`
	TotalCount int           `json:"totalCount"`
}

func (c *ObjectConnection) build(nodes []*Object, pager *objectPager, after *Cursor, first *int, before *Cursor, last *int) {
	c.PageInfo.HasNextPage = before != nil
	c.PageInfo.HasPreviousPage = after != nil
	if first != nil && *first+1 == len(nodes) {
		c.PageInfo.HasNextPage = true
		nodes = nodes[:len(nodes)-1]
	} else if last != nil && *last+1 == len(nodes) {
		c.PageInfo.HasPreviousPage = true
		nodes = nodes[:len(nodes)-1]
	}
	var nodeAt func(int) *Object
	if last != nil {
		n := len(nodes) - 1
		nodeAt = func(i int) *Object {
			return nodes[n-i]
		}
	} else {
		nodeAt = func(i int) *Object {
			return nodes[i]
		}
	}
	c.Edges = make([]*ObjectEdge, len(nodes))
	for i := range nodes {
		node := nodeAt(i)
		c.Edges[i] = &ObjectEdge{
			Node:   node,
			Cursor: pager.toCursor(node),
		}
	}
	if l := len(c.Edges); l > 0 {
		c.PageInfo.StartCursor = &c.Edges[0].Cursor
		c.PageInfo.EndCursor = &c.Edges[l-1].Cursor
	}
	if c.TotalCount == 0 {
		c.TotalCount = len(nodes)
	}
}

// ObjectPaginateOption enables pagination customization.
type ObjectPaginateOption func(*objectPager) error

// WithObjectOrder configures pagination ordering.
func WithObjectOrder(order *ObjectOrder) ObjectPaginateOption {
	if order == nil {
		order = DefaultObjectOrder
	}
	o := *order
	return func(pager *objectPager) error {
		if err := o.Direction.Validate(); err != nil {
			return err
		}
		if o.Field == nil {
			o.Field = DefaultObjectOrder.Field
		}
		pager.order = &o
		return nil
	}
}

// WithObjectFilter configures pagination filter.
func WithObjectFilter(filter func(*ObjectQuery) (*ObjectQuery, error)) ObjectPaginateOption {
	return func(pager *objectPager) error {
		if filter == nil {
			return errors.New("ObjectQuery filter cannot be nil")
		}
		pager.filter = filter
		return nil
	}
}

type objectPager struct {
	reverse bool
	order   *ObjectOrder
	filter  func(*ObjectQuery) (*ObjectQuery, error)
}

func newObjectPager(opts []ObjectPaginateOption, reverse bool) (*objectPager, error) {
	pager := &objectPager{reverse: reverse}
	for _, opt := range opts {
		if err := opt(pager); err != nil {
			return nil, err
		}
	}
	if pager.order == nil {
		pager.order = DefaultObjectOrder
	}
	return pager, nil
}

func (p *objectPager) applyFilter(query *ObjectQuery) (*ObjectQuery, error) {
	if p.filter != nil {
		return p.filter(query)
	}
	return query, nil
}

func (p *objectPager) toCursor(o *Object) Cursor {
	return p.order.Field.toCursor(o)
}

func (p *objectPager) applyCursors(query *ObjectQuery, after, before *Cursor) (*ObjectQuery, error) {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	for _, predicate := range entgql.CursorsPredicate(after, before, DefaultObjectOrder.Field.column, p.order.Field.column, direction) {
		query = query.Where(predicate)
	}
	return query, nil
}

func (p *objectPager) applyOrder(query *ObjectQuery) *ObjectQuery {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	query = query.Order(p.order.Field.toTerm(direction.OrderTermOption()))
	if p.order.Field != DefaultObjectOrder.Field {
		query = query.Order(DefaultObjectOrder.Field.toTerm(direction.OrderTermOption()))
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(p.order.Field.column)
	}
	return query
}

func (p *objectPager) orderExpr(query *ObjectQuery) sql.Querier {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(p.order.Field.column)
	}
	return sql.ExprFunc(func(b *sql.Builder) {
		b.Ident(p.order.Field.column).Pad().WriteString(string(direction))
		if p.order.Field != DefaultObjectOrder.Field {
			b.Comma().Ident(DefaultObjectOrder.Field.column).Pad().WriteString(string(direction))
		}
	})
}

// Paginate executes the query and returns a relay based cursor connection to Object.
func (o *ObjectQuery) Paginate(
	ctx context.Context, after *Cursor, first *int,
	before *Cursor, last *int, opts ...ObjectPaginateOption,
) (*ObjectConnection, error) {
	if err := validateFirstLast(first, last); err != nil {
		return nil, err
	}
	pager, err := newObjectPager(opts, last != nil)
	if err != nil {
		return nil, err
	}
	if o, err = pager.applyFilter(o); err != nil {
		return nil, err
	}
	conn := &ObjectConnection{Edges: []*ObjectEdge{}}
	ignoredEdges := !hasCollectedField(ctx, edgesField)
	if hasCollectedField(ctx, totalCountField) || hasCollectedField(ctx, pageInfoField) {
		hasPagination := after != nil || first != nil || before != nil || last != nil
		if hasPagination || ignoredEdges {
			c := o.Clone()
			c.ctx.Fields = nil
			if conn.TotalCount, err = c.Count(ctx); err != nil {
				return nil, err
			}
			conn.PageInfo.HasNextPage = first != nil && conn.TotalCount > 0
			conn.PageInfo.HasPreviousPage = last != nil && conn.TotalCount > 0
		}
	}
	if ignoredEdges || (first != nil && *first == 0) || (last != nil && *last == 0) {
		return conn, nil
	}
	if o, err = pager.applyCursors(o, after, before); err != nil {
		return nil, err
	}
	if limit := paginateLimit(first, last); limit != 0 {
		o.Limit(limit)
	}
	if field := collectedField(ctx, edgesField, nodeField); field != nil {
		if err := o.collectField(ctx, graphql.GetOperationContext(ctx), *field, []string{edgesField, nodeField}); err != nil {
			return nil, err
		}
	}
	o = pager.applyOrder(o)
	nodes, err := o.All(ctx)
	if err != nil {
		return nil, err
	}
	conn.build(nodes, pager, after, first, before, last)
	return conn, nil
}

// ObjectOrderField defines the ordering field of Object.
type ObjectOrderField struct {
	// Value extracts the ordering value from the given Object.
	Value    func(*Object) (ent.Value, error)
	column   string // field or computed.
	toTerm   func(...sql.OrderTermOption) object.OrderOption
	toCursor func(*Object) Cursor
}

// ObjectOrder defines the ordering of Object.
type ObjectOrder struct {
	Direction OrderDirection    `json:"direction"`
	Field     *ObjectOrderField `json:"field"`
}

// DefaultObjectOrder is the default ordering of Object.
var DefaultObjectOrder = &ObjectOrder{
	Direction: entgql.OrderDirectionAsc,
	Field: &ObjectOrderField{
		Value: func(o *Object) (ent.Value, error) {
			return o.ID, nil
		},
		column: object.FieldID,
		toTerm: object.ByID,
		toCursor: func(o *Object) Cursor {
			return Cursor{ID: o.ID}
		},
	},
}

// ToEdge converts Object into ObjectEdge.
func (o *Object) ToEdge(order *ObjectOrder) *ObjectEdge {
	if order == nil {
		order = DefaultObjectOrder
	}
	return &ObjectEdge{
		Node:   o,
		Cursor: order.Field.toCursor(o),
	}
}

// SpredicateEdge is the edge representation of Spredicate.
type SpredicateEdge struct {
	Node   *Spredicate `json:"node"`
	Cursor Cursor      `json:"cursor"`
}

// SpredicateConnection is the connection containing edges to Spredicate.
type SpredicateConnection struct {
	Edges      []*SpredicateEdge `json:"edges"`
	PageInfo   PageInfo          `json:"pageInfo"`
	TotalCount int               `json:"totalCount"`
}

func (c *SpredicateConnection) build(nodes []*Spredicate, pager *spredicatePager, after *Cursor, first *int, before *Cursor, last *int) {
	c.PageInfo.HasNextPage = before != nil
	c.PageInfo.HasPreviousPage = after != nil
	if first != nil && *first+1 == len(nodes) {
		c.PageInfo.HasNextPage = true
		nodes = nodes[:len(nodes)-1]
	} else if last != nil && *last+1 == len(nodes) {
		c.PageInfo.HasPreviousPage = true
		nodes = nodes[:len(nodes)-1]
	}
	var nodeAt func(int) *Spredicate
	if last != nil {
		n := len(nodes) - 1
		nodeAt = func(i int) *Spredicate {
			return nodes[n-i]
		}
	} else {
		nodeAt = func(i int) *Spredicate {
			return nodes[i]
		}
	}
	c.Edges = make([]*SpredicateEdge, len(nodes))
	for i := range nodes {
		node := nodeAt(i)
		c.Edges[i] = &SpredicateEdge{
			Node:   node,
			Cursor: pager.toCursor(node),
		}
	}
	if l := len(c.Edges); l > 0 {
		c.PageInfo.StartCursor = &c.Edges[0].Cursor
		c.PageInfo.EndCursor = &c.Edges[l-1].Cursor
	}
	if c.TotalCount == 0 {
		c.TotalCount = len(nodes)
	}
}

// SpredicatePaginateOption enables pagination customization.
type SpredicatePaginateOption func(*spredicatePager) error

// WithSpredicateOrder configures pagination ordering.
func WithSpredicateOrder(order *SpredicateOrder) SpredicatePaginateOption {
	if order == nil {
		order = DefaultSpredicateOrder
	}
	o := *order
	return func(pager *spredicatePager) error {
		if err := o.Direction.Validate(); err != nil {
			return err
		}
		if o.Field == nil {
			o.Field = DefaultSpredicateOrder.Field
		}
		pager.order = &o
		return nil
	}
}

// WithSpredicateFilter configures pagination filter.
func WithSpredicateFilter(filter func(*SpredicateQuery) (*SpredicateQuery, error)) SpredicatePaginateOption {
	return func(pager *spredicatePager) error {
		if filter == nil {
			return errors.New("SpredicateQuery filter cannot be nil")
		}
		pager.filter = filter
		return nil
	}
}

type spredicatePager struct {
	reverse bool
	order   *SpredicateOrder
	filter  func(*SpredicateQuery) (*SpredicateQuery, error)
}

func newSpredicatePager(opts []SpredicatePaginateOption, reverse bool) (*spredicatePager, error) {
	pager := &spredicatePager{reverse: reverse}
	for _, opt := range opts {
		if err := opt(pager); err != nil {
			return nil, err
		}
	}
	if pager.order == nil {
		pager.order = DefaultSpredicateOrder
	}
	return pager, nil
}

func (p *spredicatePager) applyFilter(query *SpredicateQuery) (*SpredicateQuery, error) {
	if p.filter != nil {
		return p.filter(query)
	}
	return query, nil
}

func (p *spredicatePager) toCursor(s *Spredicate) Cursor {
	return p.order.Field.toCursor(s)
}

func (p *spredicatePager) applyCursors(query *SpredicateQuery, after, before *Cursor) (*SpredicateQuery, error) {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	for _, predicate := range entgql.CursorsPredicate(after, before, DefaultSpredicateOrder.Field.column, p.order.Field.column, direction) {
		query = query.Where(predicate)
	}
	return query, nil
}

func (p *spredicatePager) applyOrder(query *SpredicateQuery) *SpredicateQuery {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	query = query.Order(p.order.Field.toTerm(direction.OrderTermOption()))
	if p.order.Field != DefaultSpredicateOrder.Field {
		query = query.Order(DefaultSpredicateOrder.Field.toTerm(direction.OrderTermOption()))
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(p.order.Field.column)
	}
	return query
}

func (p *spredicatePager) orderExpr(query *SpredicateQuery) sql.Querier {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(p.order.Field.column)
	}
	return sql.ExprFunc(func(b *sql.Builder) {
		b.Ident(p.order.Field.column).Pad().WriteString(string(direction))
		if p.order.Field != DefaultSpredicateOrder.Field {
			b.Comma().Ident(DefaultSpredicateOrder.Field.column).Pad().WriteString(string(direction))
		}
	})
}

// Paginate executes the query and returns a relay based cursor connection to Spredicate.
func (s *SpredicateQuery) Paginate(
	ctx context.Context, after *Cursor, first *int,
	before *Cursor, last *int, opts ...SpredicatePaginateOption,
) (*SpredicateConnection, error) {
	if err := validateFirstLast(first, last); err != nil {
		return nil, err
	}
	pager, err := newSpredicatePager(opts, last != nil)
	if err != nil {
		return nil, err
	}
	if s, err = pager.applyFilter(s); err != nil {
		return nil, err
	}
	conn := &SpredicateConnection{Edges: []*SpredicateEdge{}}
	ignoredEdges := !hasCollectedField(ctx, edgesField)
	if hasCollectedField(ctx, totalCountField) || hasCollectedField(ctx, pageInfoField) {
		hasPagination := after != nil || first != nil || before != nil || last != nil
		if hasPagination || ignoredEdges {
			c := s.Clone()
			c.ctx.Fields = nil
			if conn.TotalCount, err = c.Count(ctx); err != nil {
				return nil, err
			}
			conn.PageInfo.HasNextPage = first != nil && conn.TotalCount > 0
			conn.PageInfo.HasPreviousPage = last != nil && conn.TotalCount > 0
		}
	}
	if ignoredEdges || (first != nil && *first == 0) || (last != nil && *last == 0) {
		return conn, nil
	}
	if s, err = pager.applyCursors(s, after, before); err != nil {
		return nil, err
	}
	if limit := paginateLimit(first, last); limit != 0 {
		s.Limit(limit)
	}
	if field := collectedField(ctx, edgesField, nodeField); field != nil {
		if err := s.collectField(ctx, graphql.GetOperationContext(ctx), *field, []string{edgesField, nodeField}); err != nil {
			return nil, err
		}
	}
	s = pager.applyOrder(s)
	nodes, err := s.All(ctx)
	if err != nil {
		return nil, err
	}
	conn.build(nodes, pager, after, first, before, last)
	return conn, nil
}

// SpredicateOrderField defines the ordering field of Spredicate.
type SpredicateOrderField struct {
	// Value extracts the ordering value from the given Spredicate.
	Value    func(*Spredicate) (ent.Value, error)
	column   string // field or computed.
	toTerm   func(...sql.OrderTermOption) spredicate.OrderOption
	toCursor func(*Spredicate) Cursor
}

// SpredicateOrder defines the ordering of Spredicate.
type SpredicateOrder struct {
	Direction OrderDirection        `json:"direction"`
	Field     *SpredicateOrderField `json:"field"`
}

// DefaultSpredicateOrder is the default ordering of Spredicate.
var DefaultSpredicateOrder = &SpredicateOrder{
	Direction: entgql.OrderDirectionAsc,
	Field: &SpredicateOrderField{
		Value: func(s *Spredicate) (ent.Value, error) {
			return s.ID, nil
		},
		column: spredicate.FieldID,
		toTerm: spredicate.ByID,
		toCursor: func(s *Spredicate) Cursor {
			return Cursor{ID: s.ID}
		},
	},
}

// ToEdge converts Spredicate into SpredicateEdge.
func (s *Spredicate) ToEdge(order *SpredicateOrder) *SpredicateEdge {
	if order == nil {
		order = DefaultSpredicateOrder
	}
	return &SpredicateEdge{
		Node:   s,
		Cursor: order.Field.toCursor(s),
	}
}

// StatementEdge is the edge representation of Statement.
type StatementEdge struct {
	Node   *Statement `json:"node"`
	Cursor Cursor     `json:"cursor"`
}

// StatementConnection is the connection containing edges to Statement.
type StatementConnection struct {
	Edges      []*StatementEdge `json:"edges"`
	PageInfo   PageInfo         `json:"pageInfo"`
	TotalCount int              `json:"totalCount"`
}

func (c *StatementConnection) build(nodes []*Statement, pager *statementPager, after *Cursor, first *int, before *Cursor, last *int) {
	c.PageInfo.HasNextPage = before != nil
	c.PageInfo.HasPreviousPage = after != nil
	if first != nil && *first+1 == len(nodes) {
		c.PageInfo.HasNextPage = true
		nodes = nodes[:len(nodes)-1]
	} else if last != nil && *last+1 == len(nodes) {
		c.PageInfo.HasPreviousPage = true
		nodes = nodes[:len(nodes)-1]
	}
	var nodeAt func(int) *Statement
	if last != nil {
		n := len(nodes) - 1
		nodeAt = func(i int) *Statement {
			return nodes[n-i]
		}
	} else {
		nodeAt = func(i int) *Statement {
			return nodes[i]
		}
	}
	c.Edges = make([]*StatementEdge, len(nodes))
	for i := range nodes {
		node := nodeAt(i)
		c.Edges[i] = &StatementEdge{
			Node:   node,
			Cursor: pager.toCursor(node),
		}
	}
	if l := len(c.Edges); l > 0 {
		c.PageInfo.StartCursor = &c.Edges[0].Cursor
		c.PageInfo.EndCursor = &c.Edges[l-1].Cursor
	}
	if c.TotalCount == 0 {
		c.TotalCount = len(nodes)
	}
}

// StatementPaginateOption enables pagination customization.
type StatementPaginateOption func(*statementPager) error

// WithStatementOrder configures pagination ordering.
func WithStatementOrder(order *StatementOrder) StatementPaginateOption {
	if order == nil {
		order = DefaultStatementOrder
	}
	o := *order
	return func(pager *statementPager) error {
		if err := o.Direction.Validate(); err != nil {
			return err
		}
		if o.Field == nil {
			o.Field = DefaultStatementOrder.Field
		}
		pager.order = &o
		return nil
	}
}

// WithStatementFilter configures pagination filter.
func WithStatementFilter(filter func(*StatementQuery) (*StatementQuery, error)) StatementPaginateOption {
	return func(pager *statementPager) error {
		if filter == nil {
			return errors.New("StatementQuery filter cannot be nil")
		}
		pager.filter = filter
		return nil
	}
}

type statementPager struct {
	reverse bool
	order   *StatementOrder
	filter  func(*StatementQuery) (*StatementQuery, error)
}

func newStatementPager(opts []StatementPaginateOption, reverse bool) (*statementPager, error) {
	pager := &statementPager{reverse: reverse}
	for _, opt := range opts {
		if err := opt(pager); err != nil {
			return nil, err
		}
	}
	if pager.order == nil {
		pager.order = DefaultStatementOrder
	}
	return pager, nil
}

func (p *statementPager) applyFilter(query *StatementQuery) (*StatementQuery, error) {
	if p.filter != nil {
		return p.filter(query)
	}
	return query, nil
}

func (p *statementPager) toCursor(s *Statement) Cursor {
	return p.order.Field.toCursor(s)
}

func (p *statementPager) applyCursors(query *StatementQuery, after, before *Cursor) (*StatementQuery, error) {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	for _, predicate := range entgql.CursorsPredicate(after, before, DefaultStatementOrder.Field.column, p.order.Field.column, direction) {
		query = query.Where(predicate)
	}
	return query, nil
}

func (p *statementPager) applyOrder(query *StatementQuery) *StatementQuery {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	query = query.Order(p.order.Field.toTerm(direction.OrderTermOption()))
	if p.order.Field != DefaultStatementOrder.Field {
		query = query.Order(DefaultStatementOrder.Field.toTerm(direction.OrderTermOption()))
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(p.order.Field.column)
	}
	return query
}

func (p *statementPager) orderExpr(query *StatementQuery) sql.Querier {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(p.order.Field.column)
	}
	return sql.ExprFunc(func(b *sql.Builder) {
		b.Ident(p.order.Field.column).Pad().WriteString(string(direction))
		if p.order.Field != DefaultStatementOrder.Field {
			b.Comma().Ident(DefaultStatementOrder.Field.column).Pad().WriteString(string(direction))
		}
	})
}

// Paginate executes the query and returns a relay based cursor connection to Statement.
func (s *StatementQuery) Paginate(
	ctx context.Context, after *Cursor, first *int,
	before *Cursor, last *int, opts ...StatementPaginateOption,
) (*StatementConnection, error) {
	if err := validateFirstLast(first, last); err != nil {
		return nil, err
	}
	pager, err := newStatementPager(opts, last != nil)
	if err != nil {
		return nil, err
	}
	if s, err = pager.applyFilter(s); err != nil {
		return nil, err
	}
	conn := &StatementConnection{Edges: []*StatementEdge{}}
	ignoredEdges := !hasCollectedField(ctx, edgesField)
	if hasCollectedField(ctx, totalCountField) || hasCollectedField(ctx, pageInfoField) {
		hasPagination := after != nil || first != nil || before != nil || last != nil
		if hasPagination || ignoredEdges {
			c := s.Clone()
			c.ctx.Fields = nil
			if conn.TotalCount, err = c.Count(ctx); err != nil {
				return nil, err
			}
			conn.PageInfo.HasNextPage = first != nil && conn.TotalCount > 0
			conn.PageInfo.HasPreviousPage = last != nil && conn.TotalCount > 0
		}
	}
	if ignoredEdges || (first != nil && *first == 0) || (last != nil && *last == 0) {
		return conn, nil
	}
	if s, err = pager.applyCursors(s, after, before); err != nil {
		return nil, err
	}
	if limit := paginateLimit(first, last); limit != 0 {
		s.Limit(limit)
	}
	if field := collectedField(ctx, edgesField, nodeField); field != nil {
		if err := s.collectField(ctx, graphql.GetOperationContext(ctx), *field, []string{edgesField, nodeField}); err != nil {
			return nil, err
		}
	}
	s = pager.applyOrder(s)
	nodes, err := s.All(ctx)
	if err != nil {
		return nil, err
	}
	conn.build(nodes, pager, after, first, before, last)
	return conn, nil
}

// StatementOrderField defines the ordering field of Statement.
type StatementOrderField struct {
	// Value extracts the ordering value from the given Statement.
	Value    func(*Statement) (ent.Value, error)
	column   string // field or computed.
	toTerm   func(...sql.OrderTermOption) statement.OrderOption
	toCursor func(*Statement) Cursor
}

// StatementOrder defines the ordering of Statement.
type StatementOrder struct {
	Direction OrderDirection       `json:"direction"`
	Field     *StatementOrderField `json:"field"`
}

// DefaultStatementOrder is the default ordering of Statement.
var DefaultStatementOrder = &StatementOrder{
	Direction: entgql.OrderDirectionAsc,
	Field: &StatementOrderField{
		Value: func(s *Statement) (ent.Value, error) {
			return s.ID, nil
		},
		column: statement.FieldID,
		toTerm: statement.ByID,
		toCursor: func(s *Statement) Cursor {
			return Cursor{ID: s.ID}
		},
	},
}

// ToEdge converts Statement into StatementEdge.
func (s *Statement) ToEdge(order *StatementOrder) *StatementEdge {
	if order == nil {
		order = DefaultStatementOrder
	}
	return &StatementEdge{
		Node:   s,
		Cursor: order.Field.toCursor(s),
	}
}

// SubjectEdge is the edge representation of Subject.
type SubjectEdge struct {
	Node   *Subject `json:"node"`
	Cursor Cursor   `json:"cursor"`
}

// SubjectConnection is the connection containing edges to Subject.
type SubjectConnection struct {
	Edges      []*SubjectEdge `json:"edges"`
	PageInfo   PageInfo       `json:"pageInfo"`
	TotalCount int            `json:"totalCount"`
}

func (c *SubjectConnection) build(nodes []*Subject, pager *subjectPager, after *Cursor, first *int, before *Cursor, last *int) {
	c.PageInfo.HasNextPage = before != nil
	c.PageInfo.HasPreviousPage = after != nil
	if first != nil && *first+1 == len(nodes) {
		c.PageInfo.HasNextPage = true
		nodes = nodes[:len(nodes)-1]
	} else if last != nil && *last+1 == len(nodes) {
		c.PageInfo.HasPreviousPage = true
		nodes = nodes[:len(nodes)-1]
	}
	var nodeAt func(int) *Subject
	if last != nil {
		n := len(nodes) - 1
		nodeAt = func(i int) *Subject {
			return nodes[n-i]
		}
	} else {
		nodeAt = func(i int) *Subject {
			return nodes[i]
		}
	}
	c.Edges = make([]*SubjectEdge, len(nodes))
	for i := range nodes {
		node := nodeAt(i)
		c.Edges[i] = &SubjectEdge{
			Node:   node,
			Cursor: pager.toCursor(node),
		}
	}
	if l := len(c.Edges); l > 0 {
		c.PageInfo.StartCursor = &c.Edges[0].Cursor
		c.PageInfo.EndCursor = &c.Edges[l-1].Cursor
	}
	if c.TotalCount == 0 {
		c.TotalCount = len(nodes)
	}
}

// SubjectPaginateOption enables pagination customization.
type SubjectPaginateOption func(*subjectPager) error

// WithSubjectOrder configures pagination ordering.
func WithSubjectOrder(order *SubjectOrder) SubjectPaginateOption {
	if order == nil {
		order = DefaultSubjectOrder
	}
	o := *order
	return func(pager *subjectPager) error {
		if err := o.Direction.Validate(); err != nil {
			return err
		}
		if o.Field == nil {
			o.Field = DefaultSubjectOrder.Field
		}
		pager.order = &o
		return nil
	}
}

// WithSubjectFilter configures pagination filter.
func WithSubjectFilter(filter func(*SubjectQuery) (*SubjectQuery, error)) SubjectPaginateOption {
	return func(pager *subjectPager) error {
		if filter == nil {
			return errors.New("SubjectQuery filter cannot be nil")
		}
		pager.filter = filter
		return nil
	}
}

type subjectPager struct {
	reverse bool
	order   *SubjectOrder
	filter  func(*SubjectQuery) (*SubjectQuery, error)
}

func newSubjectPager(opts []SubjectPaginateOption, reverse bool) (*subjectPager, error) {
	pager := &subjectPager{reverse: reverse}
	for _, opt := range opts {
		if err := opt(pager); err != nil {
			return nil, err
		}
	}
	if pager.order == nil {
		pager.order = DefaultSubjectOrder
	}
	return pager, nil
}

func (p *subjectPager) applyFilter(query *SubjectQuery) (*SubjectQuery, error) {
	if p.filter != nil {
		return p.filter(query)
	}
	return query, nil
}

func (p *subjectPager) toCursor(s *Subject) Cursor {
	return p.order.Field.toCursor(s)
}

func (p *subjectPager) applyCursors(query *SubjectQuery, after, before *Cursor) (*SubjectQuery, error) {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	for _, predicate := range entgql.CursorsPredicate(after, before, DefaultSubjectOrder.Field.column, p.order.Field.column, direction) {
		query = query.Where(predicate)
	}
	return query, nil
}

func (p *subjectPager) applyOrder(query *SubjectQuery) *SubjectQuery {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	query = query.Order(p.order.Field.toTerm(direction.OrderTermOption()))
	if p.order.Field != DefaultSubjectOrder.Field {
		query = query.Order(DefaultSubjectOrder.Field.toTerm(direction.OrderTermOption()))
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(p.order.Field.column)
	}
	return query
}

func (p *subjectPager) orderExpr(query *SubjectQuery) sql.Querier {
	direction := p.order.Direction
	if p.reverse {
		direction = direction.Reverse()
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(p.order.Field.column)
	}
	return sql.ExprFunc(func(b *sql.Builder) {
		b.Ident(p.order.Field.column).Pad().WriteString(string(direction))
		if p.order.Field != DefaultSubjectOrder.Field {
			b.Comma().Ident(DefaultSubjectOrder.Field.column).Pad().WriteString(string(direction))
		}
	})
}

// Paginate executes the query and returns a relay based cursor connection to Subject.
func (s *SubjectQuery) Paginate(
	ctx context.Context, after *Cursor, first *int,
	before *Cursor, last *int, opts ...SubjectPaginateOption,
) (*SubjectConnection, error) {
	if err := validateFirstLast(first, last); err != nil {
		return nil, err
	}
	pager, err := newSubjectPager(opts, last != nil)
	if err != nil {
		return nil, err
	}
	if s, err = pager.applyFilter(s); err != nil {
		return nil, err
	}
	conn := &SubjectConnection{Edges: []*SubjectEdge{}}
	ignoredEdges := !hasCollectedField(ctx, edgesField)
	if hasCollectedField(ctx, totalCountField) || hasCollectedField(ctx, pageInfoField) {
		hasPagination := after != nil || first != nil || before != nil || last != nil
		if hasPagination || ignoredEdges {
			c := s.Clone()
			c.ctx.Fields = nil
			if conn.TotalCount, err = c.Count(ctx); err != nil {
				return nil, err
			}
			conn.PageInfo.HasNextPage = first != nil && conn.TotalCount > 0
			conn.PageInfo.HasPreviousPage = last != nil && conn.TotalCount > 0
		}
	}
	if ignoredEdges || (first != nil && *first == 0) || (last != nil && *last == 0) {
		return conn, nil
	}
	if s, err = pager.applyCursors(s, after, before); err != nil {
		return nil, err
	}
	if limit := paginateLimit(first, last); limit != 0 {
		s.Limit(limit)
	}
	if field := collectedField(ctx, edgesField, nodeField); field != nil {
		if err := s.collectField(ctx, graphql.GetOperationContext(ctx), *field, []string{edgesField, nodeField}); err != nil {
			return nil, err
		}
	}
	s = pager.applyOrder(s)
	nodes, err := s.All(ctx)
	if err != nil {
		return nil, err
	}
	conn.build(nodes, pager, after, first, before, last)
	return conn, nil
}

// SubjectOrderField defines the ordering field of Subject.
type SubjectOrderField struct {
	// Value extracts the ordering value from the given Subject.
	Value    func(*Subject) (ent.Value, error)
	column   string // field or computed.
	toTerm   func(...sql.OrderTermOption) subject.OrderOption
	toCursor func(*Subject) Cursor
}

// SubjectOrder defines the ordering of Subject.
type SubjectOrder struct {
	Direction OrderDirection     `json:"direction"`
	Field     *SubjectOrderField `json:"field"`
}

// DefaultSubjectOrder is the default ordering of Subject.
var DefaultSubjectOrder = &SubjectOrder{
	Direction: entgql.OrderDirectionAsc,
	Field: &SubjectOrderField{
		Value: func(s *Subject) (ent.Value, error) {
			return s.ID, nil
		},
		column: subject.FieldID,
		toTerm: subject.ByID,
		toCursor: func(s *Subject) Cursor {
			return Cursor{ID: s.ID}
		},
	},
}

// ToEdge converts Subject into SubjectEdge.
func (s *Subject) ToEdge(order *SubjectOrder) *SubjectEdge {
	if order == nil {
		order = DefaultSubjectOrder
	}
	return &SubjectEdge{
		Node:   s,
		Cursor: order.Field.toCursor(s),
	}
}
