// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"

	"github.com/huandu/go-assert"
)

func ExampleWhereClause() {
	// Build a SQL to select a user from database.
	sb := Select("name", "level").From("users")
	sb.Where(
		sb.Equal("id", 1234),
	)
	sql, args := sb.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Query database with the sql and update this user's level...

	ub := Update("users")
	ub.Set(
		ub.Add("level", 10),
	)

	// The WHERE clause of UPDATE should be the same as the WHERE clause of SELECT.
	ub.WhereClause = sb.WhereClause

	sql, args = ub.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT name, level FROM users WHERE id = ?
	// [1234]
	// UPDATE users SET level = level + ? WHERE id = ?
	// [10 1234]
}

func ExampleWhereClause_sharedAmongBuilders() {
	// A WhereClause can be shared among builders.
	// However, as it's not thread-safe, don't use it in a concurrent environment.
	sb1 := Select("level").From("users")
	sb2 := Select("status").From("users")

	// Share the same WhereClause between sb1 and sb2.
	whereClause := NewWhereClause()
	sb1.WhereClause = whereClause
	sb2.WhereClause = whereClause

	// The Where method in sb1 and sb2 will update the same WhereClause.
	// When we call sb1.Where(), the WHERE clause in sb2 will also be updated.
	sb1.Where(
		sb1.Like("name", "Charmy%"),
	)

	// We can get a copy of the WhereClause.
	// The copy is independent from the original.
	sb3 := Select("name").From("users")
	sb3.WhereClause = CopyWhereClause(whereClause)

	// Adding more expressions to sb1 and sb2 will not affect sb3.
	sb2.Where(
		sb2.In("status", 1, 2, 3),
	)

	// Adding more expressions to sb3 will not affect sb1 and sb2.
	sb3.Where(
		sb3.GreaterEqualThan("level", 10),
	)

	sql1, args1 := sb1.Build()
	sql2, args2 := sb2.Build()
	sql3, args3 := sb3.Build()

	fmt.Println(sql1)
	fmt.Println(args1)
	fmt.Println(sql2)
	fmt.Println(args2)
	fmt.Println(sql3)
	fmt.Println(args3)

	// Output:
	// SELECT level FROM users WHERE name LIKE ? AND status IN (?, ?, ?)
	// [Charmy% 1 2 3]
	// SELECT status FROM users WHERE name LIKE ? AND status IN (?, ?, ?)
	// [Charmy% 1 2 3]
	// SELECT name FROM users WHERE name LIKE ? AND level >= ?
	// [Charmy% 10]
}

func ExampleWhereClause_clearWhereClause() {
	db := DeleteFrom("users")
	db.Where(
		db.GreaterThan("level", 10),
	)

	sql, args := db.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Clear WHERE clause.
	db.WhereClause = nil
	sql, args = db.Build()
	fmt.Println(sql)
	fmt.Println(args)

	db.Where(
		db.Equal("id", 1234),
	)
	sql, args = db.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// DELETE FROM users WHERE level > ?
	// [10]
	// DELETE FROM users
	// []
	// DELETE FROM users WHERE id = ?
	// [1234]
}

func ExampleWhereClause_AddWhereExpr() {
	// WhereClause can be used as a standalone builder to build WHERE clause.
	// It's recommended to use it with Cond.
	whereClause := NewWhereClause()
	cond := NewCond()

	whereClause.AddWhereExpr(
		cond.Args,
		cond.In("name", "Charmy", "Huan"),
		cond.LessEqualThan("level", 10),
	)

	// Set the flavor of the WhereClause to PostgreSQL.
	whereClause.SetFlavor(PostgreSQL)

	sql, args := whereClause.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Use this WhereClause in another builder.
	sb := MySQL.NewSelectBuilder()
	sb.Select("name", "level").From("users")
	sb.WhereClause = whereClause

	// The flavor of sb overrides the flavor of the WhereClause.
	sql, args = sb.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// WHERE name IN ($1, $2) AND level <= $3
	// [Charmy Huan 10]
	// SELECT name, level FROM users WHERE name IN (?, ?) AND level <= ?
	// [Charmy Huan 10]
}

func ExampleWhereClause_AddWhereClause() {
	sb := Select("level").From("users")
	sb.Where(
		sb.Equal("id", 1234),
	)

	sql, args := sb.Build()
	fmt.Println(sql)
	fmt.Println(args)

	ub := Update("users")
	ub.Set(
		ub.Add("level", 10),
	)

	// Copy the WHERE clause of sb into ub and add more expressions.
	ub.AddWhereClause(sb.WhereClause).Where(
		ub.Equal("deleted", 0),
	)

	sql, args = ub.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT level FROM users WHERE id = ?
	// [1234]
	// UPDATE users SET level = level + ? WHERE id = ? AND deleted = ?
	// [10 1234 0]
}

func TestWhereClauseSharedInstances(t *testing.T) {
	a := assert.New(t)
	sb := Select("*").From("t")
	ub := Update("t").Set("foo = 1")
	db := DeleteFrom("t")

	whereClause := NewWhereClause()
	sb.WhereClause = whereClause
	ub.WhereClause = whereClause
	db.WhereClause = whereClause
	sb.Where(sb.Equal("id", 123))
	a.Equal(sb.String(), "SELECT * FROM t WHERE id = ?")
	a.Equal(ub.String(), "UPDATE t SET foo = 1 WHERE id = ?")
	a.Equal(db.String(), "DELETE FROM t WHERE id = ?")

	// Copied WhereClause is independent from the original.
	ub.WhereClause = CopyWhereClause(whereClause)
	ub.Where(ub.GreaterEqualThan("level", 10))
	db.Where(db.In("status", 1, 2))
	a.Equal(sb.String(), "SELECT * FROM t WHERE id = ? AND status IN (?, ?)")
	a.Equal(ub.String(), "UPDATE t SET foo = 1 WHERE id = ? AND level >= ?")
	a.Equal(db.String(), "DELETE FROM t WHERE id = ? AND status IN (?, ?)")

	// Clear the WhereClause and add new where clause and expressions.
	db.WhereClause = nil
	db.AddWhereClause(ub.WhereClause)
	db.AddWhereExpr(db.Args, db.Equal("deleted", 0))
	a.Equal(sb.String(), "SELECT * FROM t WHERE id = ? AND status IN (?, ?)")
	a.Equal(ub.String(), "UPDATE t SET foo = 1 WHERE id = ? AND level >= ?")
	a.Equal(db.String(), "DELETE FROM t WHERE id = ? AND level >= ? AND deleted = ?")

	// Nested WhereClause.
	ub.Where(ub.NotIn("id", sb))
	sb.Where(sb.NotEqual("flag", "normal"))
	a.Equal(ub.String(), "UPDATE t SET foo = 1 WHERE id = ? AND level >= ? AND id NOT IN (SELECT * FROM t WHERE id = ? AND status IN (?, ?) AND flag <> ?)")
}

func TestEmptyWhereExpr(t *testing.T) {
	a := assert.New(t)
	var emptyExpr []string
	sb := Select("*").From("t").Where(emptyExpr...)
	ub := Update("t").Set("foo = 1").Where(emptyExpr...)
	db := DeleteFrom("t").Where(emptyExpr...)

	a.Equal(sb.String(), "SELECT * FROM t")
	a.Equal(ub.String(), "UPDATE t SET foo = 1")
	a.Equal(db.String(), "DELETE FROM t")
}
