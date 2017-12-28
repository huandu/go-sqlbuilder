// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
)

func ExampleInsertBuilder() {
	ib := NewInsertBuilder()
	ib.InsertInto("demo.user")
	ib.Cols("id", "name", "status", "created_at")
	ib.Values(1, "Huan Du", 1, Raw("UNIX_TIMESTAMP(NOW())"))
	ib.Values(2, "Charmy Liu", 1, 1234567890)

	sql, args := ib.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO demo.user (id, name, status, created_at) VALUES (?, ?, ?, UNIX_TIMESTAMP(NOW())), (?, ?, ?, ?)
	// [1 Huan Du 1 2 Charmy Liu 1 1234567890]
}
