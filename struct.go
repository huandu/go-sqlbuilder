// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"reflect"
	"strings"
)

var (
	// DBTag is the struct tag to describe the name for a field in struct.
	DBTag = "db"

	// FieldTag is the struct tag to describe the tag name for a field in struct.
	// Use "," to separate different tags.
	FieldTag = "fieldtag"
)

// Struct represents a struct type.
//
// All methods in Struct are thread-safe.
// We can define a global variable to hold a Struct and use it in any goroutine.
type Struct struct {
	structType   reflect.Type
	fieldAlias   map[string]string
	taggedFields map[string][]string
}

// NewStruct analyzes type information in structValue
// and creates a new Struct with all structValue fields.
// If structValue is not a struct, NewStruct returns a dummy Sturct.
func NewStruct(structValue interface{}) *Struct {
	t := reflect.TypeOf(structValue)
	t = deferencedType(t)
	s := new(Struct)

	if t.Kind() != reflect.Struct {
		return s
	}

	s.structType = t
	s.fieldAlias = map[string]string{}
	s.taggedFields = map[string][]string{}
	s.parse(t)
	return s
}

func (s *Struct) parse(t reflect.Type) {
	l := t.NumField()

	for i := 0; i < l; i++ {
		field := t.Field(i)

		if field.Anonymous {
			ft := deferencedType(field.Type)
			s.parse(ft)
			continue
		}

		// Parse DBTag.
		dbtag := field.Tag.Get(DBTag)
		aliasIdx := strings.Index(dbtag, ",")
		alias := field.Name

		if aliasIdx == 0 || dbtag == "" {
			s.fieldAlias[field.Name] = field.Name
		} else {
			if aliasIdx > 0 {
				alias = dbtag[:aliasIdx]
			} else {
				alias = dbtag
			}

			// Skip the field if DBTag is "-".
			if alias == "-" {
				continue
			}

			s.fieldAlias[alias] = field.Name
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
	}
}

// Select creates a new `SelectBuilder` with table name.
// By default, all exported fields of the s are listed as columns in SELECT
// and LIMIT is set to 1.
//
// Caller is responsible to set WHERE condition to find right record.
func (s *Struct) Select(table string) *SelectBuilder {
	return s.SelectForTag(table, "")
}

// SelectForTag creates a new `SelectBuilder` with table name for a specified tag.
// By default, all fields of the s tagged with tag are listed as columns in SELECT
// and LIMIT is set to 1.
//
// Caller is responsible to set WHERE condition to find right record.
func (s *Struct) SelectForTag(table string, tag string) *SelectBuilder {
	sb := NewSelectBuilder()
	sb.From(table)
	sb.Limit(1)

	if s.taggedFields == nil {
		return sb
	}

	fields, ok := s.taggedFields[tag]

	if ok {
		sb.Select(EscapeAll(fields...)...)
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
// If value's type is not the same as that of s, Update returns a dummy `UpdateBuilder` with table name.
//
// Caller is responsible to set WHERE condition to match right record.
func (s *Struct) UpdateForTag(table string, tag string, value interface{}) *UpdateBuilder {
	ub := NewUpdateBuilder()
	ub.Update(table)

	if s.taggedFields == nil {
		return ub
	}

	fields, ok := s.taggedFields[tag]

	if !ok {
		return ub
	}

	v := dereferencedValue(value)

	if v.Type() != s.structType {
		return ub
	}

	assignments := make([]string, 0, len(fields))

	for _, f := range fields {
		name := s.fieldAlias[f]
		data := v.FieldByName(name).Interface()
		assignments = append(assignments, ub.Assign(f, data))
	}

	ub.Set(assignments...)
	return ub
}

// InsertInto creates a new `InsertBuilder` with table name.
// By default, all exported fields of the s is inserted in INSERT with the field values from value.
// If value's type is not the same as that of s, Update returns a dummy `InsertBuilder` with table name.
func (s *Struct) InsertInto(table string, value interface{}) *InsertBuilder {
	return s.InsertIntoForTag(table, "", value)
}

// InsertIntoForTag creates a new `InsertBuilder` with table name.
// By default, all fields of the s tagged with tag is inserted in INSERT with the field values from value.
// If value's type is not the same as that of s, Update returns a dummy `InsertBuilder` with table name.
func (s *Struct) InsertIntoForTag(table string, tag string, value interface{}) *InsertBuilder {
	ib := NewInsertBuilder()
	ib.InsertInto(table)

	if s.taggedFields == nil {
		return ib
	}

	fields, ok := s.taggedFields[tag]

	if !ok {
		return ib
	}

	v := dereferencedValue(value)

	if v.Type() != s.structType {
		return ib
	}

	cols := make([]string, 0, len(fields))
	values := make([]interface{}, 0, len(fields))

	for _, f := range fields {
		name := s.fieldAlias[f]
		data := v.FieldByName(name).Interface()
		cols = append(cols, f)
		values = append(values, data)
	}

	ib.Cols(cols...)
	ib.Values(values...)
	return ib
}

// DeleteFrom creates a new `DeleteBuilder` with table name.
//
// Caller is responsible to set WHERE condition to match right record.
func (s *Struct) DeleteFrom(table string) *DeleteBuilder {
	db := NewDeleteBuilder()
	db.DeleteFrom(table)
	return db
}

// Addr take address of all exported fields of the s from the value.
// The returned result can be used in `Row#Scan` directly.
func (s *Struct) Addr(value interface{}) []interface{} {
	return s.AddrForTag("", value)
}

// AddrForTag take address of all fields of the s tagged with tag from the value.
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

// AddrWithCols take address of all columns defined in cols from the value.
// The returned result can be used in `Row#Scan` directly.
func (s *Struct) AddrWithCols(cols []string, value interface{}) []interface{} {
	v := dereferencedValue(value)

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

func deferencedType(t reflect.Type) reflect.Type {
	for k := t.Kind(); k == reflect.Ptr || k == reflect.Interface; k = t.Kind() {
		t = t.Elem()
	}

	return t
}

func dereferencedValue(value interface{}) reflect.Value {
	v := reflect.ValueOf(value)

	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}

	return v
}
