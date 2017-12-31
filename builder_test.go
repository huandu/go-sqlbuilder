// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql"
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

func ExampleBuildNamed() {
	b := BuildNamed("SELECT * FROM ${table} WHERE status IN (${status}) AND name LIKE ${name} AND created_at > ${time} AND modified_at < ${time} + 86400",
		map[string]interface{}{
			"time":   sql.Named("start", 1234567890),
			"status": List([]int{1, 2, 5}),
			"name":   "Huan%",
			"table":  Raw("user"),
		})
	sql, args := b.Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT * FROM user WHERE status IN (?, ?, ?) AND name LIKE ? AND created_at > @start AND modified_at < @start + 86400
	// [1 2 5 Huan% {{} start 1234567890}]
}
