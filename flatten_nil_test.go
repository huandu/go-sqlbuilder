package sqlbuilder

import (
	"testing"

	"github.com/huandu/go-assert"
)

func TestFlattenPreservesNilValues(t *testing.T) {
	a := assert.New(t)
	for _, tc := range []struct {
		input interface{}
		want  []interface{}
	}{
		{nil, []interface{}{nil}},
		{[]interface{}{1, nil, []interface{}{nil, 2}}, []interface{}{1, nil, nil, 2}},
		{[]interface{}{[]int{}, []int(nil)}, nil},
	} {
		a.Equal(Flatten(tc.input), tc.want)
	}
	sql, args := Build("VALUES ($?)", List([]interface{}{1, nil, 2})).Build()
	a.Equal(sql, "VALUES (?, ?, ?)")
	a.Equal(args, []interface{}{1, nil, 2})
}
