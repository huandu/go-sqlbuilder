// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"strings"
	"testing"

	"github.com/huandu/go-assert"
)

type TestPair struct {
	Expected string
	Actual   string
}

func newTestPair(expected string, fn func(c *Cond) string) *TestPair {
	cond := newCond()
	format := fn(cond)
	sql, _ := cond.Args.CompileWithFlavor(format, PostgreSQL)
	return &TestPair{
		Expected: expected,
		Actual:   sql,
	}
}

func TestCond(t *testing.T) {
	a := assert.New(t)
	cases := []*TestPair{
		newTestPair("$a = $1", func(c *Cond) string { return c.Equal("$a", 123) }),
		newTestPair("$b = $1", func(c *Cond) string { return c.E("$b", 123) }),
		newTestPair("$c = $1", func(c *Cond) string { return c.EQ("$c", 123) }),
		newTestPair("$a <> $1", func(c *Cond) string { return c.NotEqual("$a", 123) }),
		newTestPair("$b <> $1", func(c *Cond) string { return c.NE("$b", 123) }),
		newTestPair("$c <> $1", func(c *Cond) string { return c.NEQ("$c", 123) }),
		newTestPair("$a > $1", func(c *Cond) string { return c.GreaterThan("$a", 123) }),
		newTestPair("$b > $1", func(c *Cond) string { return c.G("$b", 123) }),
		newTestPair("$c > $1", func(c *Cond) string { return c.GT("$c", 123) }),
		newTestPair("$a >= $1", func(c *Cond) string { return c.GreaterEqualThan("$a", 123) }),
		newTestPair("$b >= $1", func(c *Cond) string { return c.GE("$b", 123) }),
		newTestPair("$c >= $1", func(c *Cond) string { return c.GTE("$c", 123) }),
		newTestPair("$a < $1", func(c *Cond) string { return c.LessThan("$a", 123) }),
		newTestPair("$b < $1", func(c *Cond) string { return c.L("$b", 123) }),
		newTestPair("$c < $1", func(c *Cond) string { return c.LT("$c", 123) }),
		newTestPair("$a <= $1", func(c *Cond) string { return c.LessEqualThan("$a", 123) }),
		newTestPair("$b <= $1", func(c *Cond) string { return c.LE("$b", 123) }),
		newTestPair("$c <= $1", func(c *Cond) string { return c.LTE("$c", 123) }),
		newTestPair("$a IN ($1, $2, $3)", func(c *Cond) string { return c.In("$a", 1, 2, 3) }),
		newTestPair("0 = 1", func(c *Cond) string { return c.In("$a") }),
		newTestPair("$a NOT IN ($1, $2, $3)", func(c *Cond) string { return c.NotIn("$a", 1, 2, 3) }),
		newTestPair("$a LIKE $1", func(c *Cond) string { return c.Like("$a", "%Huan%") }),
		newTestPair("$a ILIKE $1", func(c *Cond) string { return c.ILike("$a", "%Huan%") }),
		newTestPair("$a NOT LIKE $1", func(c *Cond) string { return c.NotLike("$a", "%Huan%") }),
		newTestPair("$a NOT ILIKE $1", func(c *Cond) string { return c.NotILike("$a", "%Huan%") }),
		newTestPair("$a IS NULL", func(c *Cond) string { return c.IsNull("$a") }),
		newTestPair("$a IS NOT NULL", func(c *Cond) string { return c.IsNotNull("$a") }),
		newTestPair("$a BETWEEN $1 AND $2", func(c *Cond) string { return c.Between("$a", 123, 456) }),
		newTestPair("$a NOT BETWEEN $1 AND $2", func(c *Cond) string { return c.NotBetween("$a", 123, 456) }),
		newTestPair("NOT 1 = 1", func(c *Cond) string { return c.Not("1 = 1") }),
		newTestPair("EXISTS ($1)", func(c *Cond) string { return c.Exists(1) }),
		newTestPair("NOT EXISTS ($1)", func(c *Cond) string { return c.NotExists(1) }),
		newTestPair("$a > ANY ($1, $2)", func(c *Cond) string { return c.Any("$a", ">", 1, 2) }),
		newTestPair("0 = 1", func(c *Cond) string { return c.Any("$a", ">") }),
		newTestPair("$a < ALL ($1)", func(c *Cond) string { return c.All("$a", "<", 1) }),
		newTestPair("0 = 1", func(c *Cond) string { return c.All("$a", "<") }),
		newTestPair("$a > SOME ($1, $2, $3)", func(c *Cond) string { return c.Some("$a", ">", 1, 2, 3) }),
		newTestPair("0 = 1", func(c *Cond) string { return c.Some("$a", ">") }),
		newTestPair("$a IS DISTINCT FROM $1", func(c *Cond) string { return c.IsDistinctFrom("$a", 1) }),
		newTestPair("$a IS NOT DISTINCT FROM $1", func(c *Cond) string { return c.IsNotDistinctFrom("$a", 1) }),
		newTestPair("$1", func(c *Cond) string { return c.Var(123) }),
		newTestPair("$a @> $1", func(c *Cond) string {	return c.Contains("$a","one","two")}),
	}

	for _, f := range cases {
		a.Equal(f.Actual, f.Expected)
	}
}

