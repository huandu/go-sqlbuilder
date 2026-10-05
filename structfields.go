package sqlbuilder

import (
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

var typeOfSQLScanner = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

type structFields struct {
	noTag  *structTaggedFields
	tagged map[string]*structTaggedFields
}

type structTaggedFields struct {
	// All columns for SELECT.
	ForRead     []*structField
	colsForRead map[string]*structField

	// All columns which can be used in INSERT and UPDATE.
	ForWrite     []*structField
	colsForWrite map[string]struct{}

	// All columns which can be used in INSERT.
	ForInsert     []*structField
	colsForInsert map[string]struct{}
}

type structField struct {
	Name     string
	Alias    string
	As       string
	Tags     []string
	IsQuoted bool
	DBTag    string
	Field    reflect.StructField
	Index    []int

	SliceIndex []int
	SliceDepth int

	omitEmptyTags omitEmptyTagMap
}

type structFieldExpandMode uint8

const (
	structFieldExpandDefault structFieldExpandMode = iota
	structFieldExpandEnabled
	structFieldExpandDisabled
)

type structFieldOptions struct {
	isQuoted      bool
	omitEmptyTags omitEmptyTagMap
	expandMode    structFieldExpandMode
}

type structFieldsParser func() *structFields

func makeDefaultFieldsParser(t reflect.Type) structFieldsParser {
	return makeFieldsParser(t, nil, true)
}

func makeCustomFieldsParser(t reflect.Type, mapper FieldMapperFunc) structFieldsParser {
	return makeFieldsParser(t, mapper, false)
}

func makeFieldsParser(t reflect.Type, mapper FieldMapperFunc, useDefault bool) structFieldsParser {
	var once sync.Once
	sfs := &structFields{
		noTag:  makeStructTaggedFields(),
		tagged: map[string]*structTaggedFields{},
	}

	return func() *structFields {
		once.Do(func() {
			if useDefault {
				mapper = DefaultFieldMapper
			}

			sfs.parse(t, mapper, "", nil, true, nil, 0, nil)
		})

		return sfs
	}
}

func (sfs *structFields) parse(t reflect.Type, mapper FieldMapperFunc, prefix string, index []int, allowInsert bool, slicePath []int, sliceDepth int, inheritedTags []string) {
	l := t.NumField()
	var anonymous []reflect.StructField

	for i := 0; i < l; i++ {
		field := t.Field(i)

		// Skip unexported fields that are not embedded structs.
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		if field.Anonymous {
			ft := field.Type

			// If field is an anonymous struct or pointer to struct, parse it later.
			if shouldExpandAnonymousStructField(ft) {
				anonymous = append(anonymous, field)
				continue
			}
		}

		// Parse DBTag.
		alias, dbtag := DefaultGetAlias(&field)

		if alias == "-" {
			continue
		}

		fieldOpts := parseStructFieldOptions(field)

		if expand, targetType, isSlice := shouldExpandTaggedField(field.Type, dbtag, fieldOpts); expand {
			structField := makeStructField(field, alias, dbtag, mapper, fieldOpts, prefix, index, i, slicePath, sliceDepth, inheritedTags)
			if allowInsert {
				sfs.addInsertField(structField)
			}
			newSlicePath := slicePath
			newSliceDepth := sliceDepth
			if isSlice {
				newSlicePath = appendFieldIndex(index, i)
				newSliceDepth++
			}
			var nextTags []string
			if len(structField.Tags) > 0 {
				nextTags = append(nextTags, structField.Tags...)
			}
			if len(inheritedTags) > 0 {
				nextTags = append(nextTags, inheritedTags...)
			}
			if len(nextTags) > 0 {
				sort.Strings(nextTags)
				nextTags = removeDuplicatedTags(nextTags)
			}

			sfs.parse(targetType, mapper, dbtag+".", appendFieldIndex(index, i), false, newSlicePath, newSliceDepth, nextTags)
			continue
		}

		structField := makeStructField(field, alias, dbtag, mapper, fieldOpts, prefix, index, i, slicePath, sliceDepth, inheritedTags)
		if allowInsert {
			sfs.addField(structField)
		} else {
			sfs.addReadWriteField(structField)
		}
	}

	for _, field := range anonymous {
		ft := dereferencedType(field.Type)
		sfs.parse(ft, mapper, prefix, appendFieldIndex(index, field.Index...), allowInsert, slicePath, sliceDepth, inheritedTags)
	}
}

func makeStructField(field reflect.StructField, alias, dbtag string, mapper FieldMapperFunc, fieldOpts structFieldOptions, prefix string, index []int, fieldIndex int, slicePath []int, sliceDepth int, inheritedTags []string) *structField {
	if alias == "" {
		alias = field.Name
		if mapper != nil {
			alias = mapper(alias)
		}
	}

	if prefix != "" && !strings.ContainsRune(alias, '.') {
		alias = prefix + alias
	}

	fieldas := field.Tag.Get(FieldAs)
	fieldtag := field.Tag.Get(FieldTag)
	tags := splitTags(fieldtag)
	if len(inheritedTags) > 0 {
		tags = append(tags, inheritedTags...)
		sort.Strings(tags)
		tags = removeDuplicatedTags(tags)
	}

	return &structField{
		Name:          field.Name,
		Alias:         alias,
		As:            fieldas,
		Tags:          tags,
		IsQuoted:      fieldOpts.isQuoted,
		DBTag:         dbtag,
		Field:         field,
		Index:         appendFieldIndex(index, fieldIndex),
		SliceIndex:    slicePath,
		SliceDepth:    sliceDepth,
		omitEmptyTags: fieldOpts.omitEmptyTags,
	}
}

func parseStructFieldOptions(field reflect.StructField) structFieldOptions {
	fieldopt := field.Tag.Get(FieldOpt)
	opts := optRegex.FindAllString(fieldopt, -1)
	fieldOpts := structFieldOptions{
		omitEmptyTags: omitEmptyTagMap{},
	}

	for _, opt := range opts {
		optMap := getOptMatchedMap(opt)

		switch optMap[optName] {
		case fieldOptOmitEmpty:
			tags := getTagsFromOptParams(optMap[optParams])

			for _, tag := range tags {
				fieldOpts.omitEmptyTags[tag] = struct{}{}
			}

		case fieldOptWithQuote:
			fieldOpts.isQuoted = true

		case fieldOptExpand:
			fieldOpts.expandMode = structFieldExpandEnabled

		case fieldOptNoExpand:
			fieldOpts.expandMode = structFieldExpandDisabled
		}
	}

	return fieldOpts
}

func (sfs *structFields) addField(field *structField) {
	sfs.addReadWriteField(field)
	sfs.addInsertField(field)
}

func (sfs *structFields) addReadWriteField(field *structField) {
	sfs.noTag.AddReadWrite(field)

	for _, tag := range field.Tags {
		sfs.taggedFields(tag).AddReadWrite(field)
	}
}

func (sfs *structFields) addInsertField(field *structField) {
	sfs.noTag.AddInsert(field)

	for _, tag := range field.Tags {
		sfs.taggedFields(tag).AddInsert(field)
	}
}

func appendFieldIndex(prefix []int, index ...int) []int {
	path := make([]int, 0, len(prefix)+len(index))
	path = append(path, prefix...)
	path = append(path, index...)
	return path
}

func shouldExpandAnonymousStructField(t reflect.Type) bool {
	return canExpandStructType(t)
}

func canExpandType(t reflect.Type) (expandable bool, targetType reflect.Type, isSlice bool) {
	if t == nil {
		return false, nil, false
	}

	dt := dereferencedType(t)
	if dt.Kind() == reflect.Slice {
		elem := dereferencedType(dt.Elem())
		if canExpandStructType(elem) {
			return true, elem, true
		}
		return false, nil, false
	}

	if canExpandStructType(dt) {
		return true, dt, false
	}

	return false, nil, false
}

func shouldExpandTaggedField(t reflect.Type, dbtag string, fieldOpts structFieldOptions) (expandable bool, targetType reflect.Type, isSlice bool) {
	if dbtag == "" {
		return false, nil, false
	}

	can, target, slice := canExpandType(t)
	if !can {
		return false, nil, false
	}

	switch fieldOpts.expandMode {
	case structFieldExpandEnabled:
		return true, target, slice

	case structFieldExpandDisabled:
		return false, nil, false
	}

	if NoExpand {
		return false, nil, false
	}

	return true, target, slice
}

func shouldExpandTaggedStructField(t reflect.Type, dbtag string, fieldOpts structFieldOptions) bool {
	expandable, _, _ := shouldExpandTaggedField(t, dbtag, fieldOpts)
	return expandable
}

func canExpandStructType(t reflect.Type) bool {
	if t == nil {
		return false
	}

	dt := dereferencedType(t)
	if dt.Kind() != reflect.Struct {
		return false
	}

	if implementsScannerOrValuer(t) || implementsScannerOrValuer(dt) {
		return false
	}

	if dt != t && implementsScannerOrValuer(reflect.PtrTo(dt)) {
		return false
	}

	for i := 0; i < dt.NumField(); i++ {
		field := dt.Field(i)
		if field.PkgPath == "" || field.Anonymous {
			return true
		}
	}

	return false
}

func implementsScannerOrValuer(t reflect.Type) bool {
	if t == nil {
		return false
	}

	return t.Implements(typeOfSQLDriverValuer) || t.Implements(typeOfSQLScanner)
}

func (sfs *structFields) FilterTags(with, without []string) *structTaggedFields {
	if len(with) == 0 && len(without) == 0 {
		return sfs.noTag
	}

	// Simply return the tagged fields.
	if len(with) == 1 && len(without) == 0 {
		return sfs.tagged[with[0]]
	}

	// Find out all with and without fields.
	taggedFields := makeStructTaggedFields()
	filteredReadFields := make(map[string]struct{}, len(sfs.noTag.colsForRead))
	filteredInsertFields := make(map[string]struct{}, len(sfs.noTag.colsForInsert))

	for _, tag := range without {
		if field, ok := sfs.tagged[tag]; ok {
			for k := range field.colsForRead {
				filteredReadFields[k] = struct{}{}
			}

			for k := range field.colsForInsert {
				filteredInsertFields[k] = struct{}{}
			}
		}
	}

	if len(with) == 0 {
		for _, field := range sfs.noTag.ForRead {
			k := field.Key()

			if _, ok := filteredReadFields[k]; !ok {
				taggedFields.AddReadWrite(field)
			}
		}

		for _, field := range sfs.noTag.ForInsert {
			if _, ok := filteredInsertFields[field.Alias]; !ok {
				taggedFields.AddInsert(field)
			}
		}
	} else {
		for _, tag := range with {
			if fields, ok := sfs.tagged[tag]; ok {
				for _, field := range fields.ForRead {
					k := field.Key()

					if _, ok := filteredReadFields[k]; !ok {
						taggedFields.AddReadWrite(field)
					}
				}

				for _, field := range fields.ForInsert {
					if _, ok := filteredInsertFields[field.Alias]; !ok {
						taggedFields.AddInsert(field)
					}
				}
			}
		}
	}

	return taggedFields
}

func (sfs *structFields) taggedFields(tag string) *structTaggedFields {
	fields, ok := sfs.tagged[tag]

	if !ok {
		fields = makeStructTaggedFields()
		sfs.tagged[tag] = fields
	}

	return fields
}

func makeStructTaggedFields() *structTaggedFields {
	return &structTaggedFields{
		colsForRead:   map[string]*structField{},
		colsForWrite:  map[string]struct{}{},
		colsForInsert: map[string]struct{}{},
	}
}

// Add a new field to stfs.
// If field's key exists in stfs.fields, the field is ignored.
func (stfs *structTaggedFields) Add(field *structField) {
	stfs.AddReadWrite(field)
	stfs.AddInsert(field)
}

func (stfs *structTaggedFields) AddReadWrite(field *structField) {
	key := field.Key()

	if _, ok := stfs.colsForRead[key]; !ok {
		stfs.colsForRead[key] = field
		stfs.ForRead = append(stfs.ForRead, field)
	}

	key = field.Alias

	if _, ok := stfs.colsForWrite[key]; !ok {
		stfs.colsForWrite[key] = struct{}{}
		stfs.ForWrite = append(stfs.ForWrite, field)
	}
}

func (stfs *structTaggedFields) AddInsert(field *structField) {
	key := field.Alias

	if _, ok := stfs.colsForInsert[key]; !ok {
		stfs.colsForInsert[key] = struct{}{}
		stfs.ForInsert = append(stfs.ForInsert, field)
	}
}

// Cols returns the fields whose key is one of cols.
// If any column in cols doesn't exist in sfs.fields, returns nil.
func (stfs *structTaggedFields) Cols(cols []string) []*structField {
	fields := make([]*structField, 0, len(cols))

	for _, col := range cols {
		field := stfs.colsForRead[col]

		if field == nil {
			return nil
		}

		fields = append(fields, field)
	}

	return fields
}

// Key returns the key name to identify a field.
func (sf *structField) Key() string {
	if sf.As != "" {
		return sf.As
	}

	if sf.Alias != "" {
		return sf.Alias
	}

	return sf.Name
}

// NameForSelect returns the name for SELECT.
func (sf *structField) NameForSelect(flavor Flavor) string {
	if sf.As == "" {
		return sf.Quote(flavor)
	}

	return fmt.Sprintf("%s AS %s", sf.Quote(flavor), sf.As)
}

// Quote the Alias in sf with flavor.
func (sf *structField) Quote(flavor Flavor) string {
	if !sf.IsQuoted {
		return sf.Alias
	}

	return flavor.Quote(sf.Alias)
}

// ShouldOmitEmpty returns true only if any one of tags is in the omitted tags map.
func (sf *structField) ShouldOmitEmpty(tags ...string) (ret bool) {
	omit := sf.omitEmptyTags

	if len(omit) == 0 {
		return
	}

	// Always check default tag.
	if _, ret = omit[""]; ret {
		return
	}

	for _, tag := range tags {
		if _, ret = omit[tag]; ret {
			return
		}
	}

	return
}

type omitEmptyTagMap map[string]struct{}

func getOptMatchedMap(opt string) (res map[string]string) {
	res = map[string]string{}
	sm := optRegex.FindStringSubmatch(opt)

	for i, name := range optRegex.SubexpNames() {
		if name != "" {
			res[name] = sm[i]
		}
	}

	return
}

func getTagsFromOptParams(opts string) (tags []string) {
	tags = splitTags(opts)

	if len(tags) == 0 {
		tags = append(tags, "")
	}

	return
}

func splitTags(fieldtag string) (tags []string) {
	parts := strings.Split(fieldtag, ",")

	for _, v := range parts {
		tag := strings.TrimSpace(v)

		if tag == "" {
			continue
		}

		tags = append(tags, tag)
	}

	return
}
