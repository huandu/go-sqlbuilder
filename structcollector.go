// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ScannableRows is an interface matching rows from database/sql.
type ScannableRows interface {
	Next() bool
	Scan(dest ...interface{}) error
}

type rowErrorChecker interface {
	Err() error
}

// StructCollector aggregates flat JOIN rows into grouped parent structs with expanded child slices.
// It holds per-query state and is NOT thread-safe; create one per query via Struct.NewCollector.
type StructCollector struct {
	s          *Struct
	parentKeys []string
	childKeys  map[string][]string
}

// NewCollector creates a stateful per-query StructCollector for scanning flat JOIN rows
// into grouped parent structs. parentKeys specifies the column alias(es) (e.g. "id")
// used to group rows by parent record.
func (s *Struct) NewCollector(parentKeys ...string) *StructCollector {
	return &StructCollector{
		s:          s,
		parentKeys: append([]string(nil), parentKeys...),
		childKeys:  make(map[string][]string),
	}
}

// WithChildKey designates the join/primary key column alias(es) (e.g. "expenses.id")
// for the specified expanded child slice field (e.g. "expenses").
//
// When configured, rows where any designated key column is SQL NULL (such as from a LEFT JOIN)
// are not appended to the child slice, and duplicate child keys within the same parent
// are deduplicated to prevent cartesian row inflation across multiple joined slices.
func (c *StructCollector) WithChildKey(sliceField string, keyCols ...string) *StructCollector {
	if sliceField != "" && len(keyCols) > 0 {
		c.childKeys[sliceField] = append([]string(nil), keyCols...)
	}
	return c
}

// Collect scans all rows from rows into dest, which must be a non-nil pointer to a slice
// of parent structs or struct pointers (e.g. *[]User or *[]*User).
func (c *StructCollector) Collect(rows ScannableRows, dest interface{}) error {
	if c == nil || c.s == nil || c.s.structType == nil {
		return fmt.Errorf("go-sqlbuilder: invalid collector")
	}
	sfs := c.s.structFieldsParser()
	tagged := sfs.FilterTags(c.s.withTags, c.s.withoutTags)
	if tagged == nil {
		return fmt.Errorf("go-sqlbuilder: no fields matched for %s", c.s.structType)
	}
	return c.collectFields(rows, tagged.ForRead, dest)
}

// CollectWithCols scans all rows from rows using the specified column order into dest,
// which must be a non-nil pointer to a slice of parent structs or struct pointers.
func (c *StructCollector) CollectWithCols(rows ScannableRows, cols []string, dest interface{}) error {
	if c == nil || c.s == nil || c.s.structType == nil {
		return fmt.Errorf("go-sqlbuilder: invalid collector")
	}
	sfs := c.s.structFieldsParser()
	tagged := sfs.FilterTags(c.s.withTags, c.s.withoutTags)
	if tagged == nil {
		return fmt.Errorf("go-sqlbuilder: no fields matched for %s", c.s.structType)
	}
	fields := tagged.Cols(cols)
	if fields == nil {
		return fmt.Errorf("go-sqlbuilder: unknown column in cols for %s", c.s.structType)
	}
	return c.collectFields(rows, fields, dest)
}

type collectorSliceMeta struct {
	name          string
	indexKey      string
	sliceIndex    []int
	elemType      reflect.Type
	elemIsPtr     bool
	structType    reflect.Type
	keyColIndices []int
}

type collectorFieldBinding struct {
	sf           *structField
	sliceMetaIdx int
}

