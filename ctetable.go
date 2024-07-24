// Copyright 2024 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

const (
	cteTableMarkerInit injectionMarker = iota
	cteTableMarkerAfterTable
	cteTableMarkerAfterAs
)

// CTETable creates a new CTE table builder with default flavor.
func CTETable(name string, cols ...string) *CTETableBuilder {
	return DefaultFlavor.NewCTETableBuilder().Table(name, cols...)
}

func newCTETableBuilder() *CTETableBuilder {
	return &CTETableBuilder{
		args:      &Args{},
		injection: newInjection(),
	}
}

// CTETableBuilder is a builder to build one table in CTE (Common Table Expression).
type CTETableBuilder struct {
	name       string
	cols       []string
	builderVar string

	args *Args

	injection *injection
	marker    injectionMarker
}

// Table sets the table name and columns in a CTE table.
func (ctetb *CTETableBuilder) Table(name string, cols ...string) *CTETableBuilder {
	ctetb.name = name
	ctetb.cols = cols
	ctetb.marker = cteTableMarkerAfterTable
	return ctetb
}

// As sets the builder to select data.
func (ctetb *CTETableBuilder) As(builder Builder) *CTETableBuilder {
	ctetb.builderVar = ctetb.args.Add(builder)
	ctetb.marker = cteTableMarkerAfterAs
	return ctetb
}

// String returns the compiled CTE string.
func (ctetb *CTETableBuilder) String() string {
	sql, _ := ctetb.Build()
	return sql
}

// Build returns compiled CTE string and args.
func (ctetb *CTETableBuilder) Build() (sql string, args []interface{}) {
	return ctetb.BuildWithFlavor(ctetb.args.Flavor)
}

// BuildWithFlavor builds a CTE with the specified flavor and initial arguments.
func (ctetb *CTETableBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
	buf := newStringBuilder()
	ctetb.injection.WriteTo(buf, cteTableMarkerInit)

	if ctetb.name != "" {
		buf.WriteLeadingString(ctetb.name)

		if len(ctetb.cols) > 0 {
			buf.WriteLeadingString("(")
			buf.WriteStrings(ctetb.cols, ", ")
			buf.WriteString(")")
		}

		ctetb.injection.WriteTo(buf, cteTableMarkerAfterTable)
	}

	if ctetb.builderVar != "" {
		buf.WriteLeadingString("AS (")
		buf.WriteString(ctetb.builderVar)
		buf.WriteRune(')')

		ctetb.injection.WriteTo(buf, cteTableMarkerAfterAs)
	}

	return ctetb.args.CompileWithFlavor(buf.String(), flavor, initialArg...)
}

// SetFlavor sets the flavor of compiled sql.
func (ctetb *CTETableBuilder) SetFlavor(flavor Flavor) (old Flavor) {
	old = ctetb.args.Flavor
	ctetb.args.Flavor = flavor
	return
}

// SQL adds an arbitrary sql to current position.
func (ctetb *CTETableBuilder) SQL(sql string) *CTETableBuilder {
	ctetb.injection.SQL(ctetb.marker, sql)
	return ctetb
}

// TableName returns the CTE table name.
func (ctetb *CTETableBuilder) TableName() string {
	return ctetb.name
}
