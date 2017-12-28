// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
)

func ExampleBuildf() {
	sb := NewSelectBuilder()
	sb.Select("id").From("user")

	explain := Buildf("EXPLAIN %v LEFT JOIN SELECT * FROM banned WHERE state = (%v, %v)", sb, 1, 2)
	sql, args := explain.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// EXPLAIN SELECT id FROM user LEFT JOIN SELECT * FROM banned WHERE state = (?, ?)
	// [1 2]
}
