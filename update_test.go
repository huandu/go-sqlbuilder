// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
)

func ExampleUpdateBuilder() {
	ub := NewUpdateBuilder()
	ub.Update("demo.user")
	ub.Set(
		ub.Assign("type", "sys"),
		ub.Incr("credit"),
		"modified_at = UNIX_TIMESTAMP(NOW())", // It's allowed to write arbitrary SQL.
	)
	ub.Where(
		ub.GreaterThan("id", 1234),
		ub.Like("name", "%Du"),
		ub.Or(
			ub.IsNull("id_card"),
			ub.In("status", 1, 2, 5),
		),
		"modified_at > created_at + "+ub.Var(86400), // It's allowed to write arbitrary SQL.
	)
	sql, args := ub.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// UPDATE demo.user SET type = ?, credit = credit + 1, modified_at = UNIX_TIMESTAMP(NOW()) WHERE id > ? AND name LIKE ? AND (id_card IS NULL OR status IN (?, ?, ?)) AND modified_at > created_at + ?
	// [sys 1234 %Du 1 2 5 86400]
}
