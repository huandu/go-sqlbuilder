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

type freestyleBuilder struct {
	sql  string
	args *Args
}

func (fb *freestyleBuilder) Build() (sql string, args []interface{}) {
	return fb.args.Compile(fb.sql)
}

// Buildf creates a Builder from a fmt string.
func Buildf(format string, arg ...interface{}) Builder {
	args := &Args{}
	vars := make([]interface{}, 0, len(arg))

	for _, a := range arg {
		vars = append(vars, args.Add(a))
	}

	return &freestyleBuilder{
		sql:  fmt.Sprintf(format, vars...),
		args: args,
	}
}