func (c *StructCollector) collectFields(rows ScannableRows, fields []*structField, dest interface{}) error {
	if rows == nil {
		return fmt.Errorf("go-sqlbuilder: nil rows")
	}
	if len(c.parentKeys) == 0 {
		return fmt.Errorf("go-sqlbuilder: at least one parent key column is required")
	}
	if dest == nil {
		return fmt.Errorf("go-sqlbuilder: nil destination slice")
	}

	destVal := reflect.ValueOf(dest)
	if !destVal.IsValid() || destVal.Kind() != reflect.Ptr || destVal.IsNil() {
		return fmt.Errorf("go-sqlbuilder: destination must be non-nil pointer to slice, got %T", dest)
	}
	sliceVal := destVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("go-sqlbuilder: destination must be pointer to slice, got %T", dest)
	}

	elemType := sliceVal.Type().Elem()
	elemIsPtr := elemType.Kind() == reflect.Ptr
	elemStructType := dereferencedType(elemType)
	if elemStructType != c.s.structType {
		return fmt.Errorf("go-sqlbuilder: destination slice element type %s does not match %s", elemType, c.s.structType)
	}

	scalarFields := make(map[string]*structField)
	sliceByIndex := make(map[string]int)
	var slices []collectorSliceMeta
	bindings := make([]collectorFieldBinding, len(fields))
	sliceFieldCols := make(map[string]map[string]int)

	for i, sf := range fields {
		if sf.SliceDepth == 0 {
			scalarFields[sf.Key()] = sf
			scalarFields[sf.Alias] = sf
			bindings[i] = collectorFieldBinding{sf: sf, sliceMetaIdx: -1}
			continue
		}

		idxKey := indexPathKey(sf.SliceIndex)
		sIdx, ok := sliceByIndex[idxKey]
		if !ok {
			sIdx = len(slices)
			sliceByIndex[idxKey] = sIdx

			dummyParent := reflect.New(c.s.structType).Elem()
			sfVal, found := fieldByIndex(dummyParent, sf.SliceIndex, true)
			if !found || !sfVal.IsValid() {
				return fmt.Errorf("go-sqlbuilder: cannot resolve slice field %q", sf.SliceName)
			}
			st := dereferencedType(sfVal.Type())
			if st.Kind() != reflect.Slice {
				return fmt.Errorf("go-sqlbuilder: field %q is not a slice", sf.SliceName)
			}
			cElemType := st.Elem()
			slices = append(slices, collectorSliceMeta{
				name:       sf.SliceName,
				indexKey:   idxKey,
				sliceIndex: sf.SliceIndex,
				elemType:   cElemType,
				elemIsPtr:  cElemType.Kind() == reflect.Ptr,
				structType: dereferencedType(cElemType),
			})
			sliceFieldCols[idxKey] = make(map[string]int)
		}

		bindings[i] = collectorFieldBinding{sf: sf, sliceMetaIdx: sIdx}
		sliceFieldCols[idxKey][sf.Key()] = i
		sliceFieldCols[idxKey][sf.Alias] = i
	}

	parentKeyFields := make([]*structField, len(c.parentKeys))
	for i, pk := range c.parentKeys {
		sf, ok := scalarFields[pk]
		if !ok {
			return fmt.Errorf("go-sqlbuilder: parent key column %q not found in selected scalar fields", pk)
		}
		parentKeyFields[i] = sf
	}

	for i := range slices {
		sm := &slices[i]
		keyCols, hasKey := c.childKeys[sm.name]
		if len(slices) > 1 && !hasKey {
			return fmt.Errorf("go-sqlbuilder: child key must be configured via WithChildKey for slice %q when multiple child slices are expanded", sm.name)
		}
		if hasKey {
			sm.keyColIndices = make([]int, len(keyCols))
			for j, kc := range keyCols {
				colIdx, ok := sliceFieldCols[sm.indexKey][kc]
				if !ok {
					return fmt.Errorf("go-sqlbuilder: child key column %q not found in selected fields for slice %q", kc, sm.name)
				}
				sm.keyColIndices[j] = colIdx
			}
		}
	}

	resultSlice := reflect.MakeSlice(sliceVal.Type(), 0, 0)
	parentIndexByKey := make(map[interface{}]int)
	var seenChildKeys []map[string]map[interface{}]struct{}

	for rows.Next() {
		rowParent := reflect.New(c.s.structType).Elem()
		childElems := make([]reflect.Value, len(slices))
		for i, sm := range slices {
			childElems[i] = reflect.New(sm.structType).Elem()
		}

		scanDest := make([]interface{}, len(fields))
		scanners := make([]*nullableFieldScanner, len(fields))

		for i, b := range bindings {
			if b.sliceMetaIdx < 0 {
				fVal, ok := fieldByIndex(rowParent, b.sf.Index, true)
				if !ok || !fVal.IsValid() {
					return fmt.Errorf("go-sqlbuilder: cannot get scan address for %s", b.sf.Alias)
				}
				scanDest[i] = fVal.Addr().Interface()
				continue
			}

			sm := slices[b.sliceMetaIdx]
			relIndex := b.sf.Index[len(sm.sliceIndex):]
			cVal, ok := fieldByIndex(childElems[b.sliceMetaIdx], relIndex, true)
			if !ok || !cVal.IsValid() {
				return fmt.Errorf("go-sqlbuilder: cannot get scan address for %s", b.sf.Alias)
			}
			sc := &nullableFieldScanner{target: cVal}
			scanners[i] = sc
			scanDest[i] = sc
		}

		if err := rows.Scan(scanDest...); err != nil {
			return err
		}

		parentKey, err := extractGroupKey(rowParent, parentKeyFields, nil)
		if err != nil {
			return err
		}

		parentIdx, exists := parentIndexByKey[parentKey]
		if !exists {
			parentIdx = resultSlice.Len()
			if elemIsPtr {
				pPtr := reflect.New(c.s.structType)
				pPtr.Elem().Set(rowParent)
				resultSlice = reflect.Append(resultSlice, pPtr)
			} else {
				resultSlice = reflect.Append(resultSlice, rowParent)
			}
			parentIndexByKey[parentKey] = parentIdx
			seenChildKeys = append(seenChildKeys, make(map[string]map[interface{}]struct{}))
		}

		targetParent := dereferencedValue(resultSlice.Index(parentIdx))

		for sIdx, sm := range slices {
			if isChildRowNull(sIdx, sm, bindings, scanners) {
				continue
			}

			if len(sm.keyColIndices) > 0 {
				keyFields := make([]*structField, len(sm.keyColIndices))
				for j, colIdx := range sm.keyColIndices {
					keyFields[j] = bindings[colIdx].sf
				}
				childKey, err := extractGroupKey(childElems[sIdx], keyFields, sm.sliceIndex)
				if err != nil {
					return err
				}
				seenMap := seenChildKeys[parentIdx][sm.indexKey]
				if seenMap == nil {
					seenMap = make(map[interface{}]struct{})
					seenChildKeys[parentIdx][sm.indexKey] = seenMap
				}
				if _, dup := seenMap[childKey]; dup {
					continue
				}
				seenMap[childKey] = struct{}{}
			}

			if err := appendChildElement(targetParent, sm, childElems[sIdx]); err != nil {
				return err
			}
		}
	}

	if checker, ok := rows.(rowErrorChecker); ok {
		if err := checker.Err(); err != nil {
			return err
		}
	}

	sliceVal.Set(resultSlice)
	return nil
}

