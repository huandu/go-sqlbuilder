// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

// Supported flavors.
const (
	invalidFlavor Flavor = iota

	MySQL
	PostgreSQL
)

var (
	// DefaultFlavor is the default flavor for all builders.
	DefaultFlavor = MySQL
)

// Flavor is the flag to control the format of compiled sql.
type Flavor int

// String returns the name of f.
func (f Flavor) String() string {
	switch f {
	case MySQL:
		return "MySQL"
	case PostgreSQL:
		return "PostgreSQL"
	}

	return "<invalid>"
}

// NewDeleteBuilder creates a new DELETE builder with flavor.
func (f Flavor) NewDeleteBuilder() *DeleteBuilder {
	b := newDeleteBuilder()
	b.SetFlavor(f)
	return b
}

// NewInsertBuilder creates a new INSERT builder with flavor.
func (f Flavor) NewInsertBuilder() *InsertBuilder {
	b := newInsertBuilder()
	b.SetFlavor(f)
	return b
}

// NewSelectBuilder creates a new SELECT builder with flavor.
func (f Flavor) NewSelectBuilder() *SelectBuilder {
	b := newSelectBuilder()
	b.SetFlavor(f)
	return b
}

// NewUpdateBuilder creates a new UPDATE builder with flavor.
func (f Flavor) NewUpdateBuilder() *UpdateBuilder {
	b := newUpdateBuilder()
	b.SetFlavor(f)
	return b
}
