// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
)

func ExampleCreateTableBuilder() {
	ctb := NewCreateTableBuilder()
	ctb.CreateTable("demo.user").IfNotExists()
	ctb.Define("id", "BIGINT(20)", "NOT NULL", "AUTO_INCREMENT", "PRIMARY KEY", `COMMENT "user id"`)
	ctb.Define("name", "VARCHAR(255)", "NOT NULL", `COMMENT "user name"`)
	ctb.Define("created_at", "DATETIME", "NOT NULL", `COMMENT "user create time"`)
	ctb.Define("modified_at", "DATETIME", "NOT NULL", `COMMENT "user modify time"`)
	ctb.Define("KEY", "idx_name_modified_at", "name, modified_at")
	ctb.Option("DEFAULT CHARACTER SET", "utf8mb4")

	fmt.Println(ctb)

	// Output:
	// CREATE TABLE IF NOT EXISTS demo.user (id BIGINT(20) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT "user id", name VARCHAR(255) NOT NULL COMMENT "user name", created_at DATETIME NOT NULL COMMENT "user create time", modified_at DATETIME NOT NULL COMMENT "user modify time", KEY idx_name_modified_at name, modified_at) DEFAULT CHARACTER SET utf8mb4
}

func ExampleCreateTableBuilder_tempTable() {
	ctb := NewCreateTableBuilder()
	ctb.CreateTempTable("demo.user").IfNotExists()
	ctb.Define("id", "BIGINT(20)", "NOT NULL", "AUTO_INCREMENT", "PRIMARY KEY", `COMMENT "user id"`)
	ctb.Define("name", "VARCHAR(255)", "NOT NULL", `COMMENT "user name"`)
	ctb.Define("created_at", "DATETIME", "NOT NULL", `COMMENT "user create time"`)
	ctb.Define("modified_at", "DATETIME", "NOT NULL", `COMMENT "user modify time"`)
	ctb.Define("KEY", "idx_name_modified_at", "name, modified_at")
	ctb.Option("DEFAULT CHARACTER SET", "utf8mb4")

	fmt.Println(ctb)

	// Output:
	// CREATE TEMPORARY TABLE IF NOT EXISTS demo.user (id BIGINT(20) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT "user id", name VARCHAR(255) NOT NULL COMMENT "user name", created_at DATETIME NOT NULL COMMENT "user create time", modified_at DATETIME NOT NULL COMMENT "user modify time", KEY idx_name_modified_at name, modified_at) DEFAULT CHARACTER SET utf8mb4
}
