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
//
// Usage:
//     ib := NewInsertBuilder()
//     ib.InsertInto("demo.user")
//         .Cols("id", "name", "status", "created_at")
//         .Values(1, "Huan Du", 1, ib.Raw("UNIX_TIMESTAMP(NOW())"))
//         .Values(2, "Charmy Liu", 1, 1234567890)
//     sql, args := ib.Build()
//     fmt.Println(sql)
//     fmt.Println(args)
//
//     // Output:
//     // INSERT INTO demo.user (id, name, status, creasted_at) VALEUS (?, ?, ?, ?), (?, ?, ?, ?)
//     // [1, Huan Du, 1, 2, Charmy Liu, 1, 1234567890]
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

// String returns the compiled UPDATE string.
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
		buf.WriteString("(")
		buf.WriteString(strings.Join(ib.cols, ", "))
		buf.WriteString(")")
	}

	buf.WriteString(" VALUES ")
	values := make([]string, 0, len(ib.values))

	for _, v := range ib.values {
		values = append(values, fmt.Sprintf("(%v)", strings.Join(v, ", ")))
	}

	buf.WriteString(strings.Join(values, ", "))
	return ib.args.Compile(buf.String())
}
