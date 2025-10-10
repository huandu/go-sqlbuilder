// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"

	"github.com/huandu/go-assert"
)

func ExampleDeleteFrom() {
	sql := DeleteFrom("demo.user").
		Where(
			"status = 1",
		).
		Limit(10).
		String()

	fmt.Println(sql)

	// Output:
	// DELETE FROM demo.user WHERE status = 1 LIMIT ?
}

func ExampleDeleteBuilder() {
	db := NewDeleteBuilder()
	db.DeleteFrom("demo.user")
	db.Where(
		db.GreaterThan("id", 1234),
		db.Like("name", "%Du"),
		db.Or(
			db.IsNull("id_card"),
			db.In("status", 1, 2, 5),
		),
		"modified_at > created_at + "+db.Var(86400), // It's allowed to write arbitrary SQL.
	)

	sql, args := db.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// DELETE FROM demo.user WHERE id > ? AND name LIKE ? AND (id_card IS NULL OR status IN (?, ?, ?)) AND modified_at > created_at + ?
	// [1234 %Du 1 2 5 86400]
}

func ExampleDeleteBuilder_SQL() {
	db := NewDeleteBuilder()
	db.SQL(`/* before */`)
	db.DeleteFrom("demo.user")
	db.SQL("PARTITION (p0)")
	db.Where(
		db.GreaterThan("id", 1234),
	)
	db.SQL("/* after where */")
	db.OrderBy("id")
	db.SQL("/* after order by */")
	db.Limit(10)
	db.SQL("/* after limit */")

	sql, args := db.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// /* before */ DELETE FROM demo.user PARTITION (p0) WHERE id > ? /* after where */ ORDER BY id /* after order by */ LIMIT ? /* after limit */
	// [1234 10]
}

func ExampleDeleteBuilder_With() {
	sql := With(
		CTEQuery("users").As(
			Select("id", "name").From("users").Where("name IS NULL"),
		),
	).DeleteFrom("orders").Where(
		"users.id = orders.user_id",
	).String()

	fmt.Println(sql)

	// Output:
	// WITH users AS (SELECT id, name FROM users WHERE name IS NULL) DELETE FROM orders WHERE users.id = orders.user_id
}

func TestDeleteBuilderGetFlavor(t *testing.T) {
	a := assert.New(t)
	db := newDeleteBuilder()

	db.SetFlavor(PostgreSQL)
	flavor := db.Flavor()
	a.Equal(PostgreSQL, flavor)

	dbClick := ClickHouse.NewDeleteBuilder()
	flavor = dbClick.Flavor()
	a.Equal(ClickHouse, flavor)
}

func ExampleDeleteBuilder_Returning() {
	db := NewDeleteBuilder()
	db.DeleteFrom("user")
	db.Where(db.Equal("id", 123))
	db.Returning("id", "deleted_at")

	sql, args := db.BuildWithFlavor(PostgreSQL)
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// DELETE FROM user WHERE id = $1 RETURNING id, deleted_at
	// [123]
}

func TestDeleteBuilderReturning(t *testing.T) {
	a := assert.New(t)
	db := NewDeleteBuilder()
	db.DeleteFrom("user")
	db.Where(db.Equal("id", 123))
	db.Returning("id", "deleted_at")

	sql, _ := db.BuildWithFlavor(MySQL)
	a.Equal("DELETE FROM user WHERE id = ?", sql)

	sql, _ = db.BuildWithFlavor(PostgreSQL)
	a.Equal("DELETE FROM user WHERE id = $1 RETURNING id, deleted_at", sql)

	sql, _ = db.BuildWithFlavor(SQLite)
	a.Equal("DELETE FROM user WHERE id = ? RETURNING id, deleted_at", sql)

	sql, _ = db.BuildWithFlavor(SQLServer)
	a.Equal("DELETE FROM user OUTPUT DELETED.id, DELETED.deleted_at WHERE id = @p1", sql)

	sql, _ = db.BuildWithFlavor(CQL)
	a.Equal("DELETE FROM user WHERE id = ?", sql)

	sql, _ = db.BuildWithFlavor(ClickHouse)
	a.Equal("DELETE FROM user WHERE id = ?", sql)

	sql, _ = db.BuildWithFlavor(Presto)
	a.Equal("DELETE FROM user WHERE id = ?", sql)

	// Test with no returning columns
	db2 := NewDeleteBuilder()
	db2.DeleteFrom("user")
	db2.Where(db2.Equal("id", 1))
	db2.Returning() // Empty returning

	sql, _ = db2.BuildWithFlavor(PostgreSQL)
	a.Equal("DELETE FROM user WHERE id = $1", sql)

	// Test with single column
	db3 := NewDeleteBuilder()
	db3.DeleteFrom("user")
	db3.Where(db3.Equal("id", 1))
	db3.Returning("id")

	sql, _ = db3.BuildWithFlavor(PostgreSQL)
	a.Equal("DELETE FROM user WHERE id = $1 RETURNING id", sql)

	// Test with ORDER BY and LIMIT
	db4 := NewDeleteBuilder()
	db4.DeleteFrom("user")
	db4.Where(db4.Equal("status", 1))
	db4.OrderBy("id").Asc()
	db4.Limit(5)
	db4.Returning("id", "name")

	sql, _ = db4.BuildWithFlavor(PostgreSQL)
	a.Equal("DELETE FROM user WHERE status = $1 ORDER BY id ASC LIMIT $2 RETURNING id, name", sql)

	// Test chaining
	db5 := NewDeleteBuilder().DeleteFrom("user").Where("status = 0").Returning("id").Returning("name", "deleted_at")
	sql, _ = db5.BuildWithFlavor(PostgreSQL)
	a.Equal("DELETE FROM user WHERE status = 0 RETURNING name, deleted_at", sql) // Last Returning call overwrites

	// Test SQL injection after RETURNING
	db6 := NewDeleteBuilder()
	db6.DeleteFrom("user")
	db6.Where(db6.Equal("id", 1))
	db6.Returning("id", "name")
	db6.SQL("/* comment after returning */")

	sql, _ = db6.BuildWithFlavor(PostgreSQL)
	a.Equal("DELETE FROM user WHERE id = $1 RETURNING id, name /* comment after returning */", sql)

	// Test with CTE (WITH clause)
	cte := With(CTETable("temp_user").As(Select("id").From("inactive_users")))
	db7 := cte.DeleteFrom("user")
	db7.Where("user.id IN (SELECT id FROM temp_user)")
	db7.Returning("id", "deleted_at")

	sql, _ = db7.BuildWithFlavor(PostgreSQL)
	a.Equal("WITH temp_user AS (SELECT id FROM inactive_users) DELETE FROM user, temp_user WHERE user.id IN (SELECT id FROM temp_user) RETURNING id, deleted_at", sql)
}

func TestDeleteBuilderClone(t *testing.T) {
	a := assert.New(t)
	cte := With(
		CTETable("temp").As(Select("id").From("to_delete")),
	)
	db := cte.DeleteFrom("target").Where("temp.id = target.id").OrderBy("id").Asc().Limit(3).Returning("id")

	clone := db.Clone()
	s1, args1 := db.BuildWithFlavor(PostgreSQL)
	s2, args2 := clone.BuildWithFlavor(PostgreSQL)
	a.Equal(s1, s2)
	a.Equal(args1, args2)

	clone.Desc().Limit(5)
	a.NotEqual(db.String(), clone.String())
}
