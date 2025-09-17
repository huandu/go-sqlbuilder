// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"

	"github.com/huandu/go-assert"
)

func ExampleUnion() {
	sb1 := NewSelectBuilder()
	sb1.Select("id", "name", "created_at")
	sb1.From("demo.user")
	sb1.Where(
		sb1.GreaterThan("id", 1234),
	)

	sb2 := newSelectBuilder()
	sb2.Select("id", "avatar")
	sb2.From("demo.user_profile")
	sb2.Where(
		sb2.In("status", 1, 2, 5),
	)

	ub := Union(sb1, sb2)
	ub.OrderBy("created_at").Desc()

	sql, args := ub.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// (SELECT id, name, created_at FROM demo.user WHERE id > ?) UNION (SELECT id, avatar FROM demo.user_profile WHERE status IN (?, ?, ?)) ORDER BY created_at DESC
	// [1234 1 2 5]
}

func ExampleUnionAll() {
	sb := NewSelectBuilder()
	sb.Select("id", "name", "created_at")
	sb.From("demo.user")
	sb.Where(
		sb.GreaterThan("id", 1234),
	)

	ub := UnionAll(sb, Build("TABLE demo.user_profile"))
	ub.OrderBy("created_at").Asc()
	ub.Limit(100).Offset(5)

	sql, args := ub.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// (SELECT id, name, created_at FROM demo.user WHERE id > ?) UNION ALL (TABLE demo.user_profile) ORDER BY created_at ASC LIMIT ? OFFSET ?
	// [1234 100 5]
}

func ExampleUnionBuilder_SQL() {
	sb1 := NewSelectBuilder()
	sb1.Select("id", "name", "created_at")
	sb1.From("demo.user")

	sb2 := newSelectBuilder()
	sb2.Select("id", "avatar")
	sb2.From("demo.user_profile")

	ub := NewUnionBuilder()
	ub.SQL("/* before */")
	ub.Union(sb1, sb2)
	ub.SQL("/* after union */")
	ub.OrderBy("created_at").Desc()
	ub.SQL("/* after order by */")
	ub.Limit(100).Offset(5)
	ub.SQL("/* after limit */")

	sql := ub.String()
	fmt.Println(sql)

	// Output:
	// /* before */ (SELECT id, name, created_at FROM demo.user) UNION (SELECT id, avatar FROM demo.user_profile) /* after union */ ORDER BY created_at DESC /* after order by */ LIMIT ? OFFSET ? /* after limit */
}

func TestUnionForSQLite(t *testing.T) {
	a := assert.New(t)
	sb1 := Select("id", "name").From("users").Where("created_at > DATE('now', '-15 days')")
	sb2 := Select("id", "nick_name").From("user_extras").Where("status IN (1, 2, 3)")
	sql, _ := UnionAll(sb1, sb2).OrderBy("id").Limit(100).Offset(5).BuildWithFlavor(SQLite)

	a.Equal(sql, "SELECT id, name FROM users WHERE created_at > DATE('now', '-15 days') UNION ALL SELECT id, nick_name FROM user_extras WHERE status IN (1, 2, 3) ORDER BY id LIMIT ? OFFSET ?")
}

func TestUnionBuilderGetFlavor(t *testing.T) {
	a := assert.New(t)
	ub := newUnionBuilder()

	ub.SetFlavor(PostgreSQL)
	flavor := ub.Flavor()
	a.Equal(PostgreSQL, flavor)

	ubClick := ClickHouse.NewUnionBuilder()
	flavor = ubClick.Flavor()
	a.Equal(ClickHouse, flavor)
}

