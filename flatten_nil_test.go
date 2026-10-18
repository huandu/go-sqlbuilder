package sqlbuilder

import (
	"reflect"
	"testing"
)

func TestFlattenPreservesNilValues(t *testing.T) {
	for _, tc := range []struct {
		input interface{}
		want  []interface{}
	}{
		{nil, []interface{}{nil}},
		{[]interface{}{1, nil, []interface{}{nil, 2}}, []interface{}{1, nil, nil, 2}},
		{[]interface{}{[]int{}, []int(nil)}, nil},
	} {
		if got := Flatten(tc.input); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Flatten(%#v) = %#v, want %#v", tc.input, got, tc.want)
		}
	}
	sql, args := Build("VALUES ($?)", List([]interface{}{1, nil, 2})).Build()
	if sql != "VALUES (?, ?, ?)" || !reflect.DeepEqual(args, []interface{}{1, nil, 2}) {
		t.Errorf("nil list item lost: %q, %#v", sql, args)
	}
}
