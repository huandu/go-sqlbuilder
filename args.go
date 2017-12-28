// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
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

// Flatten recursively extracts values in slices and returns
// a flattened []interface{} with all values.
// If slices is not a slice, return `[]interface{}{slices}`.
func Flatten(slices interface{}) (flattened []interface{}) {
	v := reflect.ValueOf(slices)
	slices, flattened = flatten(v)

	if slices != nil {
		return []interface{}{slices}
	}

	return flattened
}

func flatten(v reflect.Value) (elem interface{}, flattened []interface{}) {
	k := v.Kind()

	for k == reflect.Interface {
		v = v.Elem()
		k = v.Kind()
	}

	if k != reflect.Slice && k != reflect.Array {
		return v.Interface(), nil
	}

	for i, l := 0, v.Len(); i < l; i++ {
		e, f := flatten(v.Index(i))

		if e == nil {
			flattened = append(flattened, f...)
		} else {
			flattened = append(flattened, e)
		}
	}

	return
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
func (args *Args) Compile(str string) (query string, values []interface{}) {
	buf := &bytes.Buffer{}
	idx := strings.IndexRune(str, '$')
	values = make([]interface{}, 0, len(args.args))
	var namedArgs []sql.NamedArg
	usedNamedArgs := map[string]struct{}{}

	for idx >= 0 && len(str) > 0 {
		if idx > 0 {
			buf.WriteString(str[:idx])
		}

		str = str[idx+1:]

		// Should not happen.
		if len(str) == 0 {
			break
		}

		if str[0] == '$' {
			buf.WriteRune('$')
			str = str[1:]
		} else {
			i := 0

			for ; i < len(str) && '0' <= str[i] && str[i] <= '9'; i++ {
				// Nothing.
			}

			if i > 0 {
				digits := str[:i]
				str = str[i:]

				if pointer, err := strconv.Atoi(digits); err == nil && pointer < len(args.args) {
					arg := args.args[pointer]

					if b, ok := arg.(Builder); ok {
						s, nestedArgs := b.Build()
						buf.WriteString(s)
						values = append(values, nestedArgs...)
					} else if na, ok := arg.(sql.NamedArg); ok {
						buf.WriteRune('@')
						buf.WriteString(na.Name)

						if _, ok := usedNamedArgs[na.Name]; !ok {
							usedNamedArgs[na.Name] = struct{}{}
							namedArgs = append(namedArgs, na)
						}
					} else {
						buf.WriteRune('?')
						values = append(values, arg)
					}
				}
			}
		}

		idx = strings.IndexRune(str, '$')
	}

	if len(str) > 0 {
		buf.WriteString(str)
	}

	if len(namedArgs) > 0 {
		for _, na := range namedArgs {
			values = append(values, na)
		}
	}

	query = buf.String()
	return
}
