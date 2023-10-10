// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql"
	"fmt"
)

func ExampleSelect() {
	// Build a SQL to create a HIVE table.
	s := CreateTable("users").
		SQL("PARTITION BY (year)").
		SQL("AS").
		SQL(
			Select("columns[0] id", "columns[1] name", "columns[2] year").
				From("`all-users.csv`").
				Limit(100).
				String(),
		).
		String()

	fmt.Println(s)

	// Output:
	// CREATE TABLE users PARTITION BY (year) AS SELECT columns[0] id, columns[1] name, columns[2] year FROM `all-users.csv` LIMIT 100
}

func ExampleSelectBuilder() {
	sb := NewSelectBuilder()
	sb.Distinct().Select("id", "name", sb.As("COUNT(*)", "t"))
	sb.From("demo.user")
	sb.Where(
		sb.GreaterThan("id", 1234),
		sb.Like("name", "%Du"),
		sb.Or(
			sb.IsNull("id_card"),
			sb.In("status", 1, 2, 5),
		),
		sb.NotIn(
			"id",
			NewSelectBuilder().Select("id").From("banned"),
		), // Nested SELECT.
		"modified_at > created_at + "+sb.Var(86400), // It's allowed to write arbitrary SQL.
	)
	sb.GroupBy("status").Having(sb.NotIn("status", 4, 5))
	sb.OrderBy("modified_at").Asc()
	sb.Limit(10).Offset(5)

	s, args := sb.Build()
	fmt.Println(s)
	fmt.Println(args)

	// Output:
	// SELECT DISTINCT id, name, COUNT(*) AS t FROM demo.user WHERE id > ? AND name LIKE ? AND (id_card IS NULL OR status IN (?, ?, ?)) AND id NOT IN (SELECT id FROM banned) AND modified_at > created_at + ? GROUP BY status HAVING status NOT IN (?, ?) ORDER BY modified_at ASC LIMIT 10 OFFSET 5
	// [1234 %Du 1 2 5 86400 4 5]
}

func ExampleSelectBuilder_advancedUsage() {
	sb := NewSelectBuilder()
	innerSb := NewSelectBuilder()

	sb.Select("id", "name")
	sb.From(
		sb.BuilderAs(innerSb, "user"),
	)
	sb.Where(
		sb.In("status", Flatten([]int{1, 2, 3})...),
		sb.Between("created_at", sql.Named("start", 1234567890), sql.Named("end", 1234599999)),
	)
	sb.OrderBy("modified_at").Desc()

	innerSb.Select("*")
	innerSb.From("banned")
	innerSb.Where(
		innerSb.NotIn("name", Flatten([]string{"Huan Du", "Charmy Liu"})...),
	)

	s, args := sb.Build()
	fmt.Println(s)
	fmt.Println(args)

	// Output:
	// SELECT id, name FROM (SELECT * FROM banned WHERE name NOT IN (?, ?)) AS user WHERE status IN (?, ?, ?) AND created_at BETWEEN @start AND @end ORDER BY modified_at DESC
	// [Huan Du Charmy Liu 1 2 3 {{} start 1234567890} {{} end 1234599999}]
}

func ExampleSelectBuilder_join() {
	sb := NewSelectBuilder()
	sb.Select("u.id", "u.name", "c.type", "p.nickname")
	sb.From("user u")
	sb.Join("contract c",
		"u.id = c.user_id",
		sb.In("c.status", 1, 2, 5),
	)
	sb.JoinWithOption(RightOuterJoin, "person p",
		"u.id = p.user_id",
		sb.Like("p.surname", "%Du"),
	)
	sb.Where(
		"u.modified_at > u.created_at + " + sb.Var(86400), // It's allowed to write arbitrary SQL.
	)

	sql, args := sb.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT u.id, u.name, c.type, p.nickname FROM user u JOIN contract c ON u.id = c.user_id AND c.status IN (?, ?, ?) RIGHT OUTER JOIN person p ON u.id = p.user_id AND p.surname LIKE ? WHERE u.modified_at > u.created_at + ?
	// [1 2 5 %Du 86400]
}

