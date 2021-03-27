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
	// (SELECT id, name, created_at FROM demo.user WHERE id > ?) UNION ALL (TABLE demo.user_profile) ORDER BY created_at ASC LIMIT 100 OFFSET 5
	// [1234]
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
	// /* before */ (SELECT id, name, created_at FROM demo.user) UNION (SELECT id, avatar FROM demo.user_profile) /* after union */ ORDER BY created_at DESC /* after order by */ LIMIT 100 OFFSET 5 /* after limit */
}

func TestUnionForSQLite(t *testing.T) {
	a := assert.New(t)
	sb1 := Select("id", "name").From("users").Where("created_at > DATE('now', '-15 days')")
	sb2 := Select("id", "nick_name").From("user_extras").Where("status IN (1, 2, 3)")
	sql, _ := UnionAll(sb1, sb2).OrderBy("id").BuildWithFlavor(SQLite)

	a.Equal(sql, "SELECT id, name FROM users WHERE created_at > DATE('now', '-15 days') UNION ALL SELECT id, nick_name FROM user_extras WHERE status IN (1, 2, 3) ORDER BY id")
}
