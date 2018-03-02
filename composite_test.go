// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"
)

func TestComposite(t *testing.T) {
	cases := map[string]func() string{
		"$$b = $0":                  func() string { return Interpret(NewEqualClause("$b").SetOperand(123), newSelectBuilder()) },
		"$$b != $0":                 func() string { return Interpret(NewNotEqualClause("$b").SetOperand(123), newSelectBuilder()) },
		"$$b > $0":                  func() string { return Interpret(NewGreaterThanClause("$b").SetOperand(123), newSelectBuilder()) },
		"$$b >= $0":                 func() string { return Interpret(NewGreaterEqualThanClause("$b").SetOperand(123), newSelectBuilder()) },
		"$$b < $0":                  func() string { return Interpret(NewLessThanClause("$b").SetOperand(123), newSelectBuilder()) },
		"$$b <= $0":                 func() string { return Interpret(NewLessEqualThanClause("$b").SetOperand(123), newSelectBuilder()) },
		"$$a IN ($0, $1, $2)":       func() string { return Interpret(NewInClause("$a").SetOperand(1, 2, 3), newSelectBuilder()) },
		"$$a NOT IN ($0, $1, $2)":   func() string { return Interpret(NewNotInClause("$a").SetOperand(1, 2, 3), newSelectBuilder()) },
		"$$a LIKE $0":               func() string { return Interpret(NewLikeClause("$a").SetOperand("%Huan%"), newSelectBuilder()) },
		"$$a NOT LIKE $0":           func() string { return Interpret(NewNotLikeClause("$a").SetOperand("%Huan%"), newSelectBuilder()) },
		"$$a IS NULL":               func() string { return Interpret(NewIsNullClause("$a").SetOperand(), newSelectBuilder()) },
		"$$a IS NOT NULL":           func() string { return Interpret(NewNotNullClause("$a").SetOperand(), newSelectBuilder()) },
		"$$a BETWEEN $0 AND $1":     func() string { return Interpret(NewBetweenClause("$a").SetOperand(123, 456), newSelectBuilder()) },
		"$$a NOT BETWEEN $0 AND $1": func() string { return Interpret(NewNotBetweenClause("$a").SetOperand(123, 456), newSelectBuilder()) },
		"(b = $0 OR a = $1 OR c = $2 OR (NOT (d = $3 AND e = $4 AND f = $5)) OR (NOT g = $6))": func() string {
			c := NewEqualClause("b").SetOperand(123).Or(
				NewEqualClause("a").SetOperand(456),
				NewEqualClause("c").SetOperand(789),
				NewEqualClause("d").SetOperand(1).And(
					NewEqualClause("e").SetOperand(2),
					NewEqualClause("f").SetOperand(3),
				).Not(),
				NewEqualClause("g").SetOperand(4).Not(),
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
	c := fooEClause.SetOperand(1).And(barGEClause.SetOperand(2))
	sql, args := query(c)

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT * FROM table WHERE (foo = ? AND bar >= ?)
	// [1 2]
}

var (
	fooEClause = NewEqualClause("foo")
	barGEClause = NewGreaterEqualThanClause("bar")
)

func query(clause Clause) (string, []interface{}) {
	sb := NewSelectBuilder()
	sb.Select("*").From("table").Where(
		Interpret(clause, sb),
	)
	sql, args := sb.Build()
	return sql, args
}
