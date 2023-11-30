// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	selectMarkerInit injectionMarker = iota
	selectMarkerAfterSelect
	selectMarkerAfterFrom
	selectMarkerAfterJoin
	selectMarkerAfterWhere
	selectMarkerAfterGroupBy
	selectMarkerAfterOrderBy
	selectMarkerAfterLimit
	selectMarkerAfterFor
)

// JoinOption is the option in JOIN.
type JoinOption string

// Join options.
const (
	FullJoin       JoinOption = "FULL"
	FullOuterJoin  JoinOption = "FULL OUTER"
	InnerJoin      JoinOption = "INNER"
	LeftJoin       JoinOption = "LEFT"
	LeftOuterJoin  JoinOption = "LEFT OUTER"
	RightJoin      JoinOption = "RIGHT"
	RightOuterJoin JoinOption = "RIGHT OUTER"
)

// NewSelectBuilder creates a new SELECT builder.
func NewSelectBuilder() *SelectBuilder {
	return DefaultFlavor.NewSelectBuilder()
}

func newSelectBuilder() *SelectBuilder {
	args := &Args{}
	return &SelectBuilder{
		Cond: Cond{
			Args: args,
		},
		limit:     -1,
		offset:    -1,
		args:      args,
		injection: newInjection(),
	}
}

// SelectBuilder is a builder to build SELECT.
type SelectBuilder struct {
	Cond

	distinct    bool
	tables      []string
	selectCols  []string
	joinOptions []JoinOption
	joinTables  []string
	joinExprs   [][]string
	whereExprs  []string
	havingExprs []string
	groupByCols []string
	orderByCols []string
	order       string
	limit       int
	offset      int
	forWhat     string

	args *Args

	injection *injection
	marker    injectionMarker
}

var _ Builder = new(SelectBuilder)

// Select sets columns in SELECT.
func Select(col ...string) *SelectBuilder {
	return DefaultFlavor.NewSelectBuilder().Select(col...)
}

// Select sets columns in SELECT.
func (sb *SelectBuilder) Select(col ...string) *SelectBuilder {
	sb.selectCols = col
	sb.marker = selectMarkerAfterSelect
	return sb
}

// Distinct marks this SELECT as DISTINCT.
func (sb *SelectBuilder) Distinct() *SelectBuilder {
	sb.distinct = true
	sb.marker = selectMarkerAfterSelect
	return sb
}

// From sets table names in SELECT.
func (sb *SelectBuilder) From(table ...string) *SelectBuilder {
	sb.tables = table
	sb.marker = selectMarkerAfterFrom
	return sb
}

// Join sets expressions of JOIN in SELECT.
//
// It builds a JOIN expression like
//
//	JOIN table ON onExpr[0] AND onExpr[1] ...
func (sb *SelectBuilder) Join(table string, onExpr ...string) *SelectBuilder {
	sb.marker = selectMarkerAfterJoin
	return sb.JoinWithOption("", table, onExpr...)
}

// JoinWithOption sets expressions of JOIN with an option.
//
// It builds a JOIN expression like
//
//	option JOIN table ON onExpr[0] AND onExpr[1] ...
//
// Here is a list of supported options.
//   - FullJoin: FULL JOIN
//   - FullOuterJoin: FULL OUTER JOIN
//   - InnerJoin: INNER JOIN
//   - LeftJoin: LEFT JOIN
//   - LeftOuterJoin: LEFT OUTER JOIN
//   - RightJoin: RIGHT JOIN
//   - RightOuterJoin: RIGHT OUTER JOIN
func (sb *SelectBuilder) JoinWithOption(option JoinOption, table string, onExpr ...string) *SelectBuilder {
	sb.joinOptions = append(sb.joinOptions, option)
	sb.joinTables = append(sb.joinTables, table)
	sb.joinExprs = append(sb.joinExprs, onExpr)
	sb.marker = selectMarkerAfterJoin
	return sb
}

// Where sets expressions of WHERE in SELECT.
func (sb *SelectBuilder) Where(andExpr ...string) *SelectBuilder {
	sb.whereExprs = append(sb.whereExprs, andExpr...)
	sb.marker = selectMarkerAfterWhere
	return sb
}

// Having sets expressions of HAVING in SELECT.
func (sb *SelectBuilder) Having(andExpr ...string) *SelectBuilder {
	sb.havingExprs = append(sb.havingExprs, andExpr...)
	sb.marker = selectMarkerAfterGroupBy
	return sb
}

// GroupBy sets columns of GROUP BY in SELECT.
func (sb *SelectBuilder) GroupBy(col ...string) *SelectBuilder {
	sb.groupByCols = append(sb.groupByCols, col...)
	sb.marker = selectMarkerAfterGroupBy
	return sb
}