func ExampleSelectBuilder_limit_offset() {
	flavors := []Flavor{MySQL, PostgreSQL, SQLite, SQLServer, CQL, ClickHouse, Presto, Oracle}
	results := make([][]string, len(flavors))
	sb := NewSelectBuilder()
	saveResults := func() {
		for i, f := range flavors {
			s, _ := sb.BuildWithFlavor(f)
			results[i] = append(results[i], s)
		}
	}

	sb.Select("*")
	sb.From("user")

	// Case #1: limit < 0 and offset < 0
	//
	// All: No limit or offset in query.
	sb.Limit(-1)
	sb.Offset(-1)
	saveResults()

	// Case #2: limit < 0 and offset >= 0
	//
	// MySQL and SQLite: Ignore offset if the limit is not set.
	// PostgreSQL: Offset can be set without limit.
	// SQLServer: Offset can be set without limit.
	// CQL: Ignore offset.
	// Oracle: Offset can be set without limit.
	sb.Limit(-1)
	sb.Offset(0)
	saveResults()

	// Case #3: limit >= 0 and offset >= 0
	//
	// CQL: Ignore offset.
	// All others: Set both limit and offset.
	sb.Limit(1)
	sb.Offset(0)
	saveResults()

	// Case #4: limit >= 0 and offset < 0
	//
	// All: Set limit in query.
	sb.Limit(1)
	sb.Offset(-1)
	saveResults()

	// Case #5: limit >= 0 and offset >= 0 order by id
	//
	// CQL: Ignore offset.
	// All others: Set both limit and offset.
	sb.Limit(1)
	sb.Offset(1)
	sb.OrderBy("id")
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
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user
	// #3: SELECT * FROM user LIMIT 1 OFFSET 0
	// #4: SELECT * FROM user LIMIT 1
	// #5: SELECT * FROM user ORDER BY id LIMIT 1 OFFSET 1
	//
	// PostgreSQL
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user OFFSET 0
	// #3: SELECT * FROM user LIMIT 1 OFFSET 0
	// #4: SELECT * FROM user LIMIT 1
	// #5: SELECT * FROM user ORDER BY id LIMIT 1 OFFSET 1
	//
	// SQLite
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user
	// #3: SELECT * FROM user LIMIT 1 OFFSET 0
	// #4: SELECT * FROM user LIMIT 1
	// #5: SELECT * FROM user ORDER BY id LIMIT 1 OFFSET 1
	//
	// SQLServer
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user ORDER BY 1 OFFSET 0 ROWS
	// #3: SELECT * FROM user ORDER BY 1 OFFSET 0 ROWS FETCH NEXT 1 ROWS ONLY
	// #4: SELECT * FROM user ORDER BY 1 OFFSET 0 ROWS FETCH NEXT 1 ROWS ONLY
	// #5: SELECT * FROM user ORDER BY id OFFSET 1 ROWS FETCH NEXT 1 ROWS ONLY
	//
	// CQL
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user
	// #3: SELECT * FROM user LIMIT 1
	// #4: SELECT * FROM user LIMIT 1
	// #5: SELECT * FROM user ORDER BY id LIMIT 1
	//
	// ClickHouse
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user
	// #3: SELECT * FROM user LIMIT 1 OFFSET 0
	// #4: SELECT * FROM user LIMIT 1
	// #5: SELECT * FROM user ORDER BY id LIMIT 1 OFFSET 1
	//
	// Presto
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user OFFSET 0
	// #3: SELECT * FROM user LIMIT 1 OFFSET 0
	// #4: SELECT * FROM user LIMIT 1
	// #5: SELECT * FROM user ORDER BY id LIMIT 1 OFFSET 1
	//
	// Oracle
	// #1: SELECT * FROM user
	// #2: SELECT * FROM ( SELECT ROWNUM r, * FROM ( SELECT * FROM user ) user ) WHERE r >= 1
	// #3: SELECT * FROM ( SELECT ROWNUM r, * FROM ( SELECT * FROM user ) user ) WHERE r BETWEEN 1 AND 1
	// #4: SELECT * FROM ( SELECT ROWNUM r, * FROM ( SELECT * FROM user ) user ) WHERE r BETWEEN 1 AND 1
	// #5: SELECT * FROM ( SELECT ROWNUM r, * FROM ( SELECT * FROM user ORDER BY id ) user ) WHERE r BETWEEN 2 AND 2
}

func ExampleSelectBuilder_ForUpdate() {
	sb := newSelectBuilder()
	sb.Select("*").From("user").Where(
		sb.Equal("id", 1234),
	).ForUpdate()

	sql, args := sb.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT * FROM user WHERE id = ? FOR UPDATE
	// [1234]
}

func ExampleSelectBuilder_varInCols() {
	// Column name may contain some characters, e.g. the $ sign, which have special meanings in builders.
	// It's recommended to call Escape() or EscapeAll() to escape the name.

	sb := NewSelectBuilder()
	v := sb.Var("foo")
	sb.Select(Escape("colHasA$Sign"), v)
	sb.From("table")

	s, args := sb.Build()
	fmt.Println(s)
	fmt.Println(args)

	// Output:
	// SELECT colHasA$Sign, ? FROM table
	// [foo]
}

func ExampleSelectBuilder_SQL() {
	sb := NewSelectBuilder()
	sb.SQL("/* before */")
	sb.Select("u.id", "u.name", "c.type", "p.nickname")
	sb.SQL("/* after select */")
	sb.From("user u")
	sb.SQL("/* after from */")
	sb.Join("contract c",
		"u.id = c.user_id",
	)
	sb.JoinWithOption(RightOuterJoin, "person p",
		"u.id = p.user_id",
	)
	sb.SQL("/* after join */")
	sb.Where(
		"u.modified_at > u.created_at",
	)
	sb.SQL("/* after where */")
	sb.OrderBy("id")
	sb.SQL("/* after order by */")
	sb.Limit(10)
	sb.SQL("/* after limit */")
	sb.ForShare()
	sb.SQL("/* after for */")

	s := sb.String()
	fmt.Println(s)

	// Output:
	// /* before */ SELECT u.id, u.name, c.type, p.nickname /* after select */ FROM user u /* after from */ JOIN contract c ON u.id = c.user_id RIGHT OUTER JOIN person p ON u.id = p.user_id /* after join */ WHERE u.modified_at > u.created_at /* after where */ ORDER BY id /* after order by */ LIMIT 10 /* after limit */ FOR SHARE /* after for */
}

// Example for issue #115.
func ExampleSelectBuilder_customSELECT() {
	sb := NewSelectBuilder()

	// Set a custom SELECT clause.
	sb.SQL("SELECT id, name FROM user").Where(
		sb.In("id", 1, 2, 3),
	)

	s, args := sb.Build()
	fmt.Println(s)
	fmt.Println(args)

	// Output:
	// SELECT id, name FROM user WHERE id IN (?, ?, ?)
	// [1 2 3]
}
