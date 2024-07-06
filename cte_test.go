// Copyright 2024 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"

	"github.com/huandu/go-assert"
)

func ExampleWith() {
	sb := With("users", "id", "name").As(
		Select("id", "name").From("users").Where("name IS NOT NULL"),
	).Select("users.id", "orders.id").Join("orders", "users.id = orders.user_id")

	fmt.Println(sb)

	// Output:
	// WITH users (id, name) AS (SELECT id, name FROM users WHERE name IS NOT NULL) SELECT users.id, orders.id FROM users JOIN orders ON users.id = orders.user_id
}

func ExampleCTEBuilder() {
	usersBuilder := Select("id", "name", "level").From("users")
	usersBuilder.Where(
		usersBuilder.GreaterEqualThan("level", 10),
	)
	cteb := With("valid_users").As(usersBuilder)
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
	cteb.SQL("/* init */")
	cteb.With("t", "a", "b")
	cteb.SQL("/* after with */")

	// Make sure that calling Var() will not affect the As().
	cteb.Var(123)

	cteb.As(Select("a", "b").From("t"))
	cteb.SQL("/* after as */")

	sql, args := cteb.Build()
	a.Equal(sql, "/* init */ WITH t (a, b) /* after with */ AS (SELECT a, b FROM t) /* after as */")
	a.Assert(args == nil)
}
