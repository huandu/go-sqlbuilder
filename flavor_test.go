// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestFlavor(t *testing.T) {
	cases := map[Flavor]string{
		0:          "<invalid>",
		MySQL:      "MySQL",
		PostgreSQL: "PostgreSQL",
	}

	for f, expected := range cases {
		if actual := f.String(); actual != expected {
			t.Fatalf("invalid flavor name. [expected:%v] [actual:%v]", expected, actual)
		}
	}
}

func TestFlavorInterpolate(t *testing.T) {
	dt := time.Date(2019, 4, 24, 12, 23, 34, 123456789, time.FixedZone("CST", 8*60*60)) // 2019-04-24 12:23:34.987654321 CST
	_, errOutOfRange := strconv.ParseInt("12345678901234567890", 10, 32)
	cases := []struct {
		flavor Flavor
		sql    string
		args   []interface{}
		query  string
		err    error
	}{
		{
			MySQL,
			"SELECT * FROM a WHERE name = ? AND state IN (?, ?, ?, ?, ?)", []interface{}{"I'm fine", 42, int8(8), int16(-16), int32(32), int64(64)},
			"SELECT * FROM a WHERE name = 'I\\'m fine' AND state IN (42, 8, -16, 32, 64)", nil,
		},
		{
			MySQL,
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN (?, '?', ?, ?, ?, ?, ?)", []interface{}{"\r\n\b\t\x1a\x00\\\"'", uint(42), uint8(8), uint16(16), uint32(32), uint64(64), "useless"},
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN ('\\r\\n\\b\\t\\Z\\0\\\\\\\"\\'', '?', 42, 8, 16, 32, 64)", nil,
		},
		{
			MySQL,
			"SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?", []interface{}{true, false, float32(1.234567), float64(9.87654321), []byte(nil), []byte("I'm bytes"), dt, time.Time{}, nil},
			"SELECT TRUE, FALSE, 1.234567, 9.87654321, NULL, _binary'I\\'m bytes', '2019-04-24 12:23:34.123457', '0000-00-00', NULL", nil,
		},
		{
			MySQL,
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\?", []interface{}{MySQL},
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\'MySQL'", nil,
		},
		{
			MySQL,
			"SELECT ?", nil,
			"", ErrInterpolateMissingArgs,
		},
		{
			MySQL,
			"SELECT ?", []interface{}{complex(1, 2)},
			"", ErrInterpolateUnsupportedArgs,
		},

		{
			PostgreSQL,
			"SELECT * FROM a WHERE name = $3 AND state IN ($2, $4, $1, $6, $5)", []interface{}{"I'm fine", 42, int8(8), int16(-16), int32(32), int64(64)},
			"SELECT * FROM a WHERE name = 8 AND state IN (42, -16, E'I\\'m fine', 64, 32)", nil,
		},
		{
			PostgreSQL,
			"SELECT * FROM $abc$$1$abc$1$1 WHERE name = \"$1\" AND state IN ($2, '$1', $3, $6, $5, $4, $2) $3", []interface{}{"\r\n\b\t\x1a\x00\\\"'", uint(42), uint8(8), uint16(16), uint32(32), uint64(64), "useless"},
			"SELECT * FROM $abc$$1$abc$1E'\\r\\n\\b\\t\\Z\\0\\\\\\\"\\'' WHERE name = \"$1\" AND state IN (42, '$1', 8, 64, 32, 16, 42) 8", nil,
		},
		{
			PostgreSQL,
			"SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $11, $a", []interface{}{true, false, float32(1.234567), float64(9.87654321), []byte(nil), []byte("I'm bytes"), dt, time.Time{}, nil, 10, 11, 12},
			"SELECT TRUE, FALSE, 1.234567, 9.87654321, NULL, E'\\\\x49276D206279746573'::bytea, '2019-04-24 12:23:34.123457 CST', '0000-00-00', NULL, 11, $a", nil,
		},
		{
			PostgreSQL,
			"SELECT '\\'$1', \"\\\"$1\", `$1`, \\$1a, $$1$$, $a $b$ $a $ $1$b$1$1 $a$ $", []interface{}{MySQL},
			"SELECT '\\'$1', \"\\\"$1\", `E'MySQL'`, \\E'MySQL'a, $$1$$, $a $b$ $a $ $1$b$1E'MySQL' $a$ $", nil,
		},
		{
			PostgreSQL,
			"SELECT * FROM a WHERE name = 'Huan''Du''$1' AND desc = $1", []interface{}{"c'mon"},
			"SELECT * FROM a WHERE name = 'Huan''Du''$1' AND desc = E'c\\'mon'", nil,
		},
		{
			PostgreSQL,
			"SELECT $1", nil,
			"", ErrInterpolateMissingArgs,
		},
		{
			PostgreSQL,
			"SELECT $1", []interface{}{complex(1, 2)},
			"", ErrInterpolateUnsupportedArgs,
		},
		{
			PostgreSQL,
			"SELECT $12345678901234567890", nil,
			"", errOutOfRange,
		},
	}

	for idx, c := range cases {
		query, err := c.flavor.Interpolate(c.sql, c.args)

		if query != c.query || (err != c.err && err.Error() != c.err.Error()) {
			t.Fatalf("unexpected interpolate result. [idx:%v] [err:%v] [case:%#v]\n  expected: %v\n  actual:   %v",
				idx, err, c, c.query, query)
		}
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

func ExampleFlavor_Interpolate() {
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
