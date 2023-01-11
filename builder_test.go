// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/huandu/go-assert"
)

func ExampleBuildf() {
	sb := NewSelectBuilder()
	sb.Select("id").From("user")

	explain := Buildf("EXPLAIN %v LEFT JOIN SELECT * FROM banned WHERE state IN (%v, %v)", sb, 1, 2)
	s, args := explain.Build()
	fmt.Println(s)
	fmt.Println(args)

	// Output:
	// EXPLAIN SELECT id FROM user LEFT JOIN SELECT * FROM banned WHERE state IN (?, ?)
	// [1 2]
}

func ExampleBuild() {
	sb := NewSelectBuilder()
	sb.Select("id").From("user").Where(sb.In("status", 1, 2))

	b := Build("EXPLAIN $? LEFT JOIN SELECT * FROM $? WHERE created_at > $? AND state IN (${states}) AND modified_at BETWEEN $2 AND $?",
		sb, Raw("banned"), 1514458225, 1514544625, Named("states", List([]int{3, 4, 5})))
	s, args := b.Build()

	fmt.Println(s)
	fmt.Println(args)

	// Output:
	// EXPLAIN SELECT id FROM user WHERE status IN (?, ?) LEFT JOIN SELECT * FROM banned WHERE created_at > ? AND state IN (?, ?, ?) AND modified_at BETWEEN ? AND ?
	// [1 2 1514458225 3 4 5 1514458225 1514544625]
}

func ExampleBuildNamed() {
	b := BuildNamed("SELECT * FROM ${table} WHERE status IN (${status}) AND name LIKE ${name} AND created_at > ${time} AND modified_at < ${time} + 86400",
		map[string]interface{}{
			"time":   sql.Named("start", 1234567890),
			"status": List([]int{1, 2, 5}),
			"name":   "Huan%",
			"table":  Raw("user"),
		})
	s, args := b.Build()

	fmt.Println(s)
	fmt.Println(args)

	// Output:
	// SELECT * FROM user WHERE status IN (?, ?, ?) AND name LIKE ? AND created_at > @start AND modified_at < @start + 86400
	// [1 2 5 Huan% {{} start 1234567890}]
}

func ExampleWithFlavor() {
	sql, args := WithFlavor(Buildf("SELECT * FROM foo WHERE id = %v", 1234), PostgreSQL).Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Explicitly use MySQL as the flavor.
	sql, args = WithFlavor(Buildf("SELECT * FROM foo WHERE id = %v", 1234), PostgreSQL).BuildWithFlavor(MySQL)

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT * FROM foo WHERE id = $1
	// [1234]
	// SELECT * FROM foo WHERE id = ?
	// [1234]
}

func TestBuildWithPostgreSQL(t *testing.T) {
	a := assert.New(t)
	sb1 := PostgreSQL.NewSelectBuilder()
	sb1.Select("col1", "col2").From("t1").Where(sb1.E("id", 1234), sb1.G("level", 2))

	sb2 := PostgreSQL.NewSelectBuilder()
	sb2.Select("col3", "col4").From("t2").Where(sb2.E("id", 4567), sb2.LE("level", 5))

	// Use DefaultFlavor (MySQL) instead of PostgreSQL.
	sql, args := Build("SELECT $1 AS col5 LEFT JOIN $0 LEFT JOIN $2", sb1, 7890, sb2).Build()

	a.Equal(sql, "SELECT ? AS col5 LEFT JOIN SELECT col1, col2 FROM t1 WHERE id = ? AND level > ? LEFT JOIN SELECT col3, col4 FROM t2 WHERE id = ? AND level <= ?")
	a.Equal(args, []interface{}{7890, 1234, 2, 4567, 5})

	old := DefaultFlavor
	DefaultFlavor = PostgreSQL
	defer func() {
		DefaultFlavor = old
	}()

	sql, args = Build("SELECT $1 AS col5 LEFT JOIN $0 LEFT JOIN $2", sb1, 7890, sb2).Build()

	a.Equal(sql, "SELECT $1 AS col5 LEFT JOIN SELECT col1, col2 FROM t1 WHERE id = $2 AND level > $3 LEFT JOIN SELECT col3, col4 FROM t2 WHERE id = $4 AND level <= $5")
	a.Equal(args, []interface{}{7890, 1234, 2, 4567, 5})
}

func TestBuildWithCQL(t *testing.T) {
	a := assert.New(t)

	ib1 := CQL.NewInsertBuilder()
	ib1.InsertInto("t1").Cols("col1", "col2").Values(1, 2)

	ib2 := CQL.NewInsertBuilder()
	ib2.InsertInto("t2").Cols("col3", "col4").Values(3, 4)

	old := DefaultFlavor
	DefaultFlavor = CQL
	defer func() {
		DefaultFlavor = old
	}()

	sql, args := Build("BEGIN BATCH USING TIMESTAMP $0 $1; $2; APPLY BATCH;", 1481124356754405, ib1, ib2).Build()

	a.Equal(sql, "BEGIN BATCH USING TIMESTAMP ? INSERT INTO t1 (col1, col2) VALUES (?, ?); INSERT INTO t2 (col3, col4) VALUES (?, ?); APPLY BATCH;")
	a.Equal(args, []interface{}{1481124356754405, 1, 2, 3, 4})
}
