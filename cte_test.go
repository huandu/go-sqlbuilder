// Copyright 2024 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"

	"github.com/huandu/go-assert"
)

func ExampleWith() {
	sb := With(
		CTETable("users", "id", "name").As(
			Select("id", "name").From("users").Where("name IS NOT NULL"),
		),
		CTETable("devices").As(
			Select("device_id").From("devices"),
		),
	).Select("users.id", "orders.id", "devices.device_id").Join(
		"orders",
		"users.id = orders.user_id",
		"devices.device_id = orders.device_id",
	)

	fmt.Println(sb)

	// Output:
	// WITH users (id, name) AS (SELECT id, name FROM users WHERE name IS NOT NULL), devices AS (SELECT device_id FROM devices) SELECT users.id, orders.id, devices.device_id FROM users, devices JOIN orders ON users.id = orders.user_id AND devices.device_id = orders.device_id
}

func ExampleCTEBuilder() {
	usersBuilder := Select("id", "name", "level").From("users")
	usersBuilder.Where(
		usersBuilder.GreaterEqualThan("level", 10),
	)
	cteb := With(
		CTETable("valid_users").As(usersBuilder),
	)
	fmt.Println(cteb)

	sb := Select("valid_users.id", "valid_users.name", "orders.id").With(cteb)
	sb.Join("orders", "valid_users.id = orders.user_id")
	sb.Where(
		sb.LessEqualThan("orders.price", 200),
		"valid_users.level < orders.min_level",
	).OrderBy("orders.price").Desc()

	sql, args := sb.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// WITH valid_users AS (SELECT id, name, level FROM users WHERE level >= ?)
	// WITH valid_users AS (SELECT id, name, level FROM users WHERE level >= ?) SELECT valid_users.id, valid_users.name, orders.id FROM valid_users JOIN orders ON valid_users.id = orders.user_id WHERE orders.price <= ? AND valid_users.level < orders.min_level ORDER BY orders.price DESC
	// [10 200]
}

func TestCTEBuilder(t *testing.T) {
	a := assert.New(t)
	cteb := newCTEBuilder()
	ctetb := newCTETableBuilder()
	cteb.SQL("/* init */")
	cteb.With(ctetb)
	cteb.SQL("/* after with */")

	ctetb.SQL("/* table init */")
	ctetb.Table("t", "a", "b")
	ctetb.SQL("/* after table */")

	ctetb.As(Select("a", "b").From("t"))
	ctetb.SQL("/* after table as */")

	sql, args := cteb.Build()
	a.Equal(sql, "/* init */ WITH /* table init */ t (a, b) /* after table */ AS (SELECT a, b FROM t) /* after table as */ /* after with */")
	a.Assert(args == nil)

	sql = ctetb.String()
	a.Equal(sql, "/* table init */ t (a, b) /* after table */ AS (SELECT a, b FROM t) /* after table as */")
}