func isChildRowNull(sIdx int, sm collectorSliceMeta, bindings []collectorFieldBinding, scanners []*nullableFieldScanner) bool {
	if len(sm.keyColIndices) > 0 {
		for _, colIdx := range sm.keyColIndices {
			if scanners[colIdx] != nil && scanners[colIdx].wasNull {
				return true
			}
		}
		return false
	}

	for i, b := range bindings {
		if b.sliceMetaIdx == sIdx && scanners[i] != nil && !scanners[i].wasNull {
			return false
		}
	}
	return true
}

func appendChildElement(parent reflect.Value, sm collectorSliceMeta, childElem reflect.Value) error {
	sliceField, ok := fieldByIndex(parent, sm.sliceIndex, true)
	if !ok || !sliceField.IsValid() {
		return fmt.Errorf("go-sqlbuilder: cannot access slice field %q", sm.name)
	}

	for sliceField.Kind() == reflect.Ptr {
		if sliceField.IsNil() {
			sliceField.Set(reflect.New(sliceField.Type().Elem()))
		}
		sliceField = sliceField.Elem()
	}

	if sm.elemIsPtr {
		ptr := reflect.New(sm.structType)
		ptr.Elem().Set(childElem)
		sliceField.Set(reflect.Append(sliceField, ptr))
	} else {
		sliceField.Set(reflect.Append(sliceField, childElem))
	}
	return nil
}

func extractGroupKey(v reflect.Value, keyFields []*structField, prefixIndex []int) (interface{}, error) {
	if len(keyFields) == 1 {
		relIdx := keyFields[0].Index[len(prefixIndex):]
		fVal, ok := fieldByIndex(v, relIdx, false)
		if !ok || !fVal.IsValid() {
			return nil, fmt.Errorf("go-sqlbuilder: cannot read key column %q", keyFields[0].Alias)
		}
		fVal = dereferencedValue(fVal)
		if !fVal.IsValid() {
			return nil, nil
		}
		if fVal.Type().Comparable() {
			return fVal.Interface(), nil
		}
		return fmt.Sprintf("%T:%v", fVal.Interface(), fVal.Interface()), nil
	}

	var buf strings.Builder
	for i, kf := range keyFields {
		if i > 0 {
			buf.WriteByte('|')
		}
		relIdx := kf.Index[len(prefixIndex):]
		fVal, ok := fieldByIndex(v, relIdx, false)
		if !ok || !fVal.IsValid() {
			return nil, fmt.Errorf("go-sqlbuilder: cannot read key column %q", kf.Alias)
		}
		fVal = dereferencedValue(fVal)
		if !fVal.IsValid() {
			buf.WriteString("<nil>")
			continue
		}
		fmt.Fprintf(&buf, "%T:%v", fVal.Interface(), fVal.Interface())
	}
	return buf.String(), nil
}

