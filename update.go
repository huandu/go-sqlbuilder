// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"fmt"
	"strings"
)

// NewUpdateBuilder creates a new UPDATE builder.
func NewUpdateBuilder() *UpdateBuilder {
	args := &Args{}
	return &UpdateBuilder{
		Cond: Cond{
			Args: args,
		},
		args: args,
	}
}

// UpdateBuilder is a builder to build UPDATE.
//
// Usage:
//     ub := NewUpdateBuilder()
//     ub.Update("demo.user")
//         .Set(
//             ub.Assign("type", "sys"),
//             ub.Incr("credit"),
//             "modified_at = UNIX_TIMESTAMP(NOW())", // It's allowed to write arbitrary SQL.
//         )
//         .Where(
//             ub.GreaterThan("id", 1234),
//             ub.Like("name", "%Du"),
//             ub.Or(
//                 ub.IsNull("id_card"),
//                 ub.In("status", 1, 2, 5),
//             ),
//             "modified_at > created_at + " + ub.Var(86400), // It's allowed to write arbitrary SQL.
//         )
//     sql, args := ub.Build()
//     fmt.Println(sql)
//     fmt.Println(args)
//
//     // Output:
//     // UPDATE demo.user SET type = ?, credit = credit + 1, modified_at = UNIX_TIMESTAMP(NOW()) WHERE id > ? AND name LIKE ? AND (id_card IS NULL OR status IN (?, ?, ?)) AND modified_at > created_at + ?
//     // [sys, 1234, %Du, 1, 2, 5, 86400]
type UpdateBuilder struct {
	Cond

	table       string
	assignments []string
	whereExprs  []string

	args *Args
}

// Update sets table name in UPDATE.
func (ub *UpdateBuilder) Update(table string) *UpdateBuilder {
	ub.table = Escape(table)
	return ub
}

// Set sets the assignements in SET.
func (ub *UpdateBuilder) Set(assignment ...string) *UpdateBuilder {
	ub.assignments = assignment
	return ub
}

// Where sets expressions of WHERE in UPDATE.
func (ub *UpdateBuilder) Where(andExpr ...string) *UpdateBuilder {
	ub.whereExprs = append(ub.whereExprs, andExpr...)
	return ub
}

// Assign represents SET "field = value" in UPDATE.
func (ub *UpdateBuilder) Assign(field string, value interface{}) string {
	return fmt.Sprintf("%v = %v", Escape(field), ub.args.Add(value))
}

// Incr represents SET "field = field + 1" in UPDATE.
func (ub *UpdateBuilder) Incr(field string) string {
	f := Escape(field)
	return fmt.Sprintf("%v = %v + 1", f, f)
}

// Decr represents SET "field = field - 1" in UPDATE.
func (ub *UpdateBuilder) Decr(field string) string {
	f := Escape(field)
	return fmt.Sprintf("%v = %v - 1", f, f)
}

// Add represents SET "field = field + value" in UPDATE.
func (ub *UpdateBuilder) Add(field string, value interface{}) string {
	f := Escape(field)
	return fmt.Sprintf("%v = %v + %v", f, f, ub.args.Add(value))
}

// Sub represents SET "field = field - value" in UPDATE.
func (ub *UpdateBuilder) Sub(field string, value interface{}) string {
	f := Escape(field)
	return fmt.Sprintf("%v = %v - %v", f, f, ub.args.Add(value))
}

// Mul represents SET "field = field * value" in UPDATE.
func (ub *UpdateBuilder) Mul(field string, value interface{}) string {
	f := Escape(field)
	return fmt.Sprintf("%v = %v * %v", f, f, ub.args.Add(value))
}

// Div represents SET "field = field / value" in UPDATE.
func (ub *UpdateBuilder) Div(field string, value interface{}) string {
	f := Escape(field)
	return fmt.Sprintf("%v = %v / %v", f, f, ub.args.Add(value))
}

// String returns the compiled UPDATE string.
func (ub *UpdateBuilder) String() string {
	s, _ := ub.Build()
	return s
}

// Build returns compiled UPDATE string and args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (ub *UpdateBuilder) Build() (sql string, args []interface{}) {
	buf := &bytes.Buffer{}
	buf.WriteString("UPDATE ")
	buf.WriteString(ub.table)
	buf.WriteString(" SET ")
	buf.WriteString(strings.Join(ub.assignments, ", "))

	if len(ub.whereExprs) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(ub.whereExprs, " AND "))
	}

	return ub.args.Compile(buf.String())
}
