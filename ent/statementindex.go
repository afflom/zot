// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"zotregistry.io/zot/ent/statementindex"
	"zotregistry.io/zot/pkg/search/schema"
)

// StatementIndex is the model entity for the StatementIndex schema.
type StatementIndex struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Object holds the value of the "object" field.
	Object schema.Object `json:"object,omitempty"`
	// Predicate holds the value of the "predicate" field.
	Predicate schema.Predicate `json:"predicate,omitempty"`
	// Subject holds the value of the "subject" field.
	Subject schema.Subject `json:"subject,omitempty"`
	// Statement holds the value of the "statement" field.
	Statement    schema.Location `json:"statement,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*StatementIndex) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case statementindex.FieldObject, statementindex.FieldPredicate, statementindex.FieldSubject, statementindex.FieldStatement:
			values[i] = new([]byte)
		case statementindex.FieldID:
			values[i] = new(sql.NullInt64)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the StatementIndex fields.
func (si *StatementIndex) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case statementindex.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			si.ID = int(value.Int64)
		case statementindex.FieldObject:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field object", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &si.Object); err != nil {
					return fmt.Errorf("unmarshal field object: %w", err)
				}
			}
		case statementindex.FieldPredicate:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field predicate", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &si.Predicate); err != nil {
					return fmt.Errorf("unmarshal field predicate: %w", err)
				}
			}
		case statementindex.FieldSubject:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field subject", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &si.Subject); err != nil {
					return fmt.Errorf("unmarshal field subject: %w", err)
				}
			}
		case statementindex.FieldStatement:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field statement", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &si.Statement); err != nil {
					return fmt.Errorf("unmarshal field statement: %w", err)
				}
			}
		default:
			si.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the StatementIndex.
// This includes values selected through modifiers, order, etc.
func (si *StatementIndex) Value(name string) (ent.Value, error) {
	return si.selectValues.Get(name)
}

// Update returns a builder for updating this StatementIndex.
// Note that you need to call StatementIndex.Unwrap() before calling this method if this StatementIndex
// was returned from a transaction, and the transaction was committed or rolled back.
func (si *StatementIndex) Update() *StatementIndexUpdateOne {
	return NewStatementIndexClient(si.config).UpdateOne(si)
}

// Unwrap unwraps the StatementIndex entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (si *StatementIndex) Unwrap() *StatementIndex {
	_tx, ok := si.config.driver.(*txDriver)
	if !ok {
		panic("ent: StatementIndex is not a transactional entity")
	}
	si.config.driver = _tx.drv
	return si
}

// String implements the fmt.Stringer.
func (si *StatementIndex) String() string {
	var builder strings.Builder
	builder.WriteString("StatementIndex(")
	builder.WriteString(fmt.Sprintf("id=%v, ", si.ID))
	builder.WriteString("object=")
	builder.WriteString(fmt.Sprintf("%v", si.Object))
	builder.WriteString(", ")
	builder.WriteString("predicate=")
	builder.WriteString(fmt.Sprintf("%v", si.Predicate))
	builder.WriteString(", ")
	builder.WriteString("subject=")
	builder.WriteString(fmt.Sprintf("%v", si.Subject))
	builder.WriteString(", ")
	builder.WriteString("statement=")
	builder.WriteString(fmt.Sprintf("%v", si.Statement))
	builder.WriteByte(')')
	return builder.String()
}

// StatementIndexes is a parsable slice of StatementIndex.
type StatementIndexes []*StatementIndex
