// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
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

func ExampleInsertBuilder_Oracle() {
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
