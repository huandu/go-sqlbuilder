// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"fmt"
	"strings"
)

// NewInsertBuilder creates a new INSERT builder.
func NewInsertBuilder() *InsertBuilder {
	args := &Args{}
	return &InsertBuilder{
		args: args,
	}
}

// InsertBuilder is a builder to build INSERT.
type InsertBuilder struct {
	table                     string
	cols                      []string
	values                    [][]string
	onDuplicateKeyUpdateExprs []string

	args *Args
}

// InsertInto sets table name in INSERT.
func (ib *InsertBuilder) InsertInto(table string) *InsertBuilder {
	ib.table = Escape(table)
	return ib
}

// Cols sets columns in INSERT.
func (ib *InsertBuilder) Cols(col ...string) *InsertBuilder {
	ib.cols = EscapeAll(col...)
	return ib
}

// Values adds a list of values for a row in INSERT.
func (ib *InsertBuilder) Values(value ...interface{}) *InsertBuilder {
	placeholders := make([]string, 0, len(value))

	for _, v := range value {
		placeholders = append(placeholders, ib.args.Add(v))
	}

	ib.values = append(ib.values, placeholders)
	return ib
}

// OnDuplicateKeyUpdate sets expressions of ON DUPLICATE KEY UPDATE in INSERT.
func (ib *InsertBuilder) OnDuplicateKeyUpdate(updateExpr ...string) *InsertBuilder {
	ib.onDuplicateKeyUpdateExprs = append(ib.onDuplicateKeyUpdateExprs, updateExpr...)
	return ib
}

// Assign represents SET "field = value" when ON DUPLICATE KEY UPDATE in INSERT.
func (ib *InsertBuilder) Assign(field string, value interface{}) string {
	return fmt.Sprintf("%v = %v", Escape(field), ib.args.Add(value))
}

// String returns the compiled DELETE string.
func (ib *InsertBuilder) String() string {
	s, _ := ib.Build()
	return s
}

// Build returns compiled INSERT string and args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (ib *InsertBuilder) Build() (sql string, args []interface{}) {
	buf := &bytes.Buffer{}
	buf.WriteString("INSERT INTO ")
	buf.WriteString(ib.table)

	if len(ib.cols) > 0 {
		buf.WriteString(" (")
		buf.WriteString(strings.Join(ib.cols, ", "))
		buf.WriteString(")")
	}

	buf.WriteString(" VALUES ")
	values := make([]string, 0, len(ib.values))

	for _, v := range ib.values {
		values = append(values, fmt.Sprintf("(%v)", strings.Join(v, ", ")))
	}

	buf.WriteString(strings.Join(values, ", "))

	if len(ib.onDuplicateKeyUpdateExprs) > 0 {
		buf.WriteString(" ON DUPLICATE KEY UPDATE ")
		buf.WriteString(strings.Join(ib.onDuplicateKeyUpdateExprs, " , "))
	}

	return ib.args.Compile(buf.String())
}
