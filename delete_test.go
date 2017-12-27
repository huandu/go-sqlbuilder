// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
)

func ExampleDeleteBuilder() {
	db := NewDeleteBuilder()
	db.DeleteFrom("demo.user")
	db.Where(
		db.GreaterThan("id", 1234),
		db.Like("name", "%Du"),
		db.Or(
			db.IsNull("id_card"),
			db.In("status", 1, 2, 5),
		),
		"modified_at > created_at + "+db.Var(86400), // It's allowed to write arbitrary SQL.
	)

	sql, args := db.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// DELETE FROM demo.user WHERE id > ? AND name LIKE ? AND (id_card IS NULL OR status IN (?, ?, ?)) AND modified_at > created_at + ?
	// [1234 %Du 1 2 5 86400]
}
