// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"strconv"
	"strings"
)

const (
	unionDistinct = " UNION " // Default union type is DISTINCT.
	unionAll      = " UNION ALL "
)

// UnionBuilder is a builder to build UNION.
type UnionBuilder struct {
	format      string
	builders    []Builder
	orderByCols []string
	order       string
	limit       int
	offset      int

	args *Args
}

var _ Builder = new(UnionBuilder)

// Union unions all builders together using UNION operator.
func Union(builders ...Builder) *UnionBuilder {
	return DefaultFlavor.Union(builders...)
}

// UnionAll unions all builders together using UNION ALL operator.
func UnionAll(builders ...Builder) *UnionBuilder {
	return DefaultFlavor.UnionAll(builders...)
}

func newUnionBuilder(opt string, builders ...Builder) *UnionBuilder {
	args := &Args{}
	vars := make([]string, 0, len(builders))

	for _, b := range builders {
		vars = append(vars, args.Add(b))
	}

	return &UnionBuilder{
		format:   strings.Join(vars, opt),
		builders: builders,
		limit:    -1,
		offset:   -1,

		args: args,
	}
}

// OrderBy sets columns of ORDER BY in SELECT.
func (ub *UnionBuilder) OrderBy(col ...string) *UnionBuilder {
	ub.orderByCols = col
	return ub
}

// Asc sets order of ORDER BY to ASC.
func (ub *UnionBuilder) Asc() *UnionBuilder {
	ub.order = "ASC"
	return ub
}

// Desc sets order of ORDER BY to DESC.
func (ub *UnionBuilder) Desc() *UnionBuilder {
	ub.order = "DESC"
	return ub
}

// Limit sets the LIMIT in SELECT.
func (ub *UnionBuilder) Limit(limit int) *UnionBuilder {
	ub.limit = limit
	return ub
}

// Offset sets the LIMIT offset in SELECT.
func (ub *UnionBuilder) Offset(offset int) *UnionBuilder {
	ub.offset = offset
	return ub
}

// String returns the compiled SELECT string.
func (ub *UnionBuilder) String() string {
	s, _ := ub.Build()
	return s
}

// Build returns compiled SELECT string and args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (ub *UnionBuilder) Build() (sql string, args []interface{}) {
	return ub.BuildWithFlavor(ub.args.Flavor)
}

// BuildWithFlavor returns compiled SELECT string and args with flavor and initial args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (ub *UnionBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
	buf := &bytes.Buffer{}

	if len(ub.builders) > 1 {
		buf.WriteRune('(')
	}

	buf.WriteString(ub.format)

	if len(ub.builders) > 1 {
		buf.WriteRune(')')
	}

	if len(ub.orderByCols) > 0 {
		buf.WriteString(" ORDER BY ")
		buf.WriteString(strings.Join(ub.orderByCols, ", "))

		if ub.order != "" {
			buf.WriteRune(' ')
			buf.WriteString(ub.order)
		}
	}

	if ub.limit >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.Itoa(ub.limit))
	}

	if MySQL == flavor && ub.limit >= 0 || PostgreSQL == flavor {
		if ub.offset >= 0 {
			buf.WriteString(" OFFSET ")
			buf.WriteString(strconv.Itoa(ub.offset))
		}
	}

	return ub.args.CompileWithFlavor(buf.String(), flavor, initialArg...)
}

// SetFlavor sets the flavor of compiled sql.
func (ub *UnionBuilder) SetFlavor(flavor Flavor) (old Flavor) {
	old = ub.args.Flavor
	ub.args.Flavor = flavor
	return
}