// OrderBy sets columns of ORDER BY in SELECT.
func (sb *SelectBuilder) OrderBy(col ...string) *SelectBuilder {
	sb.orderByCols = append(sb.orderByCols, col...)
	sb.marker = selectMarkerAfterOrderBy
	return sb
}

// Asc sets order of ORDER BY to ASC.
func (sb *SelectBuilder) Asc() *SelectBuilder {
	sb.order = "ASC"
	sb.marker = selectMarkerAfterOrderBy
	return sb
}

// Desc sets order of ORDER BY to DESC.
func (sb *SelectBuilder) Desc() *SelectBuilder {
	sb.order = "DESC"
	sb.marker = selectMarkerAfterOrderBy
	return sb
}

// Limit sets the LIMIT in SELECT.
func (sb *SelectBuilder) Limit(limit int) *SelectBuilder {
	sb.limit = limit
	sb.marker = selectMarkerAfterLimit
	return sb
}

// Offset sets the LIMIT offset in SELECT.
func (sb *SelectBuilder) Offset(offset int) *SelectBuilder {
	sb.offset = offset
	sb.marker = selectMarkerAfterLimit
	return sb
}

// ForUpdate adds FOR UPDATE at the end of SELECT statement.
func (sb *SelectBuilder) ForUpdate() *SelectBuilder {
	sb.forWhat = "UPDATE"
	sb.marker = selectMarkerAfterFor
	return sb
}

// ForShare adds FOR SHARE at the end of SELECT statement.
func (sb *SelectBuilder) ForShare() *SelectBuilder {
	sb.forWhat = "SHARE"
	sb.marker = selectMarkerAfterFor
	return sb
}

// As returns an AS expression.
func (sb *SelectBuilder) As(name, alias string) string {
	return fmt.Sprintf("%s AS %s", name, alias)
}

// BuilderAs returns an AS expression wrapping a complex SQL.
// According to SQL syntax, SQL built by builder is surrounded by parens.
func (sb *SelectBuilder) BuilderAs(builder Builder, alias string) string {
	return fmt.Sprintf("(%s) AS %s", sb.Var(builder), alias)
}

// NumCol returns the number of columns to select.
func (sb *SelectBuilder) NumCol() int {
	return len(sb.selectCols)
}

// String returns the compiled SELECT string.
func (sb *SelectBuilder) String() string {
	s, _ := sb.Build()
	return s
}

// Build returns compiled SELECT string and args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (sb *SelectBuilder) Build() (sql string, args []interface{}) {
	return sb.BuildWithFlavor(sb.args.Flavor)
}

