// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type structUserForTest struct {
	ID        int    `db:"id" fieldtag:"important"`
	Name      string `fieldtag:"important"`
	Status    int    `db:"status" fieldtag:"important"`
	CreatedAt int    `db:"created_at"`
}

var userForTest = NewStruct(new(structUserForTest))

func TestStructSelectFrom(t *testing.T) {
	sb := userForTest.SelectFrom("user")
	sql, args := sb.Build()

	if expected := "SELECT user.id, user.Name, user.status, user.created_at FROM user"; expected != sql {
		t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
	}

	if len(args) != 0 {
		t.Fatalf("args must be empty. [args:%v]", args)
	}
}

func TestStructSelectFromForTag(t *testing.T) {
	sb := userForTest.SelectFromForTag("user", "important")
	sql, args := sb.Build()

	if expected := "SELECT user.id, user.Name, user.status FROM user"; expected != sql {
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

	user2 := &structUserForTest{
		ID:        456,
		Name:      "Du Huan",
		Status:    2,
		CreatedAt: 1234567890,
	}

	fakeUser := struct {
		ID int
	}{789}

	users := []interface{}{user, user2, &fakeUser}

	testInsert := map[*InsertBuilder]string{
		userForTest.InsertInto("user", user):       "INSERT ",
		userForTest.InsertIgnoreInto("user", user): "INSERT IGNORE ",
		userForTest.ReplaceInto("user", user):      "REPLACE ",
	}

	testMulitInsert := map[*InsertBuilder]string{
		userForTest.InsertInto("user", users...):       "INSERT ",
		userForTest.InsertIgnoreInto("user", users...): "INSERT IGNORE ",
		userForTest.ReplaceInto("user", users...):      "REPLACE ",
	}

	for ib, exceptedVerb := range testInsert {
		sql, args := ib.Build()

		if expected := exceptedVerb + "INTO user (id, Name, status, created_at) VALUES (?, ?, ?, ?)"; expected != sql {
			t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
		}

		if expected := []interface{}{123, "Huan Du", 2, 1234567890}; !reflect.DeepEqual(expected, args) {
			t.Fatalf("invalid args. [expected:%v] [actual:%v]", expected, args)
		}
	}

	for ib, exceptedVerb := range testMulitInsert {
		sql, args := ib.Build()

		if expected := exceptedVerb + "INTO user (id, Name, status, created_at) VALUES (?, ?, ?, ?), (?, ?, ?, ?)"; expected != sql {
			t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
		}

		if expected := []interface{}{123, "Huan Du", 2, 1234567890, 456, "Du Huan", 2, 1234567890}; !reflect.DeepEqual(expected, args) {
			t.Fatalf("invalid args. [expected:%v] [actual:%v]", expected, args)
		}
	}

}

func TestStructInsertIntoForTag(t *testing.T) {
	user := &structUserForTest{
		ID:        123,
		Name:      "Huan Du",
		Status:    2,
		CreatedAt: 1234567890,
	}

	user2 := &structUserForTest{
		ID:        456,
		Name:      "Du Huan",
		Status:    2,
		CreatedAt: 1234567890,
	}

	fakeUser := struct {
		ID int
	}{789}

	users := []interface{}{user, user2, &fakeUser}

	testInsertForTag := map[*InsertBuilder]string{
		userForTest.InsertIntoForTag("user", "important", user):       "INSERT ",
		userForTest.InsertIgnoreIntoForTag("user", "important", user): "INSERT IGNORE ",
		userForTest.ReplaceIntoForTag("user", "important", user):      "REPLACE ",
	}

	testMulitInsertForTag := map[*InsertBuilder]string{
		userForTest.InsertIntoForTag("user", "important", users...):       "INSERT ",
		userForTest.InsertIgnoreIntoForTag("user", "important", users...): "INSERT IGNORE ",
		userForTest.ReplaceIntoForTag("user", "important", users...):      "REPLACE ",
	}

	for ib, exceptedVerb := range testInsertForTag {
		sql, args := ib.Build()

		if expected := exceptedVerb + "INTO user (id, Name, status) VALUES (?, ?, ?)"; expected != sql {
			t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
		}

		if expected := []interface{}{123, "Huan Du", 2}; !reflect.DeepEqual(expected, args) {
			t.Fatalf("invalid args. [expected:%v] [actual:%v]", expected, args)
		}
	}

	for ib, exceptedVerb := range testMulitInsertForTag {
		sql, args := ib.Build()

		if expected := exceptedVerb + "INTO user (id, Name, status) VALUES (?, ?, ?), (?, ?, ?)"; expected != sql {
			t.Fatalf("invalid SQL. [expected:%v] [actual:%v]", expected, sql)
		}

		if expected := []interface{}{123, "Huan Du", 2, 456, "Du Huan", 2}; !reflect.DeepEqual(expected, args) {
			t.Fatalf("invalid args. [expected:%v] [actual:%v]", expected, args)
		}
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

type Order struct {
	ID         int64  `db:"id" fieldtag:"new"`
	State      State  `db:"state" fieldtag:"new,paid,done"`
	SkuID      int64  `db:"sku_id" fieldtag:"new"`
	UserID     int64  `db:"user_id" fieldtag:"new"`
	Price      int64  `db:"price" fieldtag:"new,update"`
	Discount   int64  `db:"discount" fieldtag:"new,update"`
	Desc       string `db:"desc" fieldtag:"new,update" fieldopt:"withquote"`
	CreatedAt  int64  `db:"created_at" fieldtag:"new"`
	ModifiedAt int64  `db:"modified_at" fieldtag:"new,update,paid,done"`
}

type State int
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
var orderStruct = NewStruct(new(Order))
var db testDB

const (
	OrderStateInvalid State = iota
	OrderStateCreated
	OrderStatePaid
)

func ExampleStruct_useStructAsORM() {
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
	sb := userStruct.SelectFrom("user")
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
	// SELECT user.id, user.name, user.status FROM user WHERE id = ?
	// [1234]
	// sqlbuilder.User{ID:1234, Name:"huandu", Status:1}
}

func ExampleStruct_useTag() {
	// Suppose we defined following type and global variable.
	//
	//     type Order struct {
	//         ID         int64  `db:"id" fieldtag:"update,paid"`
	//         State      int    `db:"state" fieldtag:"paid"`
	//         SkuID      int64  `db:"sku_id"`
	//         UserID     int64  `db:"user_id"`
	//         Price      int64  `db:"price" fieldtag:"update"`
	//         Discount   int64  `db:"discount" fieldtag:"update"`
	//         Desc       string `db:"desc" fieldtag:"update" fieldopt:"withquote"` // `desc` is a keyword.
	//         CreatedAt  int64  `db:"created_at"`
	//         ModifiedAt int64  `db:"modified_at" fieldtag:"update,paid"`
	//     }
	//
	//     var orderStruct = NewStruct(new(Order))

	createOrder := func(table string) {
		now := time.Now().Unix()
		order := &Order{
			ID:         1234,
			State:      OrderStateCreated,
			SkuID:      5678,
			UserID:     7527,
			Price:      1000,
			Discount:   0,
			Desc:       "Best goods",
			CreatedAt:  now,
			ModifiedAt: now,
		}
		b := orderStruct.InsertInto(table, &order)
		sql, args := b.Build()
		db.Exec(sql, args)
		fmt.Println(sql)
	}
	updatePrice := func(table string) {
		tag := "update"

		// Read order from database.
		var order Order
		sql, args := orderStruct.SelectFromForTag(table, tag).Where("id = 1234").Build()
		rows, _ := db.Query(sql, args...)
		defer rows.Close()
		rows.Scan(orderStruct.AddrForTag(tag, &order)...)

		// Discount for this user.
		// Use tag "update" to update necessary columns only.
		order.Discount += 100
		order.ModifiedAt = time.Now().Unix()

		// Save the order.
		b := orderStruct.UpdateForTag(table, tag, &order)
		b.Where(b.E("id", order.ID))
		sql, args = b.Build()
		db.Exec(sql, args...)
		fmt.Println(sql)
	}
	updateState := func(table string) {
		tag := "paid"

		// Read order from database.
		var order Order
		sql, args := orderStruct.SelectFromForTag(table, tag).Where("id = 1234").Build()
		rows, _ := db.Query(sql, args...)
		defer rows.Close()
		rows.Scan(orderStruct.AddrForTag(tag, &order)...)

		// Update state to paid when user has paid for the order.
		// Use tag "paid" to update necessary columns only.
		if order.State != OrderStateCreated {
			// Report state error here.
			return
		}

		// Update order state.
		order.State = OrderStatePaid
		order.ModifiedAt = time.Now().Unix()

		// Save the order.
		b := orderStruct.UpdateForTag(table, tag, &order)
		b.Where(b.E("id", order.ID))
		sql, args = b.Build()
		db.Exec(sql, args...)
		fmt.Println(sql)
	}

	table := "order"
	createOrder(table)
	updatePrice(table)
	updateState(table)

	fmt.Println("done")

	// Output:
	// INSERT INTO order (id, state, sku_id, user_id, price, discount, `desc`, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	// UPDATE order SET price = ?, discount = ?, `desc` = ?, modified_at = ? WHERE id = ?
	// done
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

func ExampleStruct_forPostgreSQL() {
	userStruct := NewStruct(new(User)).For(PostgreSQL)

	sb := userStruct.SelectFrom("user")
	sb.Where(sb.E("id", 1234))
	sql, args := sb.Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT user.id, user.name, user.status FROM user WHERE id = $1
	// [1234]
}

type structWithQuote struct {
	A string  `db:"aa" fieldopt:"withquote"`
	B int     `db:"-" fieldopt:"withquote"` // fieldopt is ignored as db is "-".
	C float64 `db:"ccc"`
}

func TestStructWithQuote(t *testing.T) {
	sb := NewStruct(new(structWithQuote)).For(MySQL).SelectFrom("foo")
	sql, _ := sb.Build()

	if expected := "SELECT foo.`aa`, foo.ccc FROM foo"; sql != expected {
		t.Fatalf("invalid sql. [expected:%v] [actual:%v]", expected, sql)
	}

	sb = NewStruct(new(structWithQuote)).For(PostgreSQL).SelectFrom("foo")
	sql, _ = sb.Build()

	if expected := `SELECT foo."aa", foo.ccc FROM foo`; sql != expected {
		t.Fatalf("invalid sql. [expected:%v] [actual:%v]", expected, sql)
	}

	ub := NewStruct(new(structWithQuote)).For(MySQL).Update("foo", structWithQuote{A: "aaa"})
	sql, _ = ub.Build()

	if expected := "UPDATE foo SET `aa` = ?, ccc = ?"; sql != expected {
		t.Fatalf("invalid sql. [expected:%v] [actual:%v]", expected, sql)
	}

	ub = NewStruct(new(structWithQuote)).For(PostgreSQL).Update("foo", structWithQuote{A: "aaa"})
	sql, _ = ub.Build()

	if expected := `UPDATE foo SET "aa" = $1, ccc = $2`; sql != expected {
		t.Fatalf("invalid sql. [expected:%v] [actual:%v]", expected, sql)
	}

	ib := NewStruct(new(structWithQuote)).For(MySQL).InsertInto("foo", structWithQuote{A: "aaa"})
	sql, _ = ib.Build()

	if expected := "INSERT INTO foo (`aa`, ccc) VALUES (?, ?)"; sql != expected {
		t.Fatalf("invalid sql. [expected:%v] [actual:%v]", expected, sql)
	}

	ib = NewStruct(new(structWithQuote)).For(PostgreSQL).InsertInto("foo", structWithQuote{A: "aaa"})
	sql, _ = ib.Build()

	if expected := `INSERT INTO foo ("aa", ccc) VALUES ($1, $2)`; sql != expected {
		t.Fatalf("invalid sql. [expected:%v] [actual:%v]", expected, sql)
	}
}

type structOmitEmpty struct {
	A int      `db:"aa" fieldopt:"omitempty,withquote"`
	B *string  `db:"bb" fieldopt:"omitempty"`
	C uint16   `db:"cc" fieldopt:",omitempty"`
	D *float64 `fieldopt:"omitempty"`
	E bool     `db:"ee"`
}

func TestStructOmitEmpty(t *testing.T) {
	st := NewStruct(new(structOmitEmpty)).For(MySQL)
	sql1, _ := st.Update("foo", new(structOmitEmpty)).Build()

	if expected := "UPDATE foo SET ee = ?"; sql1 != expected {
		t.Fatalf("invalid sql. [expected:%v] [actual:%v]", expected, sql1)
	}

	a := 123
	b := "bbbb"
	c := uint16(234)
	d := 123.45
	e := true
	sql2, args2 := st.Update("foo", &structOmitEmpty{
		A: a,
		B: &b,
		C: c,
		D: &d,
		E: e,
	}).Build()

	if expected := "UPDATE foo SET `aa` = ?, bb = ?, cc = ?, D = ?, ee = ?"; sql2 != expected {
		t.Fatalf("invalid sql. [expected:%v] [actual:%v]", expected, sql2)
	}

	if expected := []interface{}{a, b, c, d, e}; !reflect.DeepEqual(expected, args2) {
		t.Fatalf("invalid args. [expected:%#v] [actual:%#v]", expected, args2)
	}
}

type structWithPointers struct {
	A int      `db:"aa" fieldopt:"omitempty"`
	B *string  `db:"bb"`
	C *float64 `db:"cc" fieldopt:"omitempty"`
}

func TestStructWithPointers(t *testing.T) {
	st := NewStruct(new(structWithPointers)).For(MySQL)
	sql1, _ := st.Update("foo", new(structWithPointers)).Build()

	if expected := "UPDATE foo SET bb = ?"; sql1 != expected {
		t.Fatalf("invalid sql. [expected:%v] [actual:%v]", expected, sql1)
	}

	a := 123
	c := 123.45
	sql2, args2 := st.Update("foo", &structWithPointers{
		A: a,
		C: &c,
	}).Build()

	if expected := "UPDATE foo SET aa = ?, bb = ?, cc = ?"; sql2 != expected {
		t.Fatalf("invalid sql. [expected:%v] [actual:%v]", expected, sql2)
	}

	if expected := []interface{}{a, (*string)(nil), c}; !reflect.DeepEqual(expected, args2) {
		t.Fatalf("invalid args. [expected:%#v] [actual:%#v]", expected, args2)
	}
}
