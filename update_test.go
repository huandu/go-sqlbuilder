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
		"f = f + $1|[123]": func(ub *UpdateBuilder) string { return ub.Add("f", 123) },
		"f = f - $1|[123]": func(ub *UpdateBuilder) string { return ub.Sub("f", 123) },
		"f = f * $1|[123]": func(ub *UpdateBuilder) string { return ub.Mul("f", 123) },
		"f = f / $1|[123]": func(ub *UpdateBuilder) string { return ub.Div("f", 123) },
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
	// /* before */ UPDATE demo.user /* after update */ SET type = ? /* after set */ ORDER BY id DESC /* after order by */ LIMIT 10 /* after limit */
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
		CTEQuery("users").As(
			Select("id", "name").From("users").Where("prime IS NOT NULL"),
		),
	).Update("orders").Set(
		"orders.transport_fee = 0",
	).Where(
		"users.id = orders.user_id",
	).String()

	fmt.Println(sql)

	// Output:
	// WITH users AS (SELECT id, name FROM users WHERE prime IS NOT NULL) UPDATE orders SET orders.transport_fee = 0 WHERE users.id = orders.user_id
}
