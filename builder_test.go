// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
)

func ExampleBuildf() {
	sb := NewSelectBuilder()
	sb.Select("id").From("user")

	explain := Buildf("EXPLAIN %v LEFT JOIN SELECT * FROM banned WHERE state IN (%v, %v)", sb, 1, 2)
	sql, args := explain.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// EXPLAIN SELECT id FROM user LEFT JOIN SELECT * FROM banned WHERE state IN (?, ?)
	// [1 2]
}

func ExampleBuild() {
	sb := NewSelectBuilder()
	sb.Select("id").From("user").Where(sb.In("status", 1, 2))

	b := Build("EXPLAIN $? LEFT JOIN SELECT * FROM $? WHERE created_at > $? AND state IN (${states}) AND modified_at BETWEEN $2 AND $?",
		sb, Raw("banned"), 1514458225, 1514544625, Named("states", List([]int{3, 4, 5})))
	sql, args := b.Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// EXPLAIN SELECT id FROM user WHERE status IN (?, ?) LEFT JOIN SELECT * FROM banned WHERE created_at > ? AND state IN (?, ?, ?) AND modified_at BETWEEN ? AND ?
	// [1 2 1514458225 3 4 5 1514458225 1514544625]
}
