// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/huandu/go-assert"
)

func ExampleSelect() {
	// Build a SQL to create a HIVE table using MySQL-like SQL syntax.
	sql, args := Select("columns[0] id", "columns[1] name", "columns[2] year").
		From(MySQL.Quote("all-users.csv")).
		Limit(100).
		Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT columns[0] id, columns[1] name, columns[2] year FROM `all-users.csv` LIMIT ?
	// [100]
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
	sb.OrderByAsc("modified_at")
	sb.Limit(10).Offset(5)

	s, args := sb.Build()
	fmt.Println(s)
	fmt.Println(args)

	// Output:
	// SELECT DISTINCT id, name, COUNT(*) AS t FROM demo.user WHERE id > ? AND name LIKE ? AND (id_card IS NULL OR status IN (?, ?, ?)) AND id NOT IN (SELECT id FROM banned) AND modified_at > created_at + ? GROUP BY status HAVING status NOT IN (?, ?) ORDER BY modified_at ASC LIMIT ? OFFSET ?
	// [1234 %Du 1 2 5 86400 4 5 10 5]
}

func ExampleSelectBuilder_advancedUsage() {
	sb := NewSelectBuilder()
	innerSb := NewSelectBuilder()

	// Named arguments are supported.
	start := sql.Named("start", 1234567890)
	end := sql.Named("end", 1234599999)
	level := sql.Named("level", 20)

	sb.Select("id", "name")
	sb.From(
		sb.BuilderAs(innerSb, "user"),
	)
	sb.Where(
		sb.In("status", Flatten([]int{1, 2, 3})...),
		sb.Between("created_at", start, end),
	)
	sb.OrderByDesc("modified_at")

	innerSb.Select("*")
	innerSb.From("banned")
	innerSb.Where(
		innerSb.GreaterThan("level", level),
		innerSb.LessEqualThan("updated_at", end),
		innerSb.NotIn("name", Flatten([]string{"Huan Du", "Charmy Liu"})...),
	)

	s, args := sb.Build()
	fmt.Println(s)
	fmt.Println(args)

	// Output:
	// SELECT id, name FROM (SELECT * FROM banned WHERE level > @level AND updated_at <= @end AND name NOT IN (?, ?)) AS user WHERE status IN (?, ?, ?) AND created_at BETWEEN @start AND @end ORDER BY modified_at DESC
	// [Huan Du Charmy Liu 1 2 3 {{} level 20} {{} end 1234599999} {{} start 1234567890}]
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

func ExampleSelectBuilder_nestedJoin() {
	sb := NewSelectBuilder()
	nestedSb := NewSelectBuilder()

	// Build the nested subquery
	nestedSb.Select("b.id", "b.user_id")
	nestedSb.From("users2 AS b")
	nestedSb.Where(nestedSb.GreaterThan("b.age", 20))

	// Build the main query with nested join
	sb.Select("a.id", "a.user_id")
	sb.From("users AS a")
	sb.Join(
		sb.BuilderAs(nestedSb, "b"),
		"a.user_id = b.user_id",
	)

	sql, args := sb.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT a.id, a.user_id FROM users AS a JOIN (SELECT b.id, b.user_id FROM users2 AS b WHERE b.age > ?) AS b ON a.user_id = b.user_id
	// [20]
}

func ExampleSelectBuilder_limit_offset() {
	flavors := []Flavor{MySQL, PostgreSQL, SQLite, SQLServer, CQL, ClickHouse, Presto, Oracle, Informix, Doris}
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
	// #3: SELECT * FROM user LIMIT ? OFFSET ?
	// #4: SELECT * FROM user LIMIT ?
	// #5: SELECT * FROM user ORDER BY id LIMIT ? OFFSET ?
	//
	// PostgreSQL
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user OFFSET $1
	// #3: SELECT * FROM user LIMIT $1 OFFSET $2
	// #4: SELECT * FROM user LIMIT $1
	// #5: SELECT * FROM user ORDER BY id LIMIT $1 OFFSET $2
	//
	// SQLite
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user
	// #3: SELECT * FROM user LIMIT ? OFFSET ?
	// #4: SELECT * FROM user LIMIT ?
	// #5: SELECT * FROM user ORDER BY id LIMIT ? OFFSET ?
	//
	// SQLServer
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user ORDER BY 1 OFFSET @p1 ROWS
	// #3: SELECT * FROM user ORDER BY 1 OFFSET @p1 ROWS FETCH NEXT @p2 ROWS ONLY
	// #4: SELECT * FROM user ORDER BY 1 OFFSET 0 ROWS FETCH NEXT @p1 ROWS ONLY
	// #5: SELECT * FROM user ORDER BY id OFFSET @p1 ROWS FETCH NEXT @p2 ROWS ONLY
	//
	// CQL
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user
	// #3: SELECT * FROM user LIMIT ?
	// #4: SELECT * FROM user LIMIT ?
	// #5: SELECT * FROM user ORDER BY id LIMIT ?
	//
	// ClickHouse
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user
	// #3: SELECT * FROM user LIMIT ? OFFSET ?
	// #4: SELECT * FROM user LIMIT ?
	// #5: SELECT * FROM user ORDER BY id LIMIT ? OFFSET ?
	//
	// Presto
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user OFFSET ?
	// #3: SELECT * FROM user OFFSET ? LIMIT ?
	// #4: SELECT * FROM user LIMIT ?
	// #5: SELECT * FROM user ORDER BY id OFFSET ? LIMIT ?
	//
	// Oracle
	// #1: SELECT * FROM user
	// #2: SELECT * FROM (SELECT ROWNUM r, * FROM (SELECT * FROM user) user) WHERE r >= :1 + 1
	// #3: SELECT * FROM (SELECT ROWNUM r, * FROM (SELECT * FROM user) user) WHERE r BETWEEN :1 + 1 AND :2 + :3
	// #4: SELECT * FROM (SELECT ROWNUM r, * FROM (SELECT * FROM user) user) WHERE r BETWEEN 1 AND :1 + 1
	// #5: SELECT * FROM (SELECT ROWNUM r, * FROM (SELECT * FROM user ORDER BY id) user) WHERE r BETWEEN :1 + 1 AND :2 + :3
	//
	// Informix
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user
	// #3: SELECT * FROM user SKIP ? FIRST ?
	// #4: SELECT * FROM user FIRST ?
	// #5: SELECT * FROM user ORDER BY id SKIP ? FIRST ?
	//
	// Doris
	// #1: SELECT * FROM user
	// #2: SELECT * FROM user
	// #3: SELECT * FROM user LIMIT 1 OFFSET 0
	// #4: SELECT * FROM user LIMIT 1
	// #5: SELECT * FROM user ORDER BY id LIMIT 1 OFFSET 1
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
	// /* before */ SELECT u.id, u.name, c.type, p.nickname /* after select */ FROM user u /* after from */ JOIN contract c ON u.id = c.user_id RIGHT OUTER JOIN person p ON u.id = p.user_id /* after join */ WHERE u.modified_at > u.created_at /* after where */ ORDER BY id /* after order by */ LIMIT ? /* after limit */ FOR SHARE /* after for */
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

func ExampleSelectBuilder_NumCol() {
	sb := NewSelectBuilder()
	sb.Select("id", "name", "created_at")
	sb.From("demo.user")
	sb.Where(
		sb.GreaterThan("id", 1234),
	)

	// Count the number of columns.
	fmt.Println(sb.NumCol())

	// Output:
	// 3
}

func ExampleSelectBuilder_With() {
	sql := With(
		CTEQuery("users").As(
			Select("id", "name").From("users").Where("prime IS NOT NULL"),
		),

		// The CTE table orders will be added to table list of FROM clause automatically.
		CTETable("orders").As(
			Select("id", "user_id").From("orders"),
		),
	).Select("orders.id").Join("users", "orders.user_id = users.id").Limit(10).String()

	fmt.Println(sql)

	// Output:
	// WITH users AS (SELECT id, name FROM users WHERE prime IS NOT NULL), orders AS (SELECT id, user_id FROM orders) SELECT orders.id FROM orders JOIN users ON orders.user_id = users.id LIMIT ?
}

func TestSelectBuilderSelectMore(t *testing.T) {
	a := assert.New(t)
	sb := Select("id").SQL("/* first */").Where(
		"name IS NOT NULL",
	).SQL("/* second */").SelectMore("name").SQL("/* third */")
	a.Equal(sb.String(), "SELECT id, name /* first */ /* third */ WHERE name IS NOT NULL /* second */")
}

func TestSelectBuilderGetFlavor(t *testing.T) {
	a := assert.New(t)
	sb := newSelectBuilder()

	sb.SetFlavor(PostgreSQL)
	flavor := sb.Flavor()
	a.Equal(PostgreSQL, flavor)

	sbClick := ClickHouse.NewSelectBuilder()
	flavor = sbClick.Flavor()
	a.Equal(ClickHouse, flavor)
}

func ExampleSelectBuilder_LateralAs() {
	// Demo SQL comes from a sample on https://dev.mysql.com/doc/refman/8.4/en/lateral-derived-tables.html.
	sb := Select(
		"salesperson.name",
		"max_sale.amount",
		"max_sale.customer_name",
	)
	sb.From(
		"salesperson",
		sb.LateralAs(
			Select("amount", "customer_name").
				From("all_sales").
				Where(
					"all_sales.salesperson_id = salesperson.id",
				).
				OrderByDesc("amount").Limit(1),
			"max_sale",
		),
	)

	fmt.Println(sb)

	// Output:
	// SELECT salesperson.name, max_sale.amount, max_sale.customer_name FROM salesperson, LATERAL (SELECT amount, customer_name FROM all_sales WHERE all_sales.salesperson_id = salesperson.id ORDER BY amount DESC LIMIT ?) AS max_sale
}

func TestNilPointerWhere(t *testing.T) {
	NewSelectBuilder().SQL("$0").Build()
	NewSelectBuilder().SQL("$0").BuildWithFlavor(DefaultFlavor)
}

func TestSelectBuilderClone(t *testing.T) {
	a := assert.New(t)

	cte := With(
		CTETable("users").As(
			Select("id", "name").From("users").Where("name IS NOT NULL"),
		),
	)

	sb := cte.Select("users.id", "orders.id").From("orders").Where(
		"users.id = orders.user_id",
	).OrderBy("orders.id").Desc().Limit(10).Offset(2)

	// Clone and compare
	clone := sb.Clone()
	s1, args1 := sb.Build()
	s2, args2 := clone.Build()
	a.Equal(s1, s2)
	a.Equal(args1, args2)

	// Mutate clone and ensure original unchanged
	clone.Limit(100).Asc()
	s1After := sb.String()
	s2After := clone.String()
	a.NotEqual(s1After, s2After)
}

func ExampleSelectBuilder_OrderByAsc() {
	sb := NewSelectBuilder()
	sb.Select("id", "name", "score")
	sb.From("users")
	sb.Where(sb.GreaterThan("score", 0))
	sb.OrderByAsc("name")

	sql, args := sb.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT id, name, score FROM users WHERE score > ? ORDER BY name ASC
	// [0]
}

func ExampleSelectBuilder_OrderByDesc() {
	sb := NewSelectBuilder()
	sb.Select("id", "name", "score")
	sb.From("users")
	sb.Where(sb.GreaterThan("score", 0))
	sb.OrderByDesc("score")

	sql, args := sb.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT id, name, score FROM users WHERE score > ? ORDER BY score DESC
	// [0]
}

func ExampleSelectBuilder_OrderByAsc_multiple() {
	sb := NewSelectBuilder()
	sb.Select("id", "name", "score")
	sb.From("users")
	sb.Where(sb.GreaterThan("score", 0))
	// Chain multiple OrderByAsc and OrderByDesc calls with different directions
	sb.OrderByDesc("score").OrderByAsc("name").OrderByDesc("id")

	sql, args := sb.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT id, name, score FROM users WHERE score > ? ORDER BY score DESC, name ASC, id DESC
	// [0]
}

func TestSelectBuilder_OrderByAscDesc(t *testing.T) {
	a := assert.New(t)

	// Test OrderByAsc with single column
	sb := NewSelectBuilder()
	sb.Select("*").From("users").OrderByAsc("name")
	sql, _ := sb.Build()
	a.Equal("SELECT * FROM users ORDER BY name ASC", sql)

	// Test OrderByDesc with single column
	sb = NewSelectBuilder()
	sb.Select("*").From("users").OrderByDesc("id")
	sql, _ = sb.Build()
	a.Equal("SELECT * FROM users ORDER BY id DESC", sql)

	// Test chaining OrderByAsc and OrderByDesc
	sb = NewSelectBuilder()
	sb.Select("*").From("users")
	sb.OrderByDesc("score").OrderByAsc("name")
	sql, _ = sb.Build()
	a.Equal("SELECT * FROM users ORDER BY score DESC, name ASC", sql)

	// Test multiple OrderByDesc calls
	sb = NewSelectBuilder()
	sb.Select("*").From("users")
	sb.OrderByDesc("score").OrderByDesc("id")
	sql, _ = sb.Build()
	a.Equal("SELECT * FROM users ORDER BY score DESC, id DESC", sql)

	// Test multiple OrderByAsc calls
	sb = NewSelectBuilder()
	sb.Select("*").From("users")
	sb.OrderByAsc("name").OrderByAsc("email")
	sql, _ = sb.Build()
	a.Equal("SELECT * FROM users ORDER BY name ASC, email ASC", sql)

	// Test mixed ordering with more complex scenario
	sb = NewSelectBuilder()
	sb.Select("id", "name", "score", "created_at").From("users")
	sb.Where(sb.GreaterThan("score", 0))
	sb.OrderByDesc("score").OrderByAsc("name").OrderByDesc("created_at")
	sql, args := sb.Build()
	a.Equal("SELECT id, name, score, created_at FROM users WHERE score > ? ORDER BY score DESC, name ASC, created_at DESC", sql)
	a.Equal([]interface{}{0}, args)

	// Test that OrderByAsc/OrderByDesc work with table aliases
	sb = NewSelectBuilder()
	sb.Select("u.id", "u.name", "o.total").From("users u")
	sb.Join("orders o", "u.id = o.user_id")
	sb.OrderByDesc("o.total").OrderByAsc("u.name")
	sql, _ = sb.Build()
	a.Equal("SELECT u.id, u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id ORDER BY o.total DESC, u.name ASC", sql)
}
