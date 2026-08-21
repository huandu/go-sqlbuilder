// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"math"
	"testing"
)

// TestClickHouseFloatLiteral verifies that ClickHouse interpolation keeps
// float arguments typed as floats: whole numbers gain a ".0" suffix so the
// server does not infer an integer and silently narrow a Float column, and
// non-finite values are rendered in the lowercase form ClickHouse's parser
// accepts. Other flavors keep the previous rendering.
func TestClickHouseFloatLiteral(t *testing.T) {
	cases := []struct {
		name   string
		flavor Flavor
		arg    interface{}
		want   string
	}{
		{"float64 whole", ClickHouse, float64(1), "SELECT 1.0"},
		{"float32 whole", ClickHouse, float32(2), "SELECT 2.0"},
		{"nan lowercase", ClickHouse, math.NaN(), "SELECT nan"},
		{"positive inf lowercase", ClickHouse, math.Inf(1), "SELECT inf"},
		{"negative inf lowercase", ClickHouse, math.Inf(-1), "SELECT -inf"},
		{"fractional untouched", ClickHouse, float64(1.5), "SELECT 1.5"},
		{"mysql whole unchanged", MySQL, float64(1), "SELECT 1"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := c.flavor.Interpolate("SELECT ?", []interface{}{c.arg})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}
