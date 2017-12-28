// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
)

// Builder is a general SQL builder.
// It's used by Args to create nested SQL like the `IN` expression in
// `SELECT * FROM t1 WHERE id IN (SELECT id FROM t2)`.
type Builder interface {
	Build() (sql string, args []interface{})
}

type compiledBuilder struct {
	sql  string
	args []interface{}
}

func (cb *compiledBuilder) Build() (sql string, args []interface{}) {
	return cb.sql, cb.args
}

// Buildf creates a Builder from a format string using `fmt.Sprintf`-like syntax.
// As all arguments will be converted to a string internally, e.g. "$0",
// only `%v` and `%s` are valid.
func Buildf(format string, arg ...interface{}) Builder {
	args := &Args{}
	vars := make([]interface{}, 0, len(arg))

	for _, a := range arg {
		vars = append(vars, args.Add(a))
	}

	str := fmt.Sprintf(format, vars...)
	sql, values := args.Compile(str)

	return &compiledBuilder{
		sql:  sql,
		args: values,
	}
}

// Build creates a Builder from a format string.
// The format string uses a special syntax to represent arguments.
// See doc in `Args#Compile` for syntax details.
func Build(format string, arg ...interface{}) Builder {
	args := &Args{}

	for _, a := range arg {
		args.Add(a)
	}

	sql, values := args.Compile(format)
	return &compiledBuilder{
		sql:  sql,
		args: values,
	}
}
