// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Escape replaces `$` with `$$` in ident.
func Escape(ident string) string {
	return strings.Replace(ident, "$", "$$", -1)
}

// EscapeAll replaces `$` with `$$` in all strings of ident.
func EscapeAll(ident ...string) []string {
	escaped := make([]string, 0, len(ident))

	for _, i := range ident {
		escaped = append(escaped, Escape(i))
	}

	return escaped
}

// Args stores arguments associated with a SQL.
type Args struct {
	args []interface{}
}

// Add adds an arg to Args and returns a placeholder.
func (args *Args) Add(arg interface{}) string {
	if r, ok := arg.(rawValue); ok {
		return r.expr
	}

	idx := len(args.args)
	args.args = append(args.args, arg)
	return fmt.Sprintf("$%v", idx)
}

type rawValue struct {
	expr string
}

// Raw marks the expr as a raw value which will not be added to args.
func (args *Args) Raw(expr string) interface{} {
	return rawValue{expr}
}

// Compile analyzes builder's sql to standard sql and returns associated args.
//
// A builder uses `$N` to represent an argument in arguments.
// Unescape replaces `$N` to `?`, which is the placeholder supported by SQL driver,
// and then creates a new args associated with the placeholder.
func (args *Args) Compile(sql string) (query string, values []interface{}) {
	buf := &bytes.Buffer{}
	idx := strings.IndexRune(sql, '$')
	values = make([]interface{}, 0, len(args.args))

	for idx >= 0 && len(sql) > 0 {
		if idx > 0 {
			buf.WriteString(sql[:idx])
		}

		sql = sql[idx+1:]

		// Should not happen.
		if len(sql) == 0 {
			break
		}

		if sql[0] == '$' {
			buf.WriteRune('$')
			sql = sql[1:]
		} else {
			i := 0

			for ; i < len(sql) && '0' <= sql[i] && sql[i] <= '9'; i++ {
				// Nothing.
			}

			if i > 0 {
				digits := sql[:i]
				sql = sql[i:]

				if pointer, err := strconv.Atoi(digits); err == nil && pointer < len(args.args) {
					buf.WriteRune('?')
					values = append(values, args.args[pointer])
				}
			}
		}

		idx = strings.IndexRune(sql, '$')
	}

	if len(sql) > 0 {
		buf.WriteString(sql)
	}

	query = buf.String()
	return
}
