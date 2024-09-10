// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"github.com/huandu/go-assert"
)

func TestCond(t *testing.T) {
	a := assert.New(t)
	cases := map[string]func() string{
		"$$a = $0":                    func() string { return newTestCond().Equal("$a", 123) },
		"$$b = $0":                    func() string { return newTestCond().E("$b", 123) },
		"$$c = $0":                    func() string { return newTestCond().EQ("$c", 123) },
		"$$a <> $0":                   func() string { return newTestCond().NotEqual("$a", 123) },
		"$$b <> $0":                   func() string { return newTestCond().NE("$b", 123) },
		"$$c <> $0":                   func() string { return newTestCond().NEQ("$c", 123) },
		"$$a > $0":                    func() string { return newTestCond().GreaterThan("$a", 123) },
		"$$b > $0":                    func() string { return newTestCond().G("$b", 123) },
		"$$c > $0":                    func() string { return newTestCond().GT("$c", 123) },
		"$$a >= $0":                   func() string { return newTestCond().GreaterEqualThan("$a", 123) },
		"$$b >= $0":                   func() string { return newTestCond().GE("$b", 123) },
		"$$c >= $0":                   func() string { return newTestCond().GTE("$c", 123) },
		"$$a < $0":                    func() string { return newTestCond().LessThan("$a", 123) },
		"$$b < $0":                    func() string { return newTestCond().L("$b", 123) },
		"$$c < $0":                    func() string { return newTestCond().LT("$c", 123) },
		"$$a <= $0":                   func() string { return newTestCond().LessEqualThan("$a", 123) },
		"$$b <= $0":                   func() string { return newTestCond().LE("$b", 123) },
		"$$c <= $0":                   func() string { return newTestCond().LTE("$c", 123) },
		"$$a IN ($0, $1, $2)":         func() string { return newTestCond().In("$a", 1, 2, 3) },
		"$$a NOT IN ($0, $1, $2)":     func() string { return newTestCond().NotIn("$a", 1, 2, 3) },
		"$$a LIKE $0":                 func() string { return newTestCond().Like("$a", "%Huan%") },
		"$$a ILIKE $0":                func() string { return newTestCond().ILike("$a", "%Huan%") },
		"$$a NOT LIKE $0":             func() string { return newTestCond().NotLike("$a", "%Huan%") },
		"$$a NOT ILIKE $0":            func() string { return newTestCond().NotILike("$a", "%Huan%") },
		"$$a IS NULL":                 func() string { return newTestCond().IsNull("$a") },
		"$$a IS NOT NULL":             func() string { return newTestCond().IsNotNull("$a") },
		"$$a BETWEEN $0 AND $1":       func() string { return newTestCond().Between("$a", 123, 456) },
		"$$a NOT BETWEEN $0 AND $1":   func() string { return newTestCond().NotBetween("$a", 123, 456) },
		"(1 = 1 OR 2 = 2 OR 3 = 3)":   func() string { return newTestCond().Or("1 = 1", "2 = 2", "3 = 3") },
		"(1 = 1 AND 2 = 2 AND 3 = 3)": func() string { return newTestCond().And("1 = 1", "2 = 2", "3 = 3") },
		"NOT 1 = 1":                   func() string { return newTestCond().Not("1 = 1") },
		"EXISTS ($0)":                 func() string { return newTestCond().Exists(1) },
		"NOT EXISTS ($0)":             func() string { return newTestCond().NotExists(1) },
		"$$a > ANY ($0, $1)":          func() string { return newTestCond().Any("$a", ">", 1, 2) },
		"$$a < ALL ($0)":              func() string { return newTestCond().All("$a", "<", 1) },
		"$$a > SOME ($0, $1, $2)":     func() string { return newTestCond().Some("$a", ">", 1, 2, 3) },
		"$0":                          func() string { return newTestCond().Var(123) },
	}

	for expected, f := range cases {
		actual := f()
		a.Equal(actual, expected)
	}
}

func newTestCond() *Cond {
	return &Cond{
		Args: &Args{},
	}
}
