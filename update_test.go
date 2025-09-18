// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"

	"github.com/huandu/go-assert"
)

func ExampleUpdate() {
	sql := Update("demo.user").
		Set(
			"visited = visited + 1",
		).
		Where(
			"id = 1234",
		).
		String()

	fmt.Println(sql)

	// Output:
	// UPDATE demo.user SET visited = visited + 1 WHERE id = 1234
}

func ExampleUpdateBuilder() {
	ub := NewUpdateBuilder()
	ub.Update("demo.user")
	ub.Set(
		ub.Assign("type", "sys"),
		ub.Incr("credit"),
		"modified_at = UNIX_TIMESTAMP(NOW())", // It's allowed to write arbitrary SQL.
	)
	ub.Where(
		ub.GreaterThan("id", 1234),
		ub.Like("name", "%Du"),
		ub.Or(
			ub.IsNull("id_card"),
			ub.In("status", 1, 2, 5),
		),
		"modified_at > created_at + "+ub.Var(86400), // It's allowed to write arbitrary SQL.
	)
	ub.OrderBy("id").Asc()

	sql, args := ub.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// UPDATE demo.user SET type = ?, credit = credit + 1, modified_at = UNIX_TIMESTAMP(NOW()) WHERE id > ? AND name LIKE ? AND (id_card IS NULL OR status IN (?, ?, ?)) AND modified_at > created_at + ? ORDER BY id ASC
	// [sys 1234 %Du 1 2 5 86400]
}

func TestUpdateAssignments(t *testing.T) {
	a := assert.New(t)
	cases := map[string]func(ub *UpdateBuilder) string{
		"f = f + 1|[]":     func(ub *UpdateBuilder) string { return ub.Incr("f") },
		"f = f - 1|[]":     func(ub *UpdateBuilder) string { return ub.Decr("f") },
		"f = f + $0|[123]": func(ub *UpdateBuilder) string { return ub.Add("f", 123) },
		"f = f - $0|[123]": func(ub *UpdateBuilder) string { return ub.Sub("f", 123) },
		"f = f * $0|[123]": func(ub *UpdateBuilder) string { return ub.Mul("f", 123) },
		"f = f / $0|[123]": func(ub *UpdateBuilder) string { return ub.Div("f", 123) },
	}

	for expected, f := range cases {
		ub := NewUpdateBuilder()
		s := f(ub)
		ub.Set(s)
		_, args := ub.Build()
		actual := fmt.Sprintf("%v|%v", s, args)

		a.Equal(actual, expected)
	}
}

func ExampleUpdateBuilder_SetMore() {
	ub := NewUpdateBuilder()
	ub.Update("demo.user")
	ub.Set(
		ub.Assign("type", "sys"),
		ub.Incr("credit"),
	)
	ub.SetMore(
		"modified_at = UNIX_TIMESTAMP(NOW())", // It's allowed to write arbitrary SQL.
	)

	sql, args := ub.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// UPDATE demo.user SET type = ?, credit = credit + 1, modified_at = UNIX_TIMESTAMP(NOW())
	// [sys]
}

func ExampleUpdateBuilder_SQL() {
	ub := NewUpdateBuilder()
	ub.SQL("/* before */")
	ub.Update("demo.user")
	ub.SQL("/* after update */")
	ub.Set(
		ub.Assign("type", "sys"),
	)
	ub.SQL("/* after set */")
	ub.OrderBy("id").Desc()
	ub.SQL("/* after order by */")
	ub.Limit(10)
	ub.SQL("/* after limit */")

	sql := ub.String()
	fmt.Println(sql)

	// Output:
	// /* before */ UPDATE demo.user /* after update */ SET type = ? /* after set */ ORDER BY id DESC /* after order by */ LIMIT ? /* after limit */
}

func ExampleUpdateBuilder_NumAssignment() {
	ub := NewUpdateBuilder()
	ub.Update("demo.user")
	ub.Set(
		ub.Assign("type", "sys"),
		ub.Incr("credit"),
		"modified_at = UNIX_TIMESTAMP(NOW())",
	)

	// Count the number of assignments.
	fmt.Println(ub.NumAssignment())

	// Output:
	// 3
}

func ExampleUpdateBuilder_With() {
	sql := With(
		CTETable("users").As(
			Select("id", "name").From("users").Where("prime IS NOT NULL"),
		),
	).Update("orders").Set(
		"orders.transport_fee = 0",
	).Where(
		"users.id = orders.user_id",
	).String()

	fmt.Println(sql)

	// Output:
	// WITH users AS (SELECT id, name FROM users WHERE prime IS NOT NULL) UPDATE orders, users SET orders.transport_fee = 0 WHERE users.id = orders.user_id
}

func TestUpdateBuilderGetFlavor(t *testing.T) {
	a := assert.New(t)
	ub := newUpdateBuilder()

	ub.SetFlavor(PostgreSQL)
	flavor := ub.Flavor()
	a.Equal(PostgreSQL, flavor)

	ubClick := ClickHouse.NewUpdateBuilder()
	flavor = ubClick.Flavor()
	a.Equal(ClickHouse, flavor)
}