func indexPathKey(index []int) string {
	if len(index) == 0 {
		return ""
	}
	var buf strings.Builder
	for i, n := range index {
		if i > 0 {
			buf.WriteByte('.')
		}
		buf.WriteString(strconv.Itoa(n))
	}
	return buf.String()
}

type nullableFieldScanner struct {
	target  reflect.Value
	wasNull bool
}

func (n *nullableFieldScanner) Scan(src interface{}) error {
	if src == nil {
		n.wasNull = true
		if n.target.CanSet() {
			if n.target.CanAddr() && n.target.Addr().Type().Implements(typeOfSQLScanner) {
				return n.target.Addr().Interface().(sql.Scanner).Scan(nil)
			}
			n.target.Set(reflect.Zero(n.target.Type()))
		}
		return nil
	}
	n.wasNull = false
	return assignScannedValue(n.target, src)
}

func assignScannedValue(dst reflect.Value, src interface{}) error {
	if dst.CanAddr() && dst.Addr().Type().Implements(typeOfSQLScanner) {
		return dst.Addr().Interface().(sql.Scanner).Scan(src)
	}

	if dst.Kind() == reflect.Ptr {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return assignScannedValue(dst.Elem(), src)
	}

	srcVal := reflect.ValueOf(src)
	if srcVal.Type().AssignableTo(dst.Type()) {
		dst.Set(srcVal)
		return nil
	}

	switch dst.Kind() {
	case reflect.String:
		switch s := src.(type) {
		case []byte:
			dst.SetString(string(s))
			return nil
		case string:
			dst.SetString(s)
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := src.(type) {
		case int64:
			if dst.OverflowInt(v) {
				return fmt.Errorf("go-sqlbuilder: value %d overflows %s", v, dst.Type())
			}
			dst.SetInt(v)
			return nil
		case []byte:
			n, err := strconv.ParseInt(string(v), 10, dst.Type().Bits())
			if err != nil {
				return err
			}
			dst.SetInt(n)
			return nil
		case string:
			n, err := strconv.ParseInt(v, 10, dst.Type().Bits())
			if err != nil {
				return err
			}
			dst.SetInt(n)
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		switch v := src.(type) {
		case int64:
			if v < 0 {
				return fmt.Errorf("go-sqlbuilder: negative value %d into %s", v, dst.Type())
			}
			if dst.OverflowUint(uint64(v)) {
				return fmt.Errorf("go-sqlbuilder: value %d overflows %s", v, dst.Type())
			}
			dst.SetUint(uint64(v))
			return nil
		case uint64:
			if dst.OverflowUint(v) {
				return fmt.Errorf("go-sqlbuilder: value %d overflows %s", v, dst.Type())
			}
			dst.SetUint(v)
			return nil
		case []byte:
			n, err := strconv.ParseUint(string(v), 10, dst.Type().Bits())
			if err != nil {
				return err
			}
			dst.SetUint(n)
			return nil
		case string:
			n, err := strconv.ParseUint(v, 10, dst.Type().Bits())
			if err != nil {
				return err
			}
			dst.SetUint(n)
			return nil
		}
	case reflect.Float32, reflect.Float64:
		switch v := src.(type) {
		case float64:
			dst.SetFloat(v)
			return nil
		case int64:
			dst.SetFloat(float64(v))
			return nil
		case []byte:
			f, err := strconv.ParseFloat(string(v), dst.Type().Bits())
			if err != nil {
				return err
			}
			dst.SetFloat(f)
			return nil
		case string:
			f, err := strconv.ParseFloat(v, dst.Type().Bits())
			if err != nil {
				return err
			}
			dst.SetFloat(f)
			return nil
		}
	case reflect.Bool:
		switch v := src.(type) {
		case bool:
			dst.SetBool(v)
			return nil
		case int64:
			dst.SetBool(v != 0)
			return nil
		case []byte:
			b, err := strconv.ParseBool(string(v))
			if err != nil {
				return err
			}
			dst.SetBool(b)
			return nil
		case string:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return err
			}
			dst.SetBool(b)
			return nil
		}
	case reflect.Slice:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			switch v := src.(type) {
			case []byte:
				b := make([]byte, len(v))
				copy(b, v)
				dst.SetBytes(b)
				return nil
			case string:
				dst.SetBytes([]byte(v))
				return nil
			}
		}
	}

	if srcVal.Type().ConvertibleTo(dst.Type()) {
		dst.Set(srcVal.Convert(dst.Type()))
		return nil
	}

	return fmt.Errorf("go-sqlbuilder: cannot scan %T into %s", src, dst.Type())
}
