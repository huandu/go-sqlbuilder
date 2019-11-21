// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"reflect"
	"strings"
)

var (
	// DBTag is the struct tag to describe the name for a field in struct.
	DBTag = "db"

	// FieldTag is the struct tag to describe the tag name for a field in struct.
	// Use "," to separate different tags.
	FieldTag = "fieldtag"

	// FieldOpt is the options for a struct field.
	// As db column can contain "," in theory, field options should be provided in a separated tag.
	FieldOpt = "fieldopt"
)

const (
	fieldOptWithQuote = "withquote"
	fieldOptOmitEmpty = "omitempty"
)

// Struct represents a struct type.
//
// All methods in Struct are thread-safe.
// We can define a global variable to hold a Struct and use it in any goroutine.
type Struct struct {
	Flavor Flavor

	structType      reflect.Type
	fieldAlias      map[string]string
	taggedFields    map[string][]string
	quotedFields    map[string]struct{}
	omitEmptyFields map[string]struct{}
}

// NewStruct analyzes type information in structValue
// and creates a new Struct with all structValue fields.
// If structValue is not a struct, NewStruct returns a dummy Sturct.
func NewStruct(structValue interface{}) *Struct {
	t := reflect.TypeOf(structValue)
	t = dereferencedType(t)
	s := &Struct{
		Flavor: DefaultFlavor,
	}

	if t.Kind() != reflect.Struct {
		return s
	}

	s.structType = t
	s.fieldAlias = map[string]string{}
	s.taggedFields = map[string][]string{}
	s.quotedFields = map[string]struct{}{}
	s.omitEmptyFields = map[string]struct{}{}
	s.parse(t)
	return s
}

// For sets the default flavor of s.
func (s *Struct) For(flavor Flavor) *Struct {
	s.Flavor = flavor
	return s
}

func (s *Struct) parse(t reflect.Type) {
	l := t.NumField()

	for i := 0; i < l; i++ {
		field := t.Field(i)

		if field.Anonymous {
			ft := dereferencedType(field.Type)
			s.parse(ft)
			continue
		}

		// Parse DBTag.
		dbtag := field.Tag.Get(DBTag)
		alias := dbtag

		if dbtag == "-" {
			continue
		}

		if dbtag == "" {
			alias = field.Name
			s.fieldAlias[field.Name] = field.Name
		} else {
			s.fieldAlias[dbtag] = field.Name
		}

		// Parse FieldTag.
		fieldtag := field.Tag.Get(FieldTag)
		tags := strings.Split(fieldtag, ",")

		for _, t := range tags {
			if t != "" {
				s.taggedFields[t] = append(s.taggedFields[t], alias)
			}
		}

		s.taggedFields[""] = append(s.taggedFields[""], alias)

		// Parse FieldOpt.
		fieldopt := field.Tag.Get(FieldOpt)
		opts := strings.Split(fieldopt, ",")

		for _, opt := range opts {
			switch opt {
			case fieldOptWithQuote:
				s.quotedFields[alias] = struct{}{}
			case fieldOptOmitEmpty:
				s.omitEmptyFields[alias] = struct{}{}
			}
		}
	}
}

// SelectFrom creates a new `SelectBuilder` with table name.
// By default, all exported fields of the s are listed as columns in SELECT.
//
// Caller is responsible to set WHERE condition to find right record.
func (s *Struct) SelectFrom(table string) *SelectBuilder {
	return s.SelectFromForTag(table, "")
}

// SelectFromForTag creates a new `SelectBuilder` with table name for a specified tag.
// By default, all fields of the s tagged with tag are listed as columns in SELECT.
//
// Caller is responsible to set WHERE condition to find right record.
func (s *Struct) SelectFromForTag(table string, tag string) *SelectBuilder {
	sb := s.Flavor.NewSelectBuilder()
	sb.From(table)

	if s.taggedFields == nil {
		return sb
	}

	fields, ok := s.taggedFields[tag]

	if ok {
		fields = s.quoteFields(fields)

		buf := &bytes.Buffer{}
		cols := make([]string, 0, len(fields))

		for _, field := range fields {
			buf.WriteString(table)
			buf.WriteRune('.')
			buf.WriteString(field)
			cols = append(cols, buf.String())
			buf.Reset()
		}

		sb.Select(cols...)
	} else {
		sb.Select("*")
	}

	return sb
}

