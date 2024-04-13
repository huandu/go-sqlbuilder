// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/huandu/go-assert"
)

func TestArgs(t *testing.T) {
	a := assert.New(t)
	start := sql.Named("start", 1234567890)
	end := sql.Named("end", 1234599999)
	named1 := Named("named1", "foo")
	named2 := Named("named2", "bar")

	cases := map[string][]interface{}{
		"abc ? def\n[123]":                   {"abc $? def", 123},
		"abc ? def\n[456]":                   {"abc $0 def", 456},
		"abc  def\n[]":                       {"abc $1 def", 123},
		"abc  def \n[]":                      {"abc ${unknown} def ", 123},
		"abc $ def\n[]":                      {"abc $$ def", 123},
		"abcdef$\n[]":                        {"abcdef$", 123},
		"abc ? ? ? ? def\n[123 456 123 456]": {"abc $? $? $0 $? def", 123, 456, 789},
		"abc ? raw ? raw def\n[123 123]":     {"abc $? $? $0 $? def", 123, Raw("raw"), 789},
		"abc $-1 $a def\n[]":                 {"abc $-1 $a def", 123},

		"abc ? def ? ?\n[foo bar foo]":                                  {"abc ${named1} def ${named2} ${named1}", named2, named1, named2},
		"@end @start @end\n[{{} end 1234599999} {{} start 1234567890}]": {"$? $? $?", end, start, end},
	}

	for expected, c := range cases {
		args := new(Args)

		for i := 1; i < len(c); i++ {
			args.Add(c[i])
		}

		sql, values := args.Compile(c[0].(string))
		actual := fmt.Sprintf("%v\n%v", sql, values)

		a.Equal(actual, expected)
	}

	old := DefaultFlavor
	defer func() {
		DefaultFlavor = old
	}()

	DefaultFlavor = PostgreSQL

	// PostgreSQL flavor compiled sql.
	for expected, c := range cases {
		args := new(Args)

		for i := 1; i < len(c); i++ {
			args.Add(c[i])
		}

		sql, values := args.Compile(c[0].(string))
		actual := fmt.Sprintf("%v\n%v", sql, values)
		expected = toPostgreSQL(expected)

		a.Equal(actual, expected)
	}

	DefaultFlavor = SQLServer

	// SQLServer flavor compiled sql.
	for expected, c := range cases {
		args := new(Args)

		for i := 1; i < len(c); i++ {
			args.Add(c[i])
		}

		sql, values := args.Compile(c[0].(string))
		actual := fmt.Sprintf("%v\n%v", sql, values)
		expected = toSQLServerSQL(expected)

		a.Equal(actual, expected)
	}

	DefaultFlavor = CQL

	for expected, c := range cases {
		args := new(Args)

		for i := 1; i < len(c); i++ {
			args.Add(c[i])
		}

		sql, values := args.Compile(c[0].(string))
		actual := fmt.Sprintf("%v\n%v", sql, values)

		a.Equal(actual, expected)
	}
}

func toPostgreSQL(sql string) string {
	parts := strings.Split(sql, "?")
	buf := &bytes.Buffer{}
	buf.WriteString(parts[0])

	for i, p := range parts[1:] {
		fmt.Fprintf(buf, "$%v", i+1)
		buf.WriteString(p)
	}

	return buf.String()
}

func toSQLServerSQL(sql string) string {
	parts := strings.Split(sql, "?")
	buf := &bytes.Buffer{}
	buf.WriteString(parts[0])

	for i, p := range parts[1:] {
		fmt.Fprintf(buf, "@p%v", i+1)
		buf.WriteString(p)
	}

	return buf.String()
}

func TestArgsAdd(t *testing.T) {
	a := assert.New(t)
	args := &Args{}

	for i := 0; i < maxPredefinedArgs*2; i++ {
		actual := args.Add(i)
		a.Equal(actual, fmt.Sprintf("$%v", i))
	}
}
