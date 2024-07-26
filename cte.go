// Copyright 2024 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

const (
	cteMarkerInit injectionMarker = iota
	cteMarkerAfterWith
)

// With creates a new CTE builder with default flavor.
func With(tables ...*CTETableBuilder) *CTEBuilder {
	return DefaultFlavor.NewCTEBuilder().With(tables...)
}

func newCTEBuilder() *CTEBuilder {
	return &CTEBuilder{
		args:      &Args{},
		injection: newInjection(),
	}
}

// CTEBuilder is a CTE (Common Table Expression) builder.
type CTEBuilder struct {
	tableNames       []string
	tableBuilderVars []string

	args *Args

	injection *injection
	marker    injectionMarker
}

var _ Builder = new(CTEBuilder)

// With sets the CTE name and columns.
func (cteb *CTEBuilder) With(tables ...*CTETableBuilder) *CTEBuilder {
	tableNames := make([]string, 0, len(tables))
	tableBuilderVars := make([]string, 0, len(tables))

	for _, table := range tables {
		tableNames = append(tableNames, table.TableName())
		tableBuilderVars = append(tableBuilderVars, cteb.args.Add(table))
	}

	cteb.tableNames = tableNames
	cteb.tableBuilderVars = tableBuilderVars
	cteb.marker = cteMarkerAfterWith
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

	if len(cteb.tableBuilderVars) > 0 {
		buf.WriteLeadingString("WITH ")
		buf.WriteStrings(cteb.tableBuilderVars, ", ")
	}

	cteb.injection.WriteTo(buf, cteMarkerAfterWith)
	return cteb.args.CompileWithFlavor(buf.String(), flavor, initialArg...)
}

// SetFlavor sets the flavor of compiled sql.
func (cteb *CTEBuilder) SetFlavor(flavor Flavor) (old Flavor) {
	old = cteb.args.Flavor
	cteb.args.Flavor = flavor
	return
}

// SQL adds an arbitrary sql to current position.
func (cteb *CTEBuilder) SQL(sql string) *CTEBuilder {
	cteb.injection.SQL(cteb.marker, sql)
	return cteb
}

// TableNames returns all table names in a CTE.
func (cteb *CTEBuilder) TableNames() []string {
	return cteb.tableNames
}
