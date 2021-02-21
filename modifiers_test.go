// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
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
