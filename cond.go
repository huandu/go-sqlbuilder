// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

// Cond provides several helper methods to build conditions.
type Cond struct {
	Args *Args
}

// NewCond returns a new Cond.
func NewCond() *Cond {
	return &Cond{
		Args: &Args{},
	}
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

// EQ is an alias of Equal.
func (c *Cond) EQ(field string, value interface{}) string {
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

// NEQ is an alias of NotEqual.
func (c *Cond) NEQ(field string, value interface{}) string {
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

// GT is an alias of GreaterThan.
func (c *Cond) GT(field string, value interface{}) string {
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

// GTE is an alias of GreaterEqualThan.
func (c *Cond) GTE(field string, value interface{}) string {
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

// LT is an alias of LessThan.
func (c *Cond) LT(field string, value interface{}) string {
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

// LTE is an alias of LessEqualThan.
func (c *Cond) LTE(field string, value interface{}) string {
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
	buf.WriteStrings(vs, ", ")
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
	buf.WriteStrings(vs, ", ")
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

// ILike represents "field ILIKE value".
func (c *Cond) ILike(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" ILIKE ")
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
	buf.WriteStrings(orExpr, " OR ")
	buf.WriteString(")")
	return buf.String()
}

// And represents AND logic like "expr1 AND expr2 AND expr3".
func (c *Cond) And(andExpr ...string) string {
	buf := newStringBuilder()
	buf.WriteString("(")
	buf.WriteStrings(andExpr, " AND ")
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
	buf.WriteStrings(vs, ", ")
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
	buf.WriteStrings(vs, ", ")
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
	buf.WriteStrings(vs, ", ")
	buf.WriteString(")")
	return buf.String()
}

// Var returns a placeholder for value.
func (c *Cond) Var(value interface{}) string {
	return c.Args.Add(value)
}

func (c *Cond) MatchAll(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" MATCH_ALL ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}

func (c *Cond) MatchAny(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" MATCH_ANY ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}

func (c *Cond) MatchPhrase(field, slop string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" MATCH_PHRASE ")
	buf.WriteString(c.Args.Add(value))
	if slop != "" {
		buf.WriteString(" " + slop)
	}
	return buf.String()
}

func (c *Cond) MatchPhrasePrefix(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" MATCH_PHRASE_PREFIX ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}

func (c *Cond) MatchRegexp(field string, value interface{}) string {
	buf := newStringBuilder()
	buf.WriteString(Escape(field))
	buf.WriteString(" MATCH_REGEXP ")
	buf.WriteString(c.Args.Add(value))
	return buf.String()
}