// BuildWithFlavor returns compiled SELECT string and args with flavor and initial args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (sb *SelectBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
	buf := newStringBuilder()
	sb.injection.WriteTo(buf, selectMarkerInit)

	oraclePage := flavor == Oracle && (sb.limit >= 0 || sb.offset >= 0)

	if len(sb.selectCols) > 0 {
		buf.WriteLeadingString("SELECT ")

		if sb.distinct {
			buf.WriteString("DISTINCT ")
		}

		if oraclePage {
			var selectCols = make([]string, 0, len(sb.selectCols))
			for i := range sb.selectCols {
				cols := strings.SplitN(sb.selectCols[i], ".", 2)
				if len(cols) == 1 {
					selectCols = append(selectCols, cols[0])
				} else {
					selectCols = append(selectCols, cols[1])
				}
			}
			buf.WriteString(strings.Join(selectCols, ", "))
		} else {
			buf.WriteString(strings.Join(sb.selectCols, ", "))
		}
	}

	sb.injection.WriteTo(buf, selectMarkerAfterSelect)

	if oraclePage {
		if len(sb.selectCols) > 0 {
			buf.WriteLeadingString("FROM ( SELECT ")

			if sb.distinct {
				buf.WriteString("DISTINCT ")
			}

			var selectCols = make([]string, 0, len(sb.selectCols)+1)
			selectCols = append(selectCols, "ROWNUM r")
			for i := range sb.selectCols {
				cols := strings.SplitN(sb.selectCols[i], ".", 2)
				if len(cols) == 1 {
					selectCols = append(selectCols, cols[0])
				} else {
					selectCols = append(selectCols, cols[1])
				}
			}
			buf.WriteString(strings.Join(selectCols, ", "))

			buf.WriteLeadingString("FROM ( SELECT ")
			buf.WriteString(strings.Join(sb.selectCols, ", "))
		}
	}

	if len(sb.tables) > 0 {
		buf.WriteLeadingString("FROM ")
		buf.WriteString(strings.Join(sb.tables, ", "))
	}

	sb.injection.WriteTo(buf, selectMarkerAfterFrom)

	for i := range sb.joinTables {
		if option := sb.joinOptions[i]; option != "" {
			buf.WriteLeadingString(string(option))
		}

		buf.WriteLeadingString("JOIN ")
		buf.WriteString(sb.joinTables[i])

		if exprs := sb.joinExprs[i]; len(exprs) > 0 {
			buf.WriteString(" ON ")
			buf.WriteString(strings.Join(sb.joinExprs[i], " AND "))
		}
	}

	if len(sb.joinTables) > 0 {
		sb.injection.WriteTo(buf, selectMarkerAfterJoin)
	}

	if len(sb.whereExprs) > 0 {
		buf.WriteLeadingString("WHERE ")
		buf.WriteString(strings.Join(sb.whereExprs, " AND "))

		sb.injection.WriteTo(buf, selectMarkerAfterWhere)
	}

	if len(sb.groupByCols) > 0 {
		buf.WriteLeadingString("GROUP BY ")
		buf.WriteString(strings.Join(sb.groupByCols, ", "))

		if len(sb.havingExprs) > 0 {
			buf.WriteString(" HAVING ")
			buf.WriteString(strings.Join(sb.havingExprs, " AND "))
		}

		sb.injection.WriteTo(buf, selectMarkerAfterGroupBy)
	}

	if len(sb.orderByCols) > 0 {
		buf.WriteLeadingString("ORDER BY ")
		buf.WriteString(strings.Join(sb.orderByCols, ", "))

		if sb.order != "" {
			buf.WriteRune(' ')
			buf.WriteString(sb.order)
		}

		sb.injection.WriteTo(buf, selectMarkerAfterOrderBy)
	}

	switch flavor {
	case MySQL, SQLite, ClickHouse:
		if sb.limit >= 0 {
			buf.WriteLeadingString("LIMIT ")
			buf.WriteString(strconv.Itoa(sb.limit))

			if sb.offset >= 0 {
				buf.WriteLeadingString("OFFSET ")
				buf.WriteString(strconv.Itoa(sb.offset))
			}
		}
	case CQL:
		if sb.limit >= 0 {
			buf.WriteLeadingString("LIMIT ")
			buf.WriteString(strconv.Itoa(sb.limit))
		}
	case PostgreSQL, Presto:
		if sb.limit >= 0 {
			buf.WriteLeadingString("LIMIT ")
			buf.WriteString(strconv.Itoa(sb.limit))
		}

		if sb.offset >= 0 {
			buf.WriteLeadingString("OFFSET ")
			buf.WriteString(strconv.Itoa(sb.offset))
		}

	case SQLServer:
		// If ORDER BY is not set, sort column #1 by default.
		// It's required to make OFFSET...FETCH work.
		if len(sb.orderByCols) == 0 && (sb.limit >= 0 || sb.offset >= 0) {
			buf.WriteLeadingString("ORDER BY 1")
		}

		if sb.offset >= 0 {
			buf.WriteLeadingString("OFFSET ")
			buf.WriteString(strconv.Itoa(sb.offset))
			buf.WriteString(" ROWS")
		}

		if sb.limit >= 0 {
			if sb.offset < 0 {
				buf.WriteLeadingString("OFFSET 0 ROWS")
			}

			buf.WriteLeadingString("FETCH NEXT ")
			buf.WriteString(strconv.Itoa(sb.limit))
			buf.WriteString(" ROWS ONLY")
		}

	case Oracle:
		if oraclePage {
			buf.WriteString(" ) ")
			if len(sb.tables) > 0 {
				buf.WriteString(strings.Join(sb.tables, ", "))
			}

			min := sb.offset
			if min < 0 {
				min = 0
			}

			buf.WriteString(" ) WHERE ")
			if sb.limit >= 0 {
				buf.WriteString("r BETWEEN ")
				buf.WriteString(strconv.Itoa(min + 1))
				buf.WriteString(" AND ")
				buf.WriteString(strconv.Itoa(sb.limit + min))
			} else {
				buf.WriteString("r >= ")
				buf.WriteString(strconv.Itoa(min + 1))
			}
		}
	}

	if sb.limit >= 0 {
		sb.injection.WriteTo(buf, selectMarkerAfterLimit)
	}

	if sb.forWhat != "" {
		buf.WriteLeadingString("FOR ")
		buf.WriteString(sb.forWhat)

		sb.injection.WriteTo(buf, selectMarkerAfterFor)
	}

	return sb.args.CompileWithFlavor(buf.String(), flavor, initialArg...)
}

// SetFlavor sets the flavor of compiled sql.
func (sb *SelectBuilder) SetFlavor(flavor Flavor) (old Flavor) {
	old = sb.args.Flavor
	sb.args.Flavor = flavor
	return
}

// SQL adds an arbitrary sql to current position.
func (sb *SelectBuilder) SQL(sql string) *SelectBuilder {
	sb.injection.SQL(sb.marker, sql)
	return sb
}
