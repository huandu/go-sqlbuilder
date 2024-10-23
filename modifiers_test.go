// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"github.com/metrumresearchgroup/wrapt"
	"testing"

	"github.com/huandu/go-assert"
)

func TestEscape(t *testing.T) {
	a := assert.New(t)
	cases := map[string]string{
		"foo":  "foo",
		"$foo": "$$foo",
		"$$$":  "$$$$$$",
	}
	var inputs, expects []string

	for s, expected := range cases {
		inputs = append(inputs, s)
		expects = append(expects, expected)
		actual := Escape(s)

		a.Equal(actual, expected)
	}

	actuals := EscapeAll(inputs...)
	a.Equal(actuals, expects)
}

func TestFlatten(t *testing.T) {
	a := assert.New(t)
	cases := [][2]interface{}{
		{
			"foo",
			[]interface{}{"foo"},
		},
		{
			[]int{1, 2, 3},
			[]interface{}{1, 2, 3},
		},
		{
			[]interface{}{"abc", []int{1, 2, 3}, [3]string{"def", "ghi"}},
			[]interface{}{"abc", 1, 2, 3, "def", "ghi", ""},
		},
	}

	for _, c := range cases {
		input, expected := c[0], c[1]
		actual := Flatten(input)

		a.Equal(actual, expected)
	}
}

func TestTuple(t *testing.T) {
	a := assert.New(t)
	cases := []struct {
		values   []interface{}
		expected string
	}{
		{
			nil,
			"()",
		},
		{
			[]interface{}{1, "bar", nil, Tuple("foo", Tuple(2, "baz"))},
			"(1, 'bar', NULL, ('foo', (2, 'baz')))",
		},
	}

	for _, c := range cases {
		sql, args := Build("$?", Tuple(c.values...)).Build()
		actual, err := DefaultFlavor.Interpolate(sql, args)
		a.NilError(err)
		a.Equal(actual, c.expected)
	}
}

func ExampleTuple() {
	sb := Select("id", "name").From("user")
	sb.Where(
		sb.In(
			TupleNames("type", "status"),
			Tuple("web", 1),
			Tuple("app", 1),
			Tuple("app", 2),
		),
	)
	sql, args := sb.Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT id, name FROM user WHERE (type, status) IN ((?, ?), (?, ?), (?, ?))
	// [web 1 app 1 app 2]
}

func TestJSONNames(tt *testing.T) {
	type args struct {
		names []string
	}
	tests := []struct {
		name      string
		names     []string
		assertion func(t *wrapt.T, got string)
	}{
		// TODO: Add test cases.
		{
			name:  "expected with single entry",
			names: []string{"foo"},
			assertion: func(t *wrapt.T, got string) {
				t.A.Equal(`{"foo"}`, got)
			},
		},
		{
			name:  "expected with multiple entries",
			names: []string{"foo", "bar"},
			assertion: func(t *wrapt.T, got string) {
				t.A.Equal(`{"foo", "bar"}`, got)
			},
		},
	}
	for _, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			t := wrapt.WrapT(tt)
			t.R.NotNil(test.assertion)
			got := JSONNames(test.names...)
			test.assertion(t, got)
		})
	}
}
