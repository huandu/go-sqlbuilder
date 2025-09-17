// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"

	"github.com/huandu/go-assert"
)

func ExampleInsertInto() {
	sql, args := InsertInto("demo.user").
		Cols("id", "name", "status").
		Values(4, "Sample", 2).
		Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO demo.user (id, name, status) VALUES (?, ?, ?)
	// [4 Sample 2]
}

func ExampleInsertIgnoreInto() {
	sql, args := InsertIgnoreInto("demo.user").
		Cols("id", "name", "status").
		Values(4, "Sample", 2).
		Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT IGNORE INTO demo.user (id, name, status) VALUES (?, ?, ?)
	// [4 Sample 2]
}

func ExampleReplaceInto() {
	sql, args := ReplaceInto("demo.user").
		Cols("id", "name", "status").
		Values(4, "Sample", 2).
		Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// REPLACE INTO demo.user (id, name, status) VALUES (?, ?, ?)
	// [4 Sample 2]
}

func ExampleInsertBuilder() {
	ib := NewInsertBuilder()
	ib.InsertInto("demo.user")
	ib.Cols("id", "name", "status", "created_at", "updated_at")
	ib.Values(1, "Huan Du", 1, Raw("UNIX_TIMESTAMP(NOW())"))
	ib.Values(2, "Charmy Liu", 1, 1234567890)

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO demo.user (id, name, status, created_at, updated_at) VALUES (?, ?, ?, UNIX_TIMESTAMP(NOW())), (?, ?, ?, ?)
	// [1 Huan Du 1 2 Charmy Liu 1 1234567890]
}

func ExampleInsertBuilder_flavorOracle() {
	ib := Oracle.NewInsertBuilder()
	ib.InsertInto("demo.user")
	ib.Cols("id", "name", "status")
	ib.Values(1, "Huan Du", 1)
	ib.Values(2, "Charmy Liu", 1)

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT ALL INTO demo.user (id, name, status) VALUES (:1, :2, :3) INTO demo.user (id, name, status) VALUES (:4, :5, :6) SELECT 1 from DUAL
	// [1 Huan Du 1 2 Charmy Liu 1]
}

func ExampleInsertBuilder_insertIgnore() {
	ib := NewInsertBuilder()
	ib.InsertIgnoreInto("demo.user")
	ib.Cols("id", "name", "status", "created_at", "updated_at")
	ib.Values(1, "Huan Du", 1, Raw("UNIX_TIMESTAMP(NOW())"))
	ib.Values(2, "Charmy Liu", 1, 1234567890)

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT IGNORE INTO demo.user (id, name, status, created_at, updated_at) VALUES (?, ?, ?, UNIX_TIMESTAMP(NOW())), (?, ?, ?, ?)
	// [1 Huan Du 1 2 Charmy Liu 1 1234567890]
}

func ExampleInsertBuilder_insertIgnore_postgres() {
	ib := PostgreSQL.NewInsertBuilder()
	ib.InsertIgnoreInto("demo.user")
	ib.Cols("id", "name", "status", "created_at")
	ib.Values(1, "Huan Du", 1, Raw("UNIX_TIMESTAMP(NOW())"))
	ib.Values(2, "Charmy Liu", 1, 1234567890)

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO demo.user (id, name, status, created_at) VALUES ($1, $2, $3, UNIX_TIMESTAMP(NOW())), ($4, $5, $6, $7) ON CONFLICT DO NOTHING
	// [1 Huan Du 1 2 Charmy Liu 1 1234567890]
}

func ExampleInsertBuilder_insertIgnore_sqlite() {
	ib := SQLite.NewInsertBuilder()
	ib.InsertIgnoreInto("demo.user")
	ib.Cols("id", "name", "status", "created_at")
	ib.Values(1, "Huan Du", 1, Raw("UNIX_TIMESTAMP(NOW())"))
	ib.Values(2, "Charmy Liu", 1, 1234567890)

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT OR IGNORE INTO demo.user (id, name, status, created_at) VALUES (?, ?, ?, UNIX_TIMESTAMP(NOW())), (?, ?, ?, ?)
	// [1 Huan Du 1 2 Charmy Liu 1 1234567890]
}

func ExampleInsertBuilder_insertIgnore_clickhouse() {
	ib := ClickHouse.NewInsertBuilder()
	ib.InsertIgnoreInto("demo.user")
	ib.Cols("id", "name", "status", "created_at")
	ib.Values(1, "Huan Du", 1, Raw("UNIX_TIMESTAMP(NOW())"))
	ib.Values(2, "Charmy Liu", 1, 1234567890)

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO demo.user (id, name, status, created_at) VALUES (?, ?, ?, UNIX_TIMESTAMP(NOW())), (?, ?, ?, ?)
	// [1 Huan Du 1 2 Charmy Liu 1 1234567890]
}

func ExampleInsertBuilder_replaceInto() {
	ib := NewInsertBuilder()
	ib.ReplaceInto("demo.user")
	ib.Cols("id", "name", "status", "created_at", "updated_at")
	ib.Values(1, "Huan Du", 1, Raw("UNIX_TIMESTAMP(NOW())"))
	ib.Values(2, "Charmy Liu", 1, 1234567890)

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// REPLACE INTO demo.user (id, name, status, created_at, updated_at) VALUES (?, ?, ?, UNIX_TIMESTAMP(NOW())), (?, ?, ?, ?)
	// [1 Huan Du 1 2 Charmy Liu 1 1234567890]
}

