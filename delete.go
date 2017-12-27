// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"strings"
)

// NewDeleteBuilder creates a new DELETE builder.
func NewDeleteBuilder() *DeleteBuilder {
	args := &Args{}
	return &DeleteBuilder{
		Cond: Cond{
			Args: args,
		},
		args: args,
	}
}

// DeleteBuilder is a builder to build DELETE.
//
// Usage:
//     db := NewDeleteBuilder()
//     db.DeleteFrom("demo.user")
//         .Where(
//             db.GreaterThan("id", 1234),
//             db.Like("name", "%Du"),
//             db.Or(
//                 db.IsNull("id_card"),
//                 db.In("status", 1, 2, 5),
//             ),
//             "modified_at > created_at + " + db.Var(86400), // It's allowed to write arbitrary SQL.
//         )
//     sql, args := db.Build()
//     fmt.Println(sql)
//     fmt.Println(args)
//
//     // Output:
//     // DELETE FROM demo.user WHERE id > ? AND name LIKE ? AND (id_card IS NULL OR status IN (?, ?, ?)) AND modified_at > created_at + ?
//     // [1234, %Du, 1, 2, 5, 86400]
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

// Where sets expressions of WHERE in UPDATE.
func (db *DeleteBuilder) Where(andExpr ...string) *DeleteBuilder {
	db.whereExprs = append(db.whereExprs, andExpr...)
	return db
}

// String returns the compiled UPDATE string.
func (db *DeleteBuilder) String() string {
	s, _ := db.Build()
	return s
}

// Build returns compiled DELETE string and args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (db *DeleteBuilder) Build() (sql string, args []interface{}) {
	buf := &bytes.Buffer{}
	buf.WriteString("DELETE FROM ")
	buf.WriteString(db.table)

	if len(db.whereExprs) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(db.whereExprs, " AND "))
	}

	return db.args.Compile(buf.String())
}
