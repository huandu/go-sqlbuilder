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
	return DefaultFlavor.NewInsertBuilder()
}

func newInsertBuilder() *InsertBuilder {
	args := &Args{}
	return &InsertBuilder{
		args: args,
	}
}

// InsertBuilder is a builder to build INSERT.
type InsertBuilder struct {
	table  string
	cols   []string
	values [][]string

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

// String returns the compiled DELETE string.
func (ib *InsertBuilder) String() string {
	s, _ := ib.Build()
	return s
}

// Build returns compiled INSERT string and args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (ib *InsertBuilder) Build() (sql string, args []interface{}) {
	return ib.BuildWithFlavor(ib.args.Flavor)
}

// BuildWithFlavor returns compiled INSERT string and args with flavor and initial args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (ib *InsertBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
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
	return ib.args.CompileWithFlavor(buf.String(), flavor, initialArg...)
}

// SetFlavor sets the flavor of compiled sql.
func (ib *InsertBuilder) SetFlavor(flavor Flavor) (old Flavor) {
	old = ib.args.Flavor
	ib.args.Flavor = flavor
	return
}
