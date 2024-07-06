// Copyright 2024 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

const (
	cteMarkerInit injectionMarker = iota
	cteMarkerAfterWith
	cteMarkerAfterAs
)

// With creates a new CTE builder with default flavor.
func With(name string, cols ...string) *CTEBuilder {
	return DefaultFlavor.NewCTEBuilder().With(name, cols...)
}

func newCTEBuilder() *CTEBuilder {
	return &CTEBuilder{
		args:      &Args{},
		injection: newInjection(),
	}
}

// CTEBuilder is a CTE (Common Table Expression) builder.
type CTEBuilder struct {
	name       string
	cols       []string
	builderVar string

	args *Args

	injection *injection
	marker    injectionMarker
}

var _ Builder = new(CTEBuilder)

// With sets the CTE name and columns.
func (cteb *CTEBuilder) With(name string, cols ...string) *CTEBuilder {
	cteb.name = name
	cteb.cols = cols
	cteb.marker = cteMarkerAfterWith
	return cteb
}

// As sets the builder to select data.
func (cteb *CTEBuilder) As(builder Builder) *CTEBuilder {
	cteb.builderVar = cteb.args.Add(builder)
	cteb.marker = cteMarkerAfterAs
	return cteb
}

// Select creates a new SelectBuilder to build a SELECT statement using this CTE.
func (cteb *CTEBuilder) Select(col ...string) *SelectBuilder {
	sb := cteb.args.Flavor.NewSelectBuilder()
	return sb.With(cteb).Select(col...)
}

// String returns the compiled CTE string.
func (cteb *CTEBuilder) String() string {
	sql, _ := cteb.Build()
	return sql
}

// Build returns compiled CTE string and args.
func (cteb *CTEBuilder) Build() (sql string, args []interface{}) {
	return cteb.BuildWithFlavor(cteb.args.Flavor)
}

// BuildWithFlavor builds a CTE with the specified flavor and initial arguments.
func (cteb *CTEBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
	buf := newStringBuilder()
	cteb.injection.WriteTo(buf, cteMarkerInit)

	if cteb.name != "" {
		buf.WriteLeadingString("WITH ")
		buf.WriteString(cteb.name)

		if len(cteb.cols) > 0 {
			buf.WriteLeadingString("(")
			buf.WriteStrings(cteb.cols, ", ")
			buf.WriteString(")")
		}

		cteb.injection.WriteTo(buf, cteMarkerAfterWith)
	}

	if cteb.builderVar != "" {
		buf.WriteLeadingString("AS (")
		buf.WriteString(cteb.builderVar)
		buf.WriteRune(')')

		cteb.injection.WriteTo(buf, cteMarkerAfterAs)
	}

	return cteb.args.CompileWithFlavor(buf.String(), flavor, initialArg...)
}

// SetFlavor sets the flavor of compiled sql.
func (cteb *CTEBuilder) SetFlavor(flavor Flavor) (old Flavor) {
	old = cteb.args.Flavor
	cteb.args.Flavor = flavor
	return
}

// Var returns a placeholder for value.
func (cteb *CTEBuilder) Var(arg interface{}) string {
	return cteb.args.Add(arg)
}

// SQL adds an arbitrary sql to current position.
func (cteb *CTEBuilder) SQL(sql string) *CTEBuilder {
	cteb.injection.SQL(cteb.marker, sql)
	return cteb
}

// TableName returns the CTE table name.
func (cteb *CTEBuilder) TableName() string {
	return cteb.name
}
