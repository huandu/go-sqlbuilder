// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"strings"
)

// Cond provides several helper methods to build conditions.
type Cond struct {
	Args *Args
}

// Equal represents "field = value".
func (c *Cond) Equal(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" = ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}

// E is an alias of Equal.
func (c *Cond) E(field string, value interface{}) string {
	return c.Equal(field, value)
}

// NotEqual represents "field <> value".
func (c *Cond) NotEqual(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" <> ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}

// NE is an alias of NotEqual.
func (c *Cond) NE(field string, value interface{}) string {
	return c.NotEqual(field, value)
}

// GreaterThan represents "field > value".
func (c *Cond) GreaterThan(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" > ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}

// G is an alias of GreaterThan.
func (c *Cond) G(field string, value interface{}) string {
	return c.GreaterThan(field, value)
}

// GreaterEqualThan represents "field >= value".
func (c *Cond) GreaterEqualThan(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" >= ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}

// GE is an alias of GreaterEqualThan.
func (c *Cond) GE(field string, value interface{}) string {
	return c.GreaterEqualThan(field, value)
}

// LessThan represents "field < value".
func (c *Cond) LessThan(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" < ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}

// L is an alias of LessThan.
func (c *Cond) L(field string, value interface{}) string {
	return c.LessThan(field, value)
}

// LessEqualThan represents "field <= value".
func (c *Cond) LessEqualThan(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" <= ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}

// LE is an alias of LessEqualThan.
func (c *Cond) LE(field string, value interface{}) string {
	return c.LessEqualThan(field, value)
}

// In represents "field IN (value...)".
func (c *Cond) In(field string, value ...interface{}) string {
	vs := make([]string, 0, len(value))

	for _, v := range value {
		vs = append(vs, c.Args.Add(v))
	}

	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" IN (")
	buf.WriteString(strings.Join(vs, ", "))
	buf.WriteString(")")
	return buf.String()
}

// NotIn represents "field NOT IN (value...)".
func (c *Cond) NotIn(field string, value ...interface{}) string {
	vs := make([]string, 0, len(value))

	for _, v := range value {
		vs = append(vs, c.Args.Add(v))
	}

	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" NOT IN (")
	buf.WriteString(strings.Join(vs, ", "))
	buf.WriteString(")")
	return buf.String()
}

// Like represents "field LIKE value".
func (c *Cond) Like(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" LIKE ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}

// NotLike represents "field NOT LIKE value".
func (c *Cond) NotLike(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" NOT LIKE ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}

// IsNull represents "field IS NULL".
func (c *Cond) IsNull(field string) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" IS NULL")
	return buf.String()
}

// IsNotNull represents "field IS NOT NULL".
func (c *Cond) IsNotNull(field string) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" IS NOT NULL")
	return buf.String()
}

// Between represents "field BETWEEN lower AND upper".
func (c *Cond) Between(field string, lower, upper interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" BETWEEN ")
	buf.WriteString(c.Args.Add(lower))
	buf.WriteString(" AND ")
	buf.WriteString(c.Args.Add(upper))
	return buf.String()
}

// NotBetween represents "field NOT BETWEEN lower AND upper".
func (c *Cond) NotBetween(field string, lower, upper interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" NOT BETWEEN ")
	buf.WriteString(c.Args.Add(lower))
	buf.WriteString(" AND ")
	buf.WriteString(c.Args.Add(upper))
	return buf.String()
}

// Or represents OR logic like "expr1 OR expr2 OR expr3".
func (c *Cond) Or(orExpr ...string) string {
	buf := newStringBuilder()
	buf.WriteString("(")
	buf.WriteString(strings.Join(orExpr, " OR "))
	buf.WriteString(")")
	return buf.String()
}

// And represents AND logic like "expr1 AND expr2 AND expr3".
func (c *Cond) And(andExpr ...string) string {
	buf := newStringBuilder()
	buf.WriteString("(")
	buf.WriteString(strings.Join(andExpr, " AND "))
	buf.WriteString(")")
	return buf.String()
}

// Exists represents "EXISTS (subquery)".
func (c *Cond) Exists(subquery interface{}) string {
	buf := newStringBuilder()
	buf.WriteString("EXISTS (")
	buf.WriteString(c.Args.Add(subquery))
	buf.WriteString(")")
	return buf.String()
}

// NotExists represents "NOT EXISTS (subquery)".
func (c *Cond) NotExists(subquery interface{}) string {
	buf := newStringBuilder()
	buf.WriteString("NOT EXISTS (")
	buf.WriteString(c.Args.Add(subquery))
	buf.WriteString(")")
	return buf.String()
}

// Any represents "field op ANY (value...)".
func (c *Cond) Any(field, op string, value ...interface{}) string {
	vs := make([]string, 0, len(value))

	for _, v := range value {
		vs = append(vs, c.Args.Add(v))
	}

	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" ")
	buf.WriteString(op)
	buf.WriteString(" ANY (")
	buf.WriteString(strings.Join(vs, ", "))
	buf.WriteString(")")
	return buf.String()
}

// All represents "field op ALL (value...)".
func (c *Cond) All(field, op string, value ...interface{}) string {
	vs := make([]string, 0, len(value))

	for _, v := range value {
		vs = append(vs, c.Args.Add(v))
	}

	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" ")
	buf.WriteString(op)
	buf.WriteString(" ALL (")
	buf.WriteString(strings.Join(vs, ", "))
	buf.WriteString(")")
	return buf.String()
}

// Some represents "field op SOME (value...)".
func (c *Cond) Some(field, op string, value ...interface{}) string {
	vs := make([]string, 0, len(value))

	for _, v := range value {
		vs = append(vs, c.Args.Add(v))
	}

	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" ")
	buf.WriteString(op)
	buf.WriteString(" SOME (")
	buf.WriteString(strings.Join(vs, ", "))
	buf.WriteString(")")
	return buf.String()
}

// Var returns a placeholder for value.
func (c *Cond) Var(value interface{}) string {
	return c.Args.Add(value)
}