func ExampleInsertBuilder_SQL() {
	ib := NewInsertBuilder()
	ib.SQL("/* before */")
	ib.InsertInto("demo.user")
	ib.SQL("PARTITION (p0)")
	ib.Cols("id", "name", "status", "created_at")
	ib.SQL("/* after cols */")
	ib.Values(3, "Shawn Du", 1, 1234567890)
	ib.SQL(ib.Var(Build("ON DUPLICATE KEY UPDATE status = $?", 1)))

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// /* before */ INSERT INTO demo.user PARTITION (p0) (id, name, status, created_at) /* after cols */ VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE status = ?
	// [3 Shawn Du 1 1234567890 1]
}

func ExampleInsertBuilder_subSelect() {
	ib := NewInsertBuilder()
	ib.InsertInto("demo.user")
	ib.Cols("id", "name")
	sb := ib.Select("id", "name").From("demo.test")
	sb.Where(sb.EQ("id", 1))

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO demo.user (id, name) SELECT id, name FROM demo.test WHERE id = ?
	// [1]
}

func ExampleInsertBuilder_subSelect_oracle() {
	ib := Oracle.NewInsertBuilder()
	ib.InsertInto("demo.user")
	ib.Cols("id", "name")
	sb := ib.Select("id", "name").From("demo.test")
	sb.Where(sb.EQ("id", 1))

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO demo.user (id, name) SELECT id, name FROM demo.test WHERE id = :1
	// [1]
}

func ExampleInsertBuilder_subSelect_informix() {
	ib := Informix.NewInsertBuilder()
	ib.InsertInto("demo.user")
	ib.Cols("id", "name")
	sb := ib.Select("id", "name").From("demo.test")
	sb.Where(sb.EQ("id", 1))

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO demo.user (id, name) SELECT id, name FROM demo.test WHERE id = ?
	// [1]
}

func ExampleInsertBuilder_NumValue() {
	ib := NewInsertBuilder()
	ib.InsertInto("demo.user")
	ib.Cols("id", "name")
	ib.Values(1, "Huan Du")
	ib.Values(2, "Charmy Liu")

	// Count the number of values.
	fmt.Println(ib.NumValue())

	// Output:
	// 2
}

func ExampleInsertBuilder_Returning() {
	sql, args := InsertInto("user").
		Cols("name").Values("Huan Du").
		Returning("id").
		BuildWithFlavor(PostgreSQL)

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO user (name) VALUES ($1) RETURNING id
	// [Huan Du]
}

func TestInsertBuilderReturning(test *testing.T) {
	a := assert.New(test)
	ib := InsertInto("user").
		Cols("name").Values("Huan Du").
		Returning("id")

	sql, _ := ib.BuildWithFlavor(MySQL)
	a.Equal("INSERT INTO user (name) VALUES (?)", sql)

	sql, _ = ib.BuildWithFlavor(PostgreSQL)
	a.Equal("INSERT INTO user (name) VALUES ($1) RETURNING id", sql)

	sql, _ = ib.BuildWithFlavor(SQLite)
	a.Equal("INSERT INTO user (name) VALUES (?) RETURNING id", sql)

	sql, _ = ib.BuildWithFlavor(SQLServer)
	a.Equal("INSERT INTO user (name) VALUES (@p1)", sql)

	sql, _ = ib.BuildWithFlavor(CQL)
	a.Equal("INSERT INTO user (name) VALUES (?)", sql)

	sql, _ = ib.BuildWithFlavor(ClickHouse)
	a.Equal("INSERT INTO user (name) VALUES (?)", sql)

	sql, _ = ib.BuildWithFlavor(Presto)
	a.Equal("INSERT INTO user (name) VALUES (?)", sql)

	sql, _ = ib.BuildWithFlavor(Oracle)
	a.Equal("INSERT INTO user (name) VALUES (:1)", sql)

	sql, _ = ib.BuildWithFlavor(Informix)
	a.Equal("INSERT INTO user (name) VALUES (?)", sql)
}

func TestInsertBuilderGetFlavor(t *testing.T) {
	a := assert.New(t)
	ib := newInsertBuilder()

	ib.SetFlavor(PostgreSQL)
	flavor := ib.Flavor()
	a.Equal(PostgreSQL, flavor)

	ibClick := ClickHouse.NewInsertBuilder()
	flavor = ibClick.Flavor()
	a.Equal(ClickHouse, flavor)
}

func TestIssue200(t *testing.T) {
	a := assert.New(t)
	ib := PostgreSQL.NewInsertBuilder()
	ib.InsertIgnoreInto("table")
	ib.Cols("data")
	sb := ib.Select("id", "data").From("table")
	sb.Where(sb.Equal("id", 1))

	query, _ := ib.Build()
	a.Equal(query, "INSERT INTO table (data) SELECT id, data FROM table WHERE id = $1 ON CONFLICT DO NOTHING")
}

func TestInsertBuilderClone(t *testing.T) {
	a := assert.New(t)

	ib := InsertInto("demo.user").Cols("id", "name")
	ib.Values(1, "A")

	clone := ib.Clone()

	s1, args1 := ib.Build()
	s2, args2 := clone.Build()
	a.Equal(s1, s2)
	a.Equal(args1, args2)

	// mutate clone and verify original unchanged
	clone.Values(2, "B")
	a.NotEqual(ib.String(), clone.String())
}
