// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"reflect"
	"testing"
)

type structUserForTest struct {
	ID        int    `db:"id" fieldtag:"important"`
	Name      string `fieldtag:"important"`
	Status    int    `db:"status" fieldtag:"important"`
	CreatedAt int    `db:"created_at"`
}

var userForTest = NewStruct(new(structUserForTest))

func TestStructSelect(t *testing.T) {
	sb := userForTest.Select("user")
	sql, args := sb.Build()

	if expected := "SELECT id, Name, status, created_at FROM user LIMIT 1"; expected != sql {
		t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
	}

	if len(args) != 0 {
		t.Fatalf("args must be empty. [args:%v]", args)
	}
}

func TestStructSelectForTag(t *testing.T) {
	sb := userForTest.SelectForTag("user", "important")
	sql, args := sb.Build()

	if expected := "SELECT id, Name, status FROM user LIMIT 1"; expected != sql {
		t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
	}

	if len(args) != 0 {
		t.Fatalf("args must be empty. [args:%v]", args)
	}
}

func TestStructUpdate(t *testing.T) {
	user := &structUserForTest{
		ID:        123,
		Name:      "Huan Du",
		Status:    2,
		CreatedAt: 1234567890,
	}
	ub := userForTest.Update("user", user)
	sql, args := ub.Build()

	if expected := "UPDATE user SET id = ?, Name = ?, status = ?, created_at = ?"; expected != sql {
		t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
	}

	if expected := []interface{}{123, "Huan Du", 2, 1234567890}; !reflect.DeepEqual(expected, args) {
		t.Fatalf("invalid args. [expected:%v] [actual:%v]", expected, args)
	}
}

func TestStructUpdateForTag(t *testing.T) {
	user := &structUserForTest{
		ID:        123,
		Name:      "Huan Du",
		Status:    2,
		CreatedAt: 1234567890,
	}
	ub := userForTest.UpdateForTag("user", "important", user)
	sql, args := ub.Build()

	if expected := "UPDATE user SET id = ?, Name = ?, status = ?"; expected != sql {
		t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
	}

	if expected := []interface{}{123, "Huan Du", 2}; !reflect.DeepEqual(expected, args) {
		t.Fatalf("invalid args. [expected:%v] [actual:%v]", expected, args)
	}
}

func TestStructInsertInto(t *testing.T) {
	user := &structUserForTest{
		ID:        123,
		Name:      "Huan Du",
		Status:    2,
		CreatedAt: 1234567890,
	}
	ib := userForTest.InsertInto("user", user)
	sql, args := ib.Build()

	if expected := "INSERT INTO user (id, Name, status, created_at) VALUES (?, ?, ?, ?)"; expected != sql {
		t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
	}

	if expected := []interface{}{123, "Huan Du", 2, 1234567890}; !reflect.DeepEqual(expected, args) {
		t.Fatalf("invalid args. [expected:%v] [actual:%v]", expected, args)
	}
}

func TestStructInsertIntoForTag(t *testing.T) {
	user := &structUserForTest{
		ID:        123,
		Name:      "Huan Du",
		Status:    2,
		CreatedAt: 1234567890,
	}
	ib := userForTest.InsertIntoForTag("user", "important", user)
	sql, args := ib.Build()

	if expected := "INSERT INTO user (id, Name, status) VALUES (?, ?, ?)"; expected != sql {
		t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
	}

	if expected := []interface{}{123, "Huan Du", 2}; !reflect.DeepEqual(expected, args) {
		t.Fatalf("invalid args. [expected:%v] [actual:%v]", expected, args)
	}
}

func TestStructDeleteFrom(t *testing.T) {
	db := userForTest.DeleteFrom("user")
	sql, args := db.Build()

	if expected := "DELETE FROM user"; expected != sql {
		t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
	}

	if len(args) != 0 {
		t.Fatalf("args must be empty. [args:%v]", args)
	}
}

func TestStructAddr(t *testing.T) {
	user := new(structUserForTest)
	expected := &structUserForTest{
		ID:        123,
		Name:      "huandu",
		Status:    2,
		CreatedAt: 1234567890,
	}
	str := fmt.Sprintf("%v %v %v %v", expected.ID, expected.Name, expected.Status, expected.CreatedAt)
	fmt.Sscanf(str, "%d%s%d%d", userForTest.Addr(user)...)

	if !reflect.DeepEqual(expected, user) {
		t.Fatalf("invalid user. [expected:%v] [actual:%v]", expected, user)
	}
}

