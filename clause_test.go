// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"
)

func TestComposite(t *testing.T) {
	cases := map[string]func() string{
		"$$b = $0":                  func() string { return Interpret(NewEqualOperation("$b").NewClause(123), newSelectBuilder()) },
		"$$b != $0":                 func() string { return Interpret(NewNotEqualOperation("$b").NewClause(123), newSelectBuilder()) },
		"$$b > $0":                  func() string { return Interpret(NewGreaterThanOperation("$b").NewClause(123), newSelectBuilder()) },
		"$$b >= $0":                 func() string { return Interpret(NewGreaterEqualThanOperation("$b").NewClause(123), newSelectBuilder()) },
		"$$b < $0":                  func() string { return Interpret(NewLessThanOperation("$b").NewClause(123), newSelectBuilder()) },
		"$$b <= $0":                 func() string { return Interpret(NewLessEqualThanOperation("$b").NewClause(123), newSelectBuilder()) },
		"$$a IN ($0, $1, $2)":       func() string { return Interpret(NewInOperation("$a").NewClause(1, 2, 3), newSelectBuilder()) },
		"$$a NOT IN ($0, $1, $2)":   func() string { return Interpret(NewNotInOperation("$a").NewClause(1, 2, 3), newSelectBuilder()) },
		"$$a LIKE $0":               func() string { return Interpret(NewLikeOperation("$a").NewClause("%Huan%"), newSelectBuilder()) },
		"$$a NOT LIKE $0":           func() string { return Interpret(NewNotLikeOperation("$a").NewClause("%Huan%"), newSelectBuilder()) },
		"$$a IS NULL":               func() string { return Interpret(NewIsNullOperation("$a").NewClause(), newSelectBuilder()) },
		"$$a IS NOT NULL":           func() string { return Interpret(NewNotNullOperation("$a").NewClause(), newSelectBuilder()) },
		"$$a BETWEEN $0 AND $1":     func() string { return Interpret(NewBetweenOperation("$a").NewClause(123, 456), newSelectBuilder()) },
		"$$a NOT BETWEEN $0 AND $1": func() string { return Interpret(NewNotBetweenOperation("$a").NewClause(123, 456), newSelectBuilder()) },
		"(b = $0 OR a = $1 OR c = $2 OR (NOT (d = $3 AND e = $4 AND f = $5)) OR (NOT g = $6))": func() string {
			c := NewEqualOperation("b").NewClause(123).Or(
				NewEqualOperation("a").NewClause(456),
				NewEqualOperation("c").NewClause(789),
				NewEqualOperation("d").NewClause(1).And(
					NewEqualOperation("e").NewClause(2),
					NewEqualOperation("f").NewClause(3),
				).Not(),
				NewEqualOperation("g").NewClause(4).Not(),
			)
			return Interpret(
				c, NewSelectBuilder())
		},
	}

	for expected, f := range cases {
		if actual := f(); expected != actual {
			t.Fatalf("invalid result. [expected:%v] [actual:%v]", expected, actual)
		}
	}
}

func ExampleComposite() {
	c := fooEOperation.NewClause(1).And(barGEOperation.NewClause(2))
	sql, args := query(c)

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT * FROM table WHERE (foo = ? AND bar >= ?)
	// [1 2]
}

var (
	fooEOperation  = NewEqualOperation("foo")
	barGEOperation = NewGreaterEqualThanOperation("bar")
)

func query(clause Clause) (string, []interface{}) {
	sb := NewSelectBuilder()
	sb.Select("*").From("table").Where(
		Interpret(clause, sb),
	)
	sql, args := sb.Build()
	return sql, args
}