func ExampleUnionBuilder_limit_offset() {
	flavors := []Flavor{MySQL, PostgreSQL, SQLite, SQLServer, CQL, ClickHouse, Presto, Oracle, Informix, Doris}
	results := make([][]string, len(flavors))

	ub := NewUnionBuilder()
	saveResults := func() {
		sb1 := NewSelectBuilder()
		sb1.Select("*").From("user1")
		sb2 := NewSelectBuilder()
		sb2.Select("*").From("user2")
		ub.Union(sb1, sb2)
		for i, f := range flavors {
			s, _ := ub.BuildWithFlavor(f)
			results[i] = append(results[i], s)
		}
	}

	// Case #1: limit < 0 and offset < 0
	//
	// All: No limit or offset in query.
	ub.Limit(-1)
	ub.Offset(-1)
	saveResults()

	// Case #2: limit < 0 and offset >= 0
	//
	// MySQL and SQLite: Ignore offset if the limit is not set.
	// PostgreSQL: Offset can be set without limit.
	// SQLServer: Offset can be set without limit.
	// CQL: Ignore offset.
	// Oracle: Offset can be set without limit.
	ub.Limit(-1)
	ub.Offset(0)
	saveResults()

	// Case #3: limit >= 0 and offset >= 0
	//
	// CQL: Ignore offset.
	// All others: Set both limit and offset.
	ub.Limit(1)
	ub.Offset(0)
	saveResults()

	// Case #4: limit >= 0 and offset < 0
	//
	// All: Set limit in query.
	ub.Limit(1)
	ub.Offset(-1)
	saveResults()

	// Case #5: limit >= 0 and offset >= 0 order by id
	//
	// CQL: Ignore offset.
	// All others: Set both limit and offset.
	ub.Limit(1)
	ub.Offset(1)
	ub.OrderBy("id")
	saveResults()

	for i, result := range results {
		fmt.Println()
		fmt.Println(flavors[i])

		for n, s := range result {
			fmt.Printf("#%d: %s\n", n+1, s)
		}
	}

	// Output:
	//
	// MySQL
	// #1: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #2: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #3: (SELECT * FROM user1) UNION (SELECT * FROM user2) LIMIT ? OFFSET ?
	// #4: (SELECT * FROM user1) UNION (SELECT * FROM user2) LIMIT ?
	// #5: (SELECT * FROM user1) UNION (SELECT * FROM user2) ORDER BY id LIMIT ? OFFSET ?
	//
	// PostgreSQL
	// #1: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #2: (SELECT * FROM user1) UNION (SELECT * FROM user2) OFFSET $1
	// #3: (SELECT * FROM user1) UNION (SELECT * FROM user2) LIMIT $1 OFFSET $2
	// #4: (SELECT * FROM user1) UNION (SELECT * FROM user2) LIMIT $1
	// #5: (SELECT * FROM user1) UNION (SELECT * FROM user2) ORDER BY id LIMIT $1 OFFSET $2
	//
	// SQLite
	// #1: SELECT * FROM user1 UNION SELECT * FROM user2
	// #2: SELECT * FROM user1 UNION SELECT * FROM user2
	// #3: SELECT * FROM user1 UNION SELECT * FROM user2 LIMIT ? OFFSET ?
	// #4: SELECT * FROM user1 UNION SELECT * FROM user2 LIMIT ?
	// #5: SELECT * FROM user1 UNION SELECT * FROM user2 ORDER BY id LIMIT ? OFFSET ?
	//
	// SQLServer
	// #1: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #2: (SELECT * FROM user1) UNION (SELECT * FROM user2) ORDER BY 1 OFFSET @p1 ROWS
	// #3: (SELECT * FROM user1) UNION (SELECT * FROM user2) ORDER BY 1 OFFSET @p1 ROWS FETCH NEXT @p2 ROWS ONLY
	// #4: (SELECT * FROM user1) UNION (SELECT * FROM user2) ORDER BY 1 OFFSET 0 ROWS FETCH NEXT @p1 ROWS ONLY
	// #5: (SELECT * FROM user1) UNION (SELECT * FROM user2) ORDER BY id OFFSET @p1 ROWS FETCH NEXT @p2 ROWS ONLY
	//
	// CQL
	// #1: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #2: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #3: (SELECT * FROM user1) UNION (SELECT * FROM user2) LIMIT ?
	// #4: (SELECT * FROM user1) UNION (SELECT * FROM user2) LIMIT ?
	// #5: (SELECT * FROM user1) UNION (SELECT * FROM user2) ORDER BY id LIMIT ?
	//
	// ClickHouse
	// #1: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #2: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #3: (SELECT * FROM user1) UNION (SELECT * FROM user2) LIMIT ? OFFSET ?
	// #4: (SELECT * FROM user1) UNION (SELECT * FROM user2) LIMIT ?
	// #5: (SELECT * FROM user1) UNION (SELECT * FROM user2) ORDER BY id LIMIT ? OFFSET ?
	//
	// Presto
	// #1: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #2: (SELECT * FROM user1) UNION (SELECT * FROM user2) OFFSET ?
	// #3: (SELECT * FROM user1) UNION (SELECT * FROM user2) OFFSET ? LIMIT ?
	// #4: (SELECT * FROM user1) UNION (SELECT * FROM user2) LIMIT ?
	// #5: (SELECT * FROM user1) UNION (SELECT * FROM user2) ORDER BY id OFFSET ? LIMIT ?
	//
	// Oracle
	// #1: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #2: SELECT * FROM ( (SELECT * FROM user1) UNION (SELECT * FROM user2) ) OFFSET :1 ROWS
	// #3: SELECT * FROM ( (SELECT * FROM user1) UNION (SELECT * FROM user2) ) OFFSET :1 ROWS FETCH NEXT :2 ROWS ONLY
	// #4: SELECT * FROM ( (SELECT * FROM user1) UNION (SELECT * FROM user2) ) OFFSET 0 ROWS FETCH NEXT :1 ROWS ONLY
	// #5: SELECT * FROM ( (SELECT * FROM user1) UNION (SELECT * FROM user2) ) ORDER BY id OFFSET :1 ROWS FETCH NEXT :2 ROWS ONLY
	//
	// Informix
	// #1: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #2: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #3: SELECT * FROM ( (SELECT * FROM user1) UNION (SELECT * FROM user2) ) SKIP ? FIRST ?
	// #4: SELECT * FROM ( (SELECT * FROM user1) UNION (SELECT * FROM user2) ) FIRST ?
	// #5: SELECT * FROM ( (SELECT * FROM user1) UNION (SELECT * FROM user2) ) ORDER BY id SKIP ? FIRST ?
	//
	// Doris
	// #1: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #2: (SELECT * FROM user1) UNION (SELECT * FROM user2)
	// #3: (SELECT * FROM user1) UNION (SELECT * FROM user2) LIMIT 1 OFFSET 0
	// #4: (SELECT * FROM user1) UNION (SELECT * FROM user2) LIMIT 1
	// #5: (SELECT * FROM user1) UNION (SELECT * FROM user2) ORDER BY id LIMIT 1 OFFSET 1
}