func TestOrCond(t *testing.T) {
	a := assert.New(t)
	cases := []*TestPair{
		newTestPair("(1 = 1 OR 2 = 2 OR 3 = 3)", func(c *Cond) string { return c.Or("1 = 1", "2 = 2", "3 = 3") }),

		newTestPair("(1 = 1 OR 2 = 2)", func(c *Cond) string { return c.Or("", "1 = 1", "2 = 2") }),
		newTestPair("(1 = 1 OR 2 = 2)", func(c *Cond) string { return c.Or("1 = 1", "2 = 2", "") }),
		newTestPair("(1 = 1 OR 2 = 2)", func(c *Cond) string { return c.Or("1 = 1", "", "2 = 2") }),

		newTestPair("(1 = 1)", func(c *Cond) string { return c.Or("1 = 1", "", "") }),
		newTestPair("(1 = 1)", func(c *Cond) string { return c.Or("", "1 = 1", "") }),
		newTestPair("(1 = 1)", func(c *Cond) string { return c.Or("", "", "1 = 1") }),
		newTestPair("(1 = 1)", func(c *Cond) string { return c.Or("1 = 1") }),

		{Expected: "", Actual: newCond().Or("")},
		{Expected: "", Actual: newCond().Or()},
		{Expected: "", Actual: newCond().Or("", "", "")},
	}

	for _, f := range cases {
		a.Equal(f.Actual, f.Expected)
	}
}

func TestAndCond(t *testing.T) {
	a := assert.New(t)
	cases := []*TestPair{
		newTestPair("(1 = 1 AND 2 = 2 AND 3 = 3)", func(c *Cond) string { return c.And("1 = 1", "2 = 2", "3 = 3") }),

		newTestPair("(1 = 1 AND 2 = 2)", func(c *Cond) string { return c.And("", "1 = 1", "2 = 2") }),
		newTestPair("(1 = 1 AND 2 = 2)", func(c *Cond) string { return c.And("1 = 1", "2 = 2", "") }),
		newTestPair("(1 = 1 AND 2 = 2)", func(c *Cond) string { return c.And("1 = 1", "", "2 = 2") }),

		newTestPair("(1 = 1)", func(c *Cond) string { return c.And("1 = 1", "", "") }),
		newTestPair("(1 = 1)", func(c *Cond) string { return c.And("", "1 = 1", "") }),
		newTestPair("(1 = 1)", func(c *Cond) string { return c.And("", "", "1 = 1") }),
		newTestPair("(1 = 1)", func(c *Cond) string { return c.And("1 = 1") }),

		{Expected: "", Actual: newCond().And("")},
		{Expected: "", Actual: newCond().And()},
		{Expected: "", Actual: newCond().And("", "", "")},
	}

	for _, f := range cases {
		a.Equal(f.Actual, f.Expected)
	}
}