func TestStructAddrForTag(t *testing.T) {
	user := new(structUserForTest)
	expected := &structUserForTest{
		ID:        123,
		Name:      "huandu",
		Status:    2,
		CreatedAt: 1234567890,
	}
	user.CreatedAt = 9876543210
	str := fmt.Sprintf("%v %v %v %v", expected.ID, expected.Name, expected.Status, expected.CreatedAt)
	fmt.Sscanf(str, "%d%s%d%d", userForTest.AddrForTag("important", user)...)
	expected.CreatedAt = 9876543210

	if !reflect.DeepEqual(expected, user) {
		t.Fatalf("invalid user. [expected:%v] [actual:%v]", expected, user)
	}
}

func TestStructAddrWithCols(t *testing.T) {
	user := new(structUserForTest)
	expected := &structUserForTest{
		ID:        123,
		Name:      "huandu",
		Status:    2,
		CreatedAt: 1234567890,
	}
	str := fmt.Sprintf("%v %v %v %v", expected.Name, expected.ID, expected.CreatedAt, expected.Status)
	fmt.Sscanf(str, "%s%d%d%d", userForTest.AddrWithCols([]string{"Name", "id", "created_at", "status"}, user)...)

	if !reflect.DeepEqual(expected, user) {
		t.Fatalf("invalid user. [expected:%v] [actual:%v]", expected, user)
	}
}

type User struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Status int    `db:"status"`
}

type testDB int
type testRows int

func (db testDB) Query(query string, args ...interface{}) (testRows, error) {
	return 0, nil
}

func (db testDB) Exec(query string, args ...interface{}) {
	return
}

func (rows testRows) Close() error {
	return nil
}

func (rows testRows) Scan(dest ...interface{}) error {
	fmt.Sscan("1234 huandu 1", dest...)
	return nil
}

var userStruct = NewStruct(new(User))
var db testDB

func ExampleStruct_buildSELECTAndUseItAsORM() {
	// Suppose we defined following type and global variable.
	//
	//     type User struct {
	//         ID     int64  `db:"id"`
	//         Name   string `db:"name"`
	//         Status int    `db:"status"`
	//     }
	//
	//     var userStruct = NewStruct(new(User))

	// Prepare SELECT query.
	sb := userStruct.Select("user")
	sb.Where(sb.E("id", 1234))

	// Execute the query.
	sql, args := sb.Build()
	rows, _ := db.Query(sql, args...)
	defer rows.Close()

	// Scan row data to user.
	var user User
	rows.Scan(userStruct.Addr(&user)...)

	fmt.Println(sql)
	fmt.Println(args)
	fmt.Printf("%#v", user)

	// Output:
	// SELECT id, name, status FROM user WHERE id = ? LIMIT 1
	// [1234]
	// sqlbuilder.User{ID:1234, Name:"huandu", Status:1}
}

func ExampleStruct_buildUPDATE() {
	// Suppose we defined following type and global variable.
	//
	//     type User struct {
	//         ID     int64  `db:"id"`
	//         Name   string `db:"name"`
	//         Status int    `db:"status"`
	//     }
	//
	//     var userStruct = NewStruct(new(User))

	// Prepare UPDATE query.
	user := &User{
		ID:     1234,
		Name:   "Huan Du",
		Status: 1,
	}
	ub := userStruct.Update("user", user)
	ub.Where(ub.E("id", user.ID))

	// Execute the query.
	sql, args := ub.Build()
	db.Exec(sql, args...)

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// UPDATE user SET id = ?, name = ?, status = ? WHERE id = ?
	// [1234 Huan Du 1 1234]
}

func ExampleStruct_buildINSERT() {
	// Suppose we defined following type and global variable.
	//
	//     type User struct {
	//         ID     int64  `db:"id"`
	//         Name   string `db:"name"`
	//         Status int    `db:"status"`
	//     }
	//
	//     var userStruct = NewStruct(new(User))

	// Prepare INSERT query.
	user := &User{
		ID:     1234,
		Name:   "Huan Du",
		Status: 1,
	}
	ib := userStruct.InsertInto("user", user)

	// Execute the query.
	sql, args := ib.Build()
	db.Exec(sql, args...)

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO user (id, name, status) VALUES (?, ?, ?)
	// [1234 Huan Du 1]
}

func ExampleStruct_buildDELETE() {
	// Suppose we defined following type and global variable.
	//
	//     type User struct {
	//         ID     int64  `db:"id"`
	//         Name   string `db:"name"`
	//         Status int    `db:"status"`
	//     }
	//
	//     var userStruct = NewStruct(new(User))

	// Prepare DELETE query.
	user := &User{
		ID:     1234,
		Name:   "Huan Du",
		Status: 1,
	}
	b := userStruct.DeleteFrom("user")
	b.Where(b.E("id", user.ID))

	// Execute the query.
	sql, args := b.Build()
	db.Exec(sql, args...)

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// DELETE FROM user WHERE id = ?
	// [1234]
}
