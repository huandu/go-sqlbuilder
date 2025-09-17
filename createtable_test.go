// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"

	"github.com/huandu/go-assert"
)

func ExampleCreateTable() {
	sql := CreateTable("demo.user").IfNotExists().
		Define("id", "BIGINT(20)", "NOT NULL", "AUTO_INCREMENT", "PRIMARY KEY", `COMMENT "user id"`).
		String()

	fmt.Println(sql)

	// Output:
	// CREATE TABLE IF NOT EXISTS demo.user (id BIGINT(20) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT "user id")
}

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

func ExampleCreateTableBuilder_SQL() {
	ctb := NewCreateTableBuilder()
	ctb.SQL(`/* before */`)
	ctb.CreateTempTable("demo.user").IfNotExists()
	ctb.SQL("/* after create */")
	ctb.Define("id", "BIGINT(20)", "NOT NULL", "AUTO_INCREMENT", "PRIMARY KEY", `COMMENT "user id"`)
	ctb.Define("name", "VARCHAR(255)", "NOT NULL", `COMMENT "user name"`)
	ctb.SQL("/* after define */")
	ctb.Option("DEFAULT CHARACTER SET", "utf8mb4")
	ctb.SQL(ctb.Var(Build("AS SELECT * FROM old.user WHERE name LIKE $?", "%Huan%")))

	sql, args := ctb.Build()
	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// /* before */ CREATE TEMPORARY TABLE IF NOT EXISTS demo.user /* after create */ (id BIGINT(20) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT "user id", name VARCHAR(255) NOT NULL COMMENT "user name") /* after define */ DEFAULT CHARACTER SET utf8mb4 AS SELECT * FROM old.user WHERE name LIKE ?
	// [%Huan%]
}

func ExampleCreateTableBuilder_NumDefine() {
	ctb := NewCreateTableBuilder()
	ctb.CreateTable("demo.user").IfNotExists()
	ctb.Define("id", "BIGINT(20)", "NOT NULL", "AUTO_INCREMENT", "PRIMARY KEY", `COMMENT "user id"`)
	ctb.Define("name", "VARCHAR(255)", "NOT NULL", `COMMENT "user name"`)
	ctb.Define("created_at", "DATETIME", "NOT NULL", `COMMENT "user create time"`)
	ctb.Define("modified_at", "DATETIME", "NOT NULL", `COMMENT "user modify time"`)
	ctb.Define("KEY", "idx_name_modified_at", "name, modified_at")
	ctb.Option("DEFAULT CHARACTER SET", "utf8mb4")

	// Count the number of definitions.
	fmt.Println(ctb.NumDefine())

	// Output:
	// 5
}

func TestCreateTableGetFlavor(t *testing.T) {
	a := assert.New(t)
	ctb := newCreateTableBuilder()

	ctb.SetFlavor(PostgreSQL)
	flavor := ctb.Flavor()
	a.Equal(PostgreSQL, flavor)

	ctbClick := ClickHouse.NewCreateTableBuilder()
	flavor = ctbClick.Flavor()
	a.Equal(ClickHouse, flavor)
}

func TestCreateTableClone(t *testing.T) {
	a := assert.New(t)
	ctb := CreateTable("demo.user").IfNotExists().
		Define("id", "BIGINT(20)", "NOT NULL", "AUTO_INCREMENT", "PRIMARY KEY", `COMMENT "user id"`).
		Option("DEFAULT CHARACTER SET", "utf8mb4")

	ctb2 := ctb.Clone()
	ctb2.Define("name", "VARCHAR(255)", "NOT NULL", `COMMENT "user name"`)

	sql1, args1 := ctb.Build()
	sql2, args2 := ctb2.Build()

	a.Equal("CREATE TABLE IF NOT EXISTS demo.user (id BIGINT(20) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT \"user id\") DEFAULT CHARACTER SET utf8mb4", sql1)
	a.Equal(0, len(args1))

	a.Equal("CREATE TABLE IF NOT EXISTS demo.user (id BIGINT(20) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT \"user id\", name VARCHAR(255) NOT NULL COMMENT \"user name\") DEFAULT CHARACTER SET utf8mb4", sql2)
	a.Equal(0, len(args2))
}
