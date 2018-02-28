// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"strings"
)

// NewDeleteBuilder creates a new DELETE builder.
func NewDeleteBuilder() *DeleteBuilder {
	return DefaultFlavor.NewDeleteBuilder()
}

func newDeleteBuilder() *DeleteBuilder {
	args := &Args{}
	return &DeleteBuilder{
		Cond: Cond{
			Args: args,
		},
		args: args,
	}
}

// DeleteBuilder is a builder to build DELETE.
type DeleteBuilder struct {
	Cond

	table      string
	whereExprs []string

	args *Args
}

// DeleteFrom sets table name in DELETE.
func (db *DeleteBuilder) DeleteFrom(table string) *DeleteBuilder {
	db.table = Escape(table)
	return db
}

// Where sets expressions of WHERE in DELETE.
func (db *DeleteBuilder) Where(andExpr ...string) *DeleteBuilder {
	db.whereExprs = append(db.whereExprs, andExpr...)
	return db
}

// String returns the compiled DELETE string.
func (db *DeleteBuilder) String() string {
	s, _ := db.Build()
	return s
}

// Build returns compiled DELETE string and args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (db *DeleteBuilder) Build() (sql string, args []interface{}) {
	return db.BuildWithFlavor(db.args.Flavor)
}

// BuildWithFlavor returns compiled DELETE string and args with flavor and initial args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (db *DeleteBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
	buf := &bytes.Buffer{}
	buf.WriteString("DELETE FROM ")
	buf.WriteString(db.table)

	if len(db.whereExprs) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(db.whereExprs, " AND "))
	}

	return db.args.CompileWithFlavor(buf.String(), flavor, initialArg...)
}

// SetFlavor sets the flavor of compiled sql.
func (db *DeleteBuilder) SetFlavor(flavor Flavor) (old Flavor) {
	old = db.args.Flavor
	db.args.Flavor = flavor
	return
}