// Update creates a new `UpdateBuilder` with table name.
// By default, all exported fields of the s is assigned in UPDATE with the field values from value.
// If value's type is not the same as that of s, Update returns a dummy `UpdateBuilder` with table name.
//
// Caller is responsible to set WHERE condition to match right record.
func (s *Struct) Update(table string, value interface{}) *UpdateBuilder {
	return s.UpdateForTag(table, "", value)
}

// UpdateForTag creates a new `UpdateBuilder` with table name.
// By default, all fields of the s tagged with tag is assigned in UPDATE with the field values from value.
// If value's type is not the same as that of s, UpdateForTag returns a dummy `UpdateBuilder` with table name.
//
// Caller is responsible to set WHERE condition to match right record.
func (s *Struct) UpdateForTag(table string, tag string, value interface{}) *UpdateBuilder {
	ub := s.Flavor.NewUpdateBuilder()
	ub.Update(table)

	if s.taggedFields == nil {
		return ub
	}

	fields, ok := s.taggedFields[tag]

	if !ok {
		return ub
	}

	v := reflect.ValueOf(value)
	v = dereferencedValue(v)

	if v.Type() != s.structType {
		return ub
	}

	quoted := s.quoteFields(fields)
	assignments := make([]string, 0, len(fields))

	for i, f := range fields {
		name := s.fieldAlias[f]
		val := v.FieldByName(name)

		if isEmptyValue(val) {
			if _, ok := s.omitEmptyFields[f]; ok {
				continue
			}
		} else {
			val = dereferencedValue(val)
		}
		data := val.Interface()
		assignments = append(assignments, ub.Assign(quoted[i], data))
	}

	ub.Set(assignments...)
	return ub
}

// InsertInto creates a new `InsertBuilder` with table name using verb INSERT INTO.
// By default, all exported fields of s are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// InsertInto never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) InsertInto(table string, value ...interface{}) *InsertBuilder {
	return s.InsertIntoForTag(table, "", value...)
}

// InsertIgnoreInto creates a new `InsertBuilder` with table name using verb INSERT IGNORE INTO.
// By default, all exported fields of s are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// InsertIgnoreInto never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) InsertIgnoreInto(table string, value ...interface{}) *InsertBuilder {
	return s.InsertIgnoreIntoForTag(table, "", value...)
}

// ReplaceInto creates a new `InsertBuilder` with table name using verb REPLACE INTO.
// By default, all exported fields of s are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// ReplaceInto never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) ReplaceInto(table string, value ...interface{}) *InsertBuilder {
	return s.ReplaceIntoForTag(table, "", value...)
}

// buildColsAndValuesForTag uses ib to set exported fields tagged with tag as columns
// and add value as a list of values.
func (s *Struct) buildColsAndValuesForTag(ib *InsertBuilder, tag string, value ...interface{}) {
	if s.taggedFields == nil {
		return
	}

	fields, ok := s.taggedFields[tag]

	if !ok {
		return
	}

	vs := make([]reflect.Value, 0, len(value))

	for _, item := range value {
		v := reflect.ValueOf(item)
		v = dereferencedValue(v)

		if v.Type() == s.structType {
			vs = append(vs, v)
		}
	}

	if len(vs) == 0 {
		return
	}
	cols := make([]string, 0, len(fields))
	values := make([][]interface{}, len(vs))

	for _, f := range fields {
		cols = append(cols, f)
		name := s.fieldAlias[f]

		for i, v := range vs {
			data := v.FieldByName(name).Interface()
			values[i] = append(values[i], data)
		}
	}

	cols = s.quoteFields(cols)
	ib.Cols(cols...)

	for _, value := range values {
		ib.Values(value...)
	}
}

