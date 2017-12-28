// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Args stores arguments associated with a SQL.
type Args struct {
	args         []interface{}
	namedArgs    map[string]int
	sqlNamedArgs map[string]int
}

// Add adds an arg to Args and returns a placeholder.
func (args *Args) Add(arg interface{}) string {
	return fmt.Sprintf("$%v", args.add(arg))
}

func (args *Args) add(arg interface{}) int {
	idx := len(args.args)

	switch a := arg.(type) {
	case sql.NamedArg:
		if args.sqlNamedArgs == nil {
			args.sqlNamedArgs = map[string]int{}
		}

		if p, ok := args.sqlNamedArgs[a.Name]; ok {
			arg = args.args[p]
			break
		}

		args.sqlNamedArgs[a.Name] = idx
	case namedArgs:
		if args.namedArgs == nil {
			args.namedArgs = map[string]int{}
		}

		if p, ok := args.namedArgs[a.name]; ok {
			arg = args.args[p]
			break
		}

		// Find out the real arg and add it to args.
		idx = args.add(a.arg)
		args.namedArgs[a.name] = idx
		return idx
	}

	args.args = append(args.args, arg)
	return idx
}

// Compile analyzes builder's format to standard sql and returns associated args.
//
// The format string uses a special syntax to represent arguments.
//
//     $? uses successive arguments passed in the call. It works similar as `%v` in `fmt.Sprintf`.
//     $0 $1 ... $n uses nth-argument passed in the call. Next $? will use arguments n+1.
//     ${name} uses a named argument created by `Named` with `name`.
//     $$ represents a "$" string.
func (args *Args) Compile(format string) (query string, values []interface{}) {
	buf := &bytes.Buffer{}
	idx := strings.IndexRune(format, '$')
	offset := 0

	for idx >= 0 && len(format) > 0 {
		if idx > 0 {
			buf.WriteString(format[:idx])
		}

		format = format[idx+1:]

		// Should not happen.
		if len(format) == 0 {
			break
		}

		if format[0] == '$' {
			buf.WriteRune('$')
			format = format[1:]
		} else if format[0] == '{' {
			format, values = args.compileNamed(buf, format, values)
		} else if '0' <= format[0] && format[0] <= '9' {
			format, values, offset = args.compileDigits(buf, format, values, offset)
		} else if format[0] == '?' {
			format, values, offset = args.compileSuccessive(buf, format[1:], values, offset)
		}

		idx = strings.IndexRune(format, '$')
	}

	if len(format) > 0 {
		buf.WriteString(format)
	}

	query = buf.String()

	if len(args.sqlNamedArgs) > 0 {
		// Stabilize the sequence to make it easier to write test cases.
		ints := make([]int, 0, len(args.sqlNamedArgs))

		for _, p := range args.sqlNamedArgs {
			ints = append(ints, p)
		}

		sort.Ints(ints)

		for _, i := range ints {
			values = append(values, args.args[i])
		}
	}

	return
}

func (args *Args) compileNamed(buf *bytes.Buffer, format string, values []interface{}) (string, []interface{}) {
	i := 1

	for ; i < len(format) && format[i] != '}'; i++ {
		// Nothing.
	}

	// Invalid $ format. Ignore it.
	if i == len(format) {
		return format, values
	}

	name := format[1:i]
	format = format[i+1:]

	if p, ok := args.namedArgs[name]; ok {
		format, values, _ = args.compileSuccessive(buf, format, values, p)
	}

	return format, values
}

func (args *Args) compileDigits(buf *bytes.Buffer, format string, values []interface{}, offset int) (string, []interface{}, int) {
	i := 1

	for ; i < len(format) && '0' <= format[i] && format[i] <= '9'; i++ {
		// Nothing.
	}

	digits := format[:i]
	format = format[i:]

	if pointer, err := strconv.Atoi(digits); err == nil {
		return args.compileSuccessive(buf, format, values, pointer)
	}

	return format, values, offset
}

func (args *Args) compileSuccessive(buf *bytes.Buffer, format string, values []interface{}, offset int) (string, []interface{}, int) {
	if offset >= len(args.args) {
		return format, values, offset
	}

	arg := args.args[offset]

	switch a := arg.(type) {
	case Builder:
		s, nestedArgs := a.Build()
		buf.WriteString(s)
		values = append(values, nestedArgs...)
	case sql.NamedArg:
		buf.WriteRune('@')
		buf.WriteString(a.Name)
	case rawArgs:
		buf.WriteString(a.expr)
	case listArgs:
		if len(a.args) > 0 {
			buf.WriteRune('?')
		}

		for j := 1; j < len(a.args); j++ {
			buf.WriteString(", ?")
		}

		values = append(values, a.args...)
	default:
		buf.WriteRune('?')
		values = append(values, arg)
	}

	return format, values, offset + 1
}
