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

func ExampleWithRecursive() {
	sb := WithRecursive(
		CTEQuery("source_accounts", "id", "parent_id").As(
			UnionAll(
				Select("p.id", "p.parent_id").
					From("accounts AS p").
					Where("p.id = 2"), // Show orders for account 2 and all its child accounts
				Select("c.id", "c.parent_id").
					From("accounts AS c").
					Join("source_accounts AS sa", "c.parent_id = sa.id"),
			),
		),
	).Select("o.id", "o.date", "o.amount").
		From("orders AS o").
		Join("source_accounts", "o.account_id = source_accounts.id")

	fmt.Println(sb)

	// Output:
	// WITH RECURSIVE source_accounts (id, parent_id) AS ((SELECT p.id, p.parent_id FROM accounts AS p WHERE p.id = 2) UNION ALL (SELECT c.id, c.parent_id FROM accounts AS c JOIN source_accounts AS sa ON c.parent_id = sa.id)) SELECT o.id, o.date, o.amount FROM orders AS o JOIN source_accounts ON o.account_id = source_accounts.id
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

	sb := Select("valid_users.id", "valid_users.name", "orders.id").
		From("users").With(cteb).
		Join("orders", "users.id = orders.user_id")
	sb.Where(
		sb.LessEqualThan("orders.price", 200),
		"valid_users.level < orders.min_level",
	).OrderBy("orders.price").Desc()

	sql, args := sb.Build()
	fmt.Println(sql)
	fmt.Println(args)
	fmt.Println(sb.TableNames())

	// Output:
	// WITH valid_users AS (SELECT id, name, level FROM users WHERE level >= ?)
	// WITH valid_users AS (SELECT id, name, level FROM users WHERE level >= ?) SELECT valid_users.id, valid_users.name, orders.id FROM users, valid_users JOIN orders ON users.id = orders.user_id WHERE orders.price <= ? AND valid_users.level < orders.min_level ORDER BY orders.price DESC
	// [10 200]
	// [users valid_users]
}

func ExampleCTEBuilder_update() {
	builder := With(
		CTETable("users", "user_id").As(
			Select("user_id").From("vip_users"),
		),
	).Update("orders").Set(
		"orders.transport_fee = 0",
	).Where(
		"users.user_id = orders.user_id",
	)

	sqlForMySQL, _ := builder.BuildWithFlavor(MySQL)
	sqlForPostgreSQL, _ := builder.BuildWithFlavor(PostgreSQL)

	fmt.Println(sqlForMySQL)
	fmt.Println(sqlForPostgreSQL)

	// Output:
	// WITH users (user_id) AS (SELECT user_id FROM vip_users) UPDATE orders, users SET orders.transport_fee = 0 WHERE users.user_id = orders.user_id
	// WITH users (user_id) AS (SELECT user_id FROM vip_users) UPDATE orders FROM users SET orders.transport_fee = 0 WHERE users.user_id = orders.user_id
}

func ExampleCTEBuilder_delete() {
	sql := With(
		CTETable("users", "user_id").As(
			Select("user_id").From("cheaters"),
		),
	).DeleteFrom("awards").Where(
		"users.user_id = awards.user_id",
	).String()

	fmt.Println(sql)

	// Output:
	// WITH users (user_id) AS (SELECT user_id FROM cheaters) DELETE FROM awards, users WHERE users.user_id = awards.user_id
}

func TestCTEBuilder(t *testing.T) {
	a := assert.New(t)
	cteb := newCTEBuilder()
	ctetb := newCTEQueryBuilder()
	cteb.SQL("/* init */")
	cteb.With(ctetb)
	cteb.SQL("/* after with */")

	ctetb.SQL("/* table init */")
	ctetb.Table("t", "a", "b")
	ctetb.SQL("/* after table */")

	ctetb.As(Select("a", "b").From("t"))
	ctetb.SQL("/* after table as */")

	a.Equal(cteb.TableNames(), []string{ctetb.TableName()})

	sql, args := cteb.Build()
	a.Equal(sql, "/* init */ WITH /* table init */ t (a, b) /* after table */ AS (SELECT a, b FROM t) /* after table as */ /* after with */")
	a.Assert(args == nil)

	sql = ctetb.String()
	a.Equal(sql, "/* table init */ t (a, b) /* after table */ AS (SELECT a, b FROM t) /* after table as */")
}

func TestRecursiveCTEBuilder(t *testing.T) {
	a := assert.New(t)
	cteb := newCTEBuilder()
	cteb.recursive = true
	ctetb := newCTEQueryBuilder()
	cteb.SQL("/* init */")
	cteb.With(ctetb)
	cteb.SQL("/* after with */")

	ctetb.SQL("/* table init */")
	ctetb.Table("t", "a", "b")
	ctetb.SQL("/* after table */")

	ctetb.As(Select("a", "b").From("t"))
	ctetb.SQL("/* after table as */")

	sql, args := cteb.Build()
	a.Equal(sql, "/* init */ WITH RECURSIVE /* table init */ t (a, b) /* after table */ AS (SELECT a, b FROM t) /* after table as */ /* after with */")
	a.Assert(args == nil)

	sql = ctetb.String()
	a.Equal(sql, "/* table init */ t (a, b) /* after table */ AS (SELECT a, b FROM t) /* after table as */")
}

func TestCTEGetFlavor(t *testing.T) {
	a := assert.New(t)
	cteb := newCTEBuilder()

	cteb.SetFlavor(PostgreSQL)
	flavor := cteb.Flavor()
	a.Equal(PostgreSQL, flavor)

	ctebClick := ClickHouse.NewCTEBuilder()
	flavor = ctebClick.Flavor()
	a.Equal(ClickHouse, flavor)
}

func TestCTEQueryBuilderGetFlavor(t *testing.T) {
	a := assert.New(t)
	ctetb := newCTEQueryBuilder()

	ctetb.SetFlavor(PostgreSQL)
	flavor := ctetb.Flavor()
	a.Equal(PostgreSQL, flavor)

	ctetbClick := ClickHouse.NewCTEQueryBuilder()
	flavor = ctetbClick.Flavor()
	a.Equal(ClickHouse, flavor)
}