// InsertIntoForTag creates a new `InsertBuilder` with table name using verb INSERT INTO.
// By default, exported fields tagged with tag are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// InsertIntoForTag never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) InsertIntoForTag(table string, tag string, value ...interface{}) *InsertBuilder {
	ib := s.Flavor.NewInsertBuilder()
	ib.InsertInto(table)

	s.buildColsAndValuesForTag(ib, tag, value...)
	return ib
}

// InsertIgnoreIntoForTag creates a new `InsertBuilder` with table name using verb INSERT IGNORE INTO.
// By default, exported fields tagged with tag are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// InsertIgnoreIntoForTag never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) InsertIgnoreIntoForTag(table string, tag string, value ...interface{}) *InsertBuilder {
	ib := s.Flavor.NewInsertBuilder()
	ib.InsertIgnoreInto(table)

	s.buildColsAndValuesForTag(ib, tag, value...)
	return ib
}

// ReplaceIntoForTag creates a new `InsertBuilder` with table name using verb REPLACE INTO.
// By default, exported fields tagged with tag are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// ReplaceIntoForTag never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) ReplaceIntoForTag(table string, tag string, value ...interface{}) *InsertBuilder {
	ib := s.Flavor.NewInsertBuilder()
	ib.ReplaceInto(table)

	s.buildColsAndValuesForTag(ib, tag, value...)
	return ib
}

// DeleteFrom creates a new `DeleteBuilder` with table name.
//
// Caller is responsible to set WHERE condition to match right record.
func (s *Struct) DeleteFrom(table string) *DeleteBuilder {
	db := s.Flavor.NewDeleteBuilder()
	db.DeleteFrom(table)
	return db
}

// Addr takes address of all exported fields of the s from the value.
// The returned result can be used in `Row#Scan` directly.
func (s *Struct) Addr(value interface{}) []interface{} {
	return s.AddrForTag("", value)
}

// AddrForTag takes address of all fields of the s tagged with tag from the value.
// The returned result can be used in `Row#Scan` directly.
//
// If tag is not defined in s in advance,
func (s *Struct) AddrForTag(tag string, value interface{}) []interface{} {
	fields, ok := s.taggedFields[tag]

	if !ok {
		return nil
	}

	return s.AddrWithCols(fields, value)
}

// AddrWithCols takes address of all columns defined in cols from the value.
// The returned result can be used in `Row#Scan` directly.
func (s *Struct) AddrWithCols(cols []string, value interface{}) []interface{} {
	v := reflect.ValueOf(value)
	v = dereferencedValue(v)

	if v.Type() != s.structType {
		return nil
	}

	for _, c := range cols {
		if _, ok := s.fieldAlias[c]; !ok {
			return nil
		}
	}

	addrs := make([]interface{}, 0, len(cols))

	for _, c := range cols {
		name := s.fieldAlias[c]
		data := v.FieldByName(name).Addr().Interface()
		addrs = append(addrs, data)
	}

	return addrs
}

func (s *Struct) quoteFields(fields []string) []string {
	// Try best not to allocate new slice.
	if len(s.quotedFields) == 0 {
		return fields
	}

	needQuote := false

	for _, field := range fields {
		if _, ok := s.quotedFields[field]; ok {
			needQuote = true
			break
		}
	}

	if !needQuote {
		return fields
	}

	quoted := make([]string, 0, len(fields))

	for _, field := range fields {
		if _, ok := s.quotedFields[field]; ok {
			quoted = append(quoted, s.Flavor.Quote(field))
		} else {
			quoted = append(quoted, field)
		}
	}

	return quoted
}

func dereferencedType(t reflect.Type) reflect.Type {
	for k := t.Kind(); k == reflect.Ptr || k == reflect.Interface; k = t.Kind() {
		t = t.Elem()
	}

	return t
}

func dereferencedValue(v reflect.Value) reflect.Value {
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}

	return v
}

func isEmptyValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Interface, reflect.Ptr, reflect.Chan, reflect.Func, reflect.Map, reflect.Slice:
		return value.IsNil()
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() == 0
	case reflect.String:
		return value.String() == ""
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	}

	return false
}