func ExampleUpdateBuilder_Returning() {
	ub := NewUpdateBuilder()
	ub.Update("user")
	ub.Set(ub.Assign("name", "Huan Du"))
	ub.Where(ub.Equal("id", 123))
	ub.Returning("id", "updated_at")

	sql, args := ub.BuildWithFlavor(PostgreSQL)
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// UPDATE user SET name = $1 WHERE id = $2 RETURNING id, updated_at
	// [Huan Du 123]
}

func TestUpdateBuilderReturning(t *testing.T) {
	a := assert.New(t)
	ub := NewUpdateBuilder()
	ub.Update("user")
	ub.Set(ub.Assign("name", "Huan Du"))
	ub.Where(ub.Equal("id", 123))
	ub.Returning("id", "updated_at")

	sql, _ := ub.BuildWithFlavor(MySQL)
	a.Equal("UPDATE user SET name = ? WHERE id = ?", sql)

	sql, _ = ub.BuildWithFlavor(PostgreSQL)
	a.Equal("UPDATE user SET name = $1 WHERE id = $2 RETURNING id, updated_at", sql)

	sql, _ = ub.BuildWithFlavor(SQLite)
	a.Equal("UPDATE user SET name = ? WHERE id = ? RETURNING id, updated_at", sql)

	sql, _ = ub.BuildWithFlavor(SQLServer)
	a.Equal("UPDATE user SET name = @p1 WHERE id = @p2", sql)

	sql, _ = ub.BuildWithFlavor(CQL)
	a.Equal("UPDATE user SET name = ? WHERE id = ?", sql)

	sql, _ = ub.BuildWithFlavor(ClickHouse)
	a.Equal("UPDATE user SET name = ? WHERE id = ?", sql)

	sql, _ = ub.BuildWithFlavor(Presto)
	a.Equal("UPDATE user SET name = ? WHERE id = ?", sql)

	// Test with no returning columns
	ub2 := NewUpdateBuilder()
	ub2.Update("user")
	ub2.Set(ub2.Assign("name", "Test"))
	ub2.Where(ub2.Equal("id", 1))
	ub2.Returning() // Empty returning

	sql, _ = ub2.BuildWithFlavor(PostgreSQL)
	a.Equal("UPDATE user SET name = $1 WHERE id = $2", sql)

	// Test with single column
	ub3 := NewUpdateBuilder()
	ub3.Update("user")
	ub3.Set(ub3.Assign("name", "Test"))
	ub3.Where(ub3.Equal("id", 1))
	ub3.Returning("id")

	sql, _ = ub3.BuildWithFlavor(PostgreSQL)
	a.Equal("UPDATE user SET name = $1 WHERE id = $2 RETURNING id", sql)

	// Test with ORDER BY and LIMIT
	ub4 := NewUpdateBuilder()
	ub4.Update("user")
	ub4.Set(ub4.Assign("name", "Test"))
	ub4.Where(ub4.Equal("status", 1))
	ub4.OrderBy("id").Asc()
	ub4.Limit(5)
	ub4.Returning("id", "name")

	sql, _ = ub4.BuildWithFlavor(PostgreSQL)
	a.Equal("UPDATE user SET name = $1 WHERE status = $2 ORDER BY id ASC LIMIT $3 RETURNING id, name", sql)

	// Test chaining
	ub5 := NewUpdateBuilder().Update("user").Set("status = 1").Returning("id").Returning("name", "updated_at")
	sql, _ = ub5.BuildWithFlavor(PostgreSQL)
	a.Equal("UPDATE user SET status = 1 RETURNING name, updated_at", sql) // Last Returning call overwrites

	// Test SQL injection after RETURNING
	ub6 := NewUpdateBuilder()
	ub6.Update("user")
	ub6.Set(ub6.Assign("name", "Test"))
	ub6.Where(ub6.Equal("id", 1))
	ub6.Returning("id", "name")
	ub6.SQL("/* comment after returning */")

	sql, _ = ub6.BuildWithFlavor(PostgreSQL)
	a.Equal("UPDATE user SET name = $1 WHERE id = $2 RETURNING id, name /* comment after returning */", sql)

	// Test with CTE (WITH clause)
	cte := With(CTETable("temp_user").As(Select("id").From("active_users")))
	ub7 := cte.Update("user")
	ub7.Set(ub7.Assign("status", "active"))
	ub7.Where("user.id IN (SELECT id FROM temp_user)")
	ub7.Returning("id", "status")

	sql, _ = ub7.BuildWithFlavor(PostgreSQL)
	a.Equal("WITH temp_user AS (SELECT id FROM active_users) UPDATE user SET status = $1 FROM temp_user WHERE user.id IN (SELECT id FROM temp_user) RETURNING id, status", sql)
}

func TestUpdateBuilderClone(t *testing.T) {
	a := assert.New(t)
	cte := With(
		CTETable("vip").As(Select("user_id").From("vip_users")),
	)
	ub := cte.Update("orders").Set("discount = 1").Where("orders.user_id = vip.user_id").OrderBy("orders.id").Desc().Limit(2).Returning("orders.id")

	clone := ub.Clone()
	s1, args1 := ub.BuildWithFlavor(PostgreSQL)
	s2, args2 := clone.BuildWithFlavor(PostgreSQL)
	a.Equal(s1, s2)
	a.Equal(args1, args2)

	clone.Asc().Limit(5)
	a.NotEqual(ub.String(), clone.String())
}
