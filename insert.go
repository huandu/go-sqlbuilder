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
		verb: "INSERT",
		args: args,
	}
}

// InsertBuilder is a builder to build INSERT.
type InsertBuilder struct {
	verb   string
	table  string
	cols   []string
	values [][]string
	upsert bool

	args *Args
}

var _ Builder = new(InsertBuilder)

// InsertInto sets table name in INSERT.
func (ib *InsertBuilder) InsertInto(table string) *InsertBuilder {
	ib.table = Escape(table)
	return ib
}

// InsertIgnoreInto sets table name in INSERT IGNORE.
func (ib *InsertBuilder) InsertIgnoreInto(table string) *InsertBuilder {
	ib.verb = "INSERT IGNORE"
	ib.table = Escape(table)
	return ib
}

// ReplaceInto sets table name and changes the verb of ib to REPLACE.
// REPLACE INTO is a MySQL extension to the SQL standard.
func (ib *InsertBuilder) ReplaceInto(table string) *InsertBuilder {
	ib.verb = "REPLACE"
	ib.table = Escape(table)
	return ib
}

func (ib *InsertBuilder) UpsertInto(table string) *InsertBuilder {
	ib.table = Escape(table)
	ib.upsert = true
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

// String returns the compiled INSERT string.
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
	buf.WriteString(ib.verb)
	buf.WriteString(" INTO ")
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

	if ib.upsert {
		buf.WriteString(strings.Join(values, ", "))
		buf.WriteString(" ON DUPLICATE KEY UPDATE ")

		values = make([]string, 0, len(ib.cols))
		for _, col := range ib.cols {
			// Use syntax as in MySQL 5.7: https://dev.mysql.com/doc/refman/5.7/en/insert-on-duplicate.html
			values = append(values, fmt.Sprintf("%s = VALUES(%s)", col, col))
		}
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
