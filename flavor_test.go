// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"

	"github.com/huandu/go-assert"
)

func TestFlavor(t *testing.T) {
	a := assert.New(t)
	cases := map[Flavor]string{
		0:          "<invalid>",
		MySQL:      "MySQL",
		PostgreSQL: "PostgreSQL",
		SQLite:     "SQLite",
		SQLServer:  "SQLServer",
	}

	for f, expected := range cases {
		actual := f.String()
		a.Equal(actual, expected)
	}
}

func ExampleFlavor() {
	// Create a flavored builder.
	sb := PostgreSQL.NewSelectBuilder()
	sb.Select("name").From("user").Where(
		sb.E("id", 1234),
		sb.G("rank", 3),
	)
	sql, args := sb.Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT name FROM user WHERE id = $1 AND rank > $2
	// [1234 3]
}

func ExampleFlavor_Interpolate_mySQL() {
	sb := MySQL.NewSelectBuilder()
	sb.Select("name").From("user").Where(
		sb.NE("id", 1234),
		sb.E("name", "Charmy Liu"),
		sb.Like("desc", "%mother's day%"),
	)
	sql, args := sb.Build()
	query, err := MySQL.Interpolate(sql, args)

	fmt.Println(query)
	fmt.Println(err)

	// Output:
	// SELECT name FROM user WHERE id <> 1234 AND name = 'Charmy Liu' AND desc LIKE '%mother\'s day%'
	// <nil>
}

func ExampleFlavor_Interpolate_postgreSQL() {
	// Only the last `$1` is interpolated.
	// Others are not interpolated as they are inside dollar quote (the `$$`).
	query, err := PostgreSQL.Interpolate(`
CREATE FUNCTION dup(in int, out f1 int, out f2 text) AS $$
    SELECT $1, CAST($1 AS text) || ' is text'
$$
LANGUAGE SQL;

SELECT * FROM dup($1);`, []interface{}{42})

	fmt.Println(query)
	fmt.Println(err)

	// Output:
	//
	// CREATE FUNCTION dup(in int, out f1 int, out f2 text) AS $$
	//     SELECT $1, CAST($1 AS text) || ' is text'
	// $$
	// LANGUAGE SQL;
	//
	// SELECT * FROM dup(42);
	// <nil>
}

func ExampleFlavor_Interpolate_sqlite() {
	sb := SQLite.NewSelectBuilder()
	sb.Select("name").From("user").Where(
		sb.NE("id", 1234),
		sb.E("name", "Charmy Liu"),
		sb.Like("desc", "%mother's day%"),
	)
	sql, args := sb.Build()
	query, err := SQLite.Interpolate(sql, args)

	fmt.Println(query)
	fmt.Println(err)

	// Output:
	// SELECT name FROM user WHERE id <> 1234 AND name = 'Charmy Liu' AND desc LIKE '%mother\'s day%'
	// <nil>
}

func ExampleFlavor_Interpolate_sqlServer() {
	sb := SQLServer.NewSelectBuilder()
	sb.Select("name").From("user").Where(
		sb.NE("id", 1234),
		sb.E("name", "Charmy Liu"),
		sb.Like("desc", "%mother's day%"),
	)
	sql, args := sb.Build()
	query, err := SQLServer.Interpolate(sql, args)

	fmt.Println(query)
	fmt.Println(err)

	// Output:
	// SELECT name FROM user WHERE id <> 1234 AND name = N'Charmy Liu' AND desc LIKE N'%mother\'s day%'
	// <nil>
}

func ExampleFlavor_Interpolate_cql() {
	sb := CQL.NewSelectBuilder()
	sb.Select("name").From("user").Where(
		sb.E("id", 1234),
		sb.E("name", "Charmy Liu"),
	)
	sql, args := sb.Build()
	query, err := CQL.Interpolate(sql, args)

	fmt.Println(query)
	fmt.Println(err)

	// Output:
	// SELECT name FROM user WHERE id = 1234 AND name = 'Charmy Liu'
	// <nil>
}

func ExampleFlavor_Interpolate_oracle() {
	sb := Oracle.NewSelectBuilder()
	sb.Select("name").From("user").Where(
		sb.E("id", 1234),
		sb.E("name", "Charmy Liu"),
		sb.E("enabled", true),
	)
	sql, args := sb.Build()
	query, err := Oracle.Interpolate(sql, args)

	fmt.Println(query)
	fmt.Println(err)

	// Output:
	// SELECT name FROM user WHERE id = 1234 AND name = 'Charmy Liu' AND enabled = 1
	// <nil>
}
