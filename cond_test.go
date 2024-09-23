// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"strings"
	"testing"

	"github.com/huandu/go-assert"
)

func TestCond(t *testing.T) {
	a := assert.New(t)
	cases := map[string]func(cond *Cond) string{
		"$$a = $1":                    func(cond *Cond) string { return cond.Equal("$a", 123) },
		"$$b = $1":                    func(cond *Cond) string { return cond.E("$b", 123) },
		"$$c = $1":                    func(cond *Cond) string { return cond.EQ("$c", 123) },
		"$$a <> $1":                   func(cond *Cond) string { return cond.NotEqual("$a", 123) },
		"$$b <> $1":                   func(cond *Cond) string { return cond.NE("$b", 123) },
		"$$c <> $1":                   func(cond *Cond) string { return cond.NEQ("$c", 123) },
		"$$a > $1":                    func(cond *Cond) string { return cond.GreaterThan("$a", 123) },
		"$$b > $1":                    func(cond *Cond) string { return cond.G("$b", 123) },
		"$$c > $1":                    func(cond *Cond) string { return cond.GT("$c", 123) },
		"$$a >= $1":                   func(cond *Cond) string { return cond.GreaterEqualThan("$a", 123) },
		"$$b >= $1":                   func(cond *Cond) string { return cond.GE("$b", 123) },
		"$$c >= $1":                   func(cond *Cond) string { return cond.GTE("$c", 123) },
		"$$a < $1":                    func(cond *Cond) string { return cond.LessThan("$a", 123) },
		"$$b < $1":                    func(cond *Cond) string { return cond.L("$b", 123) },
		"$$c < $1":                    func(cond *Cond) string { return cond.LT("$c", 123) },
		"$$a <= $1":                   func(cond *Cond) string { return cond.LessEqualThan("$a", 123) },
		"$$b <= $1":                   func(cond *Cond) string { return cond.LE("$b", 123) },
		"$$c <= $1":                   func(cond *Cond) string { return cond.LTE("$c", 123) },
		"$$a IN ($1, $2, $3)":         func(cond *Cond) string { return cond.In("$a", 1, 2, 3) },
		"$$a NOT IN ($1, $2, $3)":     func(cond *Cond) string { return cond.NotIn("$a", 1, 2, 3) },
		"$$a LIKE $1":                 func(cond *Cond) string { return cond.Like("$a", "%Huan%") },
		"$$a ILIKE $1":                func(cond *Cond) string { return cond.ILike("$a", "%Huan%") },
		"$$a NOT LIKE $1":             func(cond *Cond) string { return cond.NotLike("$a", "%Huan%") },
		"$$a NOT ILIKE $1":            func(cond *Cond) string { return cond.NotILike("$a", "%Huan%") },
		"$$a IS NULL":                 func(cond *Cond) string { return cond.IsNull("$a") },
		"$$a IS NOT NULL":             func(cond *Cond) string { return cond.IsNotNull("$a") },
		"$$a BETWEEN $1 AND $2":       func(cond *Cond) string { return cond.Between("$a", 123, 456) },
		"$$a NOT BETWEEN $1 AND $2":   func(cond *Cond) string { return cond.NotBetween("$a", 123, 456) },
		"(1 = 1 OR 2 = 2 OR 3 = 3)":   func(cond *Cond) string { return cond.Or("1 = 1", "2 = 2", "3 = 3") },
		"(1 = 1 AND 2 = 2 AND 3 = 3)": func(cond *Cond) string { return cond.And("1 = 1", "2 = 2", "3 = 3") },
		"NOT 1 = 1":                   func(cond *Cond) string { return cond.Not("1 = 1") },
		"EXISTS ($1)":                 func(cond *Cond) string { return cond.Exists(1) },
		"NOT EXISTS ($1)":             func(cond *Cond) string { return cond.NotExists(1) },
		"$$a > ANY ($1, $2)":          func(cond *Cond) string { return cond.Any("$a", ">", 1, 2) },
		"$$a < ALL ($1)":              func(cond *Cond) string { return cond.All("$a", "<", 1) },
		"$$a > SOME ($1, $2, $3)":     func(cond *Cond) string { return cond.Some("$a", ">", 1, 2, 3) },
		"$$a IS DISTINCT FROM $1":     func(cond *Cond) string { return cond.IsDistinctFrom("$a", 1) },
		"$$a IS NOT DISTINCT FROM $1": func(cond *Cond) string { return cond.IsNotDistinctFrom("$a", 1) },
		"$1":                          func(cond *Cond) string { return cond.Var(123) },
	}

	for expected, f := range cases {
		actual := callCond(f)
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