func TestEmptyCond(t *testing.T) {
	a := assert.New(t)
	cases := []string{
		newCond().Equal("", 123),
		newCond().NotEqual("", 123),
		newCond().GreaterThan("", 123),
		newCond().GreaterEqualThan("", 123),
		newCond().LessThan("", 123),
		newCond().LessEqualThan("", 123),
		newCond().In("", 1, 2, 3),
		newCond().NotIn("", 1, 2, 3),
		newCond().NotIn("a"),
		newCond().Like("", "%Huan%"),
		newCond().ILike("", "%Huan%"),
		newCond().NotLike("", "%Huan%"),
		newCond().NotILike("", "%Huan%"),
		newCond().IsNull(""),
		newCond().IsNotNull(""),
		newCond().Between("", 123, 456),
		newCond().NotBetween("", 123, 456),
		newCond().Not(""),

		newCond().Any("", "", 1, 2),
		newCond().Any("", ">", 1, 2),
		newCond().Any("$a", "", 1, 2),

		newCond().All("", "", 1),
		newCond().All("", ">", 1),
		newCond().All("$a", "", 1),

		newCond().Some("", "", 1, 2, 3),
		newCond().Some("", ">", 1, 2, 3),
		newCond().Some("$a", "", 1, 2, 3),

		newCond().IsDistinctFrom("", 1),
		newCond().IsNotDistinctFrom("", 1),
	}

	expected := ""
	for _, actual := range cases {
		a.Equal(actual, expected)
	}
}

func callCond(fn func(cond *Cond) string) (actual string) {
	cond := &Cond{
		Args: &Args{},
	}
	format := fn(cond)
	actual, _ = cond.Args.CompileWithFlavor(format, PostgreSQL)
	return
}

func TestCondWithFlavor(t *testing.T) {
	a := assert.New(t)
	cond := &Cond{
		Args: &Args{},
	}
	format := strings.Join([]string{
		cond.ILike("f1", 1),
		cond.NotILike("f2", 2),
		cond.IsDistinctFrom("f3", 3),
		cond.IsNotDistinctFrom("f4", 4),
	}, "\n")
	expectedResults := map[Flavor]string{
		PostgreSQL: `f1 ILIKE $1
f2 NOT ILIKE $2
f3 IS DISTINCT FROM $3
f4 IS NOT DISTINCT FROM $4`,
		MySQL: `LOWER(f1) LIKE LOWER(?)
LOWER(f2) NOT LIKE LOWER(?)
NOT f3 <=> ?
f4 <=> ?`,
		SQLite: `f1 ILIKE ?
f2 NOT ILIKE ?
f3 IS DISTINCT FROM ?
f4 IS NOT DISTINCT FROM ?`,
		Presto: `LOWER(f1) LIKE LOWER(?)
LOWER(f2) NOT LIKE LOWER(?)
CASE WHEN f3 IS NULL AND ? IS NULL THEN 0 WHEN f3 IS NOT NULL AND ? IS NOT NULL AND f3 = ? THEN 0 ELSE 1 END = 1
CASE WHEN f4 IS NULL AND ? IS NULL THEN 1 WHEN f4 IS NOT NULL AND ? IS NOT NULL AND f4 = ? THEN 1 ELSE 0 END = 1`,
	}

	for flavor, expected := range expectedResults {
		actual, _ := cond.Args.CompileWithFlavor(format, flavor)
		a.Equal(actual, expected)
	}
}

func TestCondExpr(t *testing.T) {
	a := assert.New(t)
	cond := &Cond{
		Args: &Args{},
	}
	sb1 := Select("1 = 1")
	sb2 := Select("FALSE")
	formats := []string{
		cond.And(),
		cond.Or(),
		cond.And(cond.Var(sb1), cond.Var(sb2)),
		cond.Or(cond.Var(sb1), cond.Var(sb2)),
		cond.Not(cond.Or(cond.Var(sb1), cond.And(cond.Var(sb1), cond.Var(sb2)))),
	}
	expectResults := []string{
		"",
		"",
		"(SELECT 1 = 1 AND SELECT FALSE)",
		"(SELECT 1 = 1 OR SELECT FALSE)",
		"NOT (SELECT 1 = 1 OR (SELECT 1 = 1 AND SELECT FALSE))",
	}

	for i, expected := range expectResults {
		actual, values := cond.Args.Compile(formats[i])
		a.Equal(len(values), 0)
		a.Equal(actual, expected)
	}
}

func TestCondMisuse(t *testing.T) {
	a := assert.New(t)

	cond := NewCond()
	sb := Select("*").
		From("t1").
		Where(cond.Equal("a", 123))
	sql, args := sb.Build()

	a.Equal(sql, "SELECT * FROM t1 WHERE /* INVALID ARG $256 */")
	a.Equal(args, nil)
}

func newCond() *Cond {
	args := &Args{}
	return &Cond{
		Args: args,
	}
}
