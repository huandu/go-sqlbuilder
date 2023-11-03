// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/huandu/go-assert"
)

type structUserForTest struct {
	ID         int    `db:"id" fieldtag:"important"`
	Name       string `fieldtag:"important"`
	Status     int    `db:"status" fieldtag:"important"`
	CreatedAt  int    `db:"created_at"`
	unexported struct{}
}

var userForTest = NewStruct(new(structUserForTest))
var _ = new(structUserForTest).unexported // disable lint warning

func TestStructSelectFrom(t *testing.T) {
	a := assert.New(t)
	sb := userForTest.SelectFrom("user")
	sql, args := sb.Build()

	a.Equal(sql, "SELECT user.id, user.Name, user.status, user.created_at FROM user")
	a.Equal(args, nil)
}

func TestStructSelectFromForTag(t *testing.T) {
	a := assert.New(t)
	sb := userForTest.SelectFromForTag("user", "important")
	sql, args := sb.Build()

	a.Equal(sql, "SELECT user.id, user.Name, user.status FROM user")
	a.Equal(args, nil)
}

func TestStructUpdate(t *testing.T) {
	a := assert.New(t)
	user := &structUserForTest{
		ID:        123,
		Name:      "Huan Du",
		Status:    2,
		CreatedAt: 1234567890,
	}
	ub := userForTest.Update("user", user)
	sql, args := ub.Build()

	a.Equal(sql, "UPDATE user SET id = ?, Name = ?, status = ?, created_at = ?")
	a.Equal(args, []interface{}{123, "Huan Du", 2, 1234567890})
}

func TestStructUpdateForTag(t *testing.T) {
	a := assert.New(t)
	user := &structUserForTest{
		ID:        123,
		Name:      "Huan Du",
		Status:    2,
		CreatedAt: 1234567890,
	}
	ub := userForTest.UpdateForTag("user", "important", user)
	sql, args := ub.Build()

	a.Equal(sql, "UPDATE user SET id = ?, Name = ?, status = ?")
	a.Equal(args, []interface{}{123, "Huan Du", 2})
}

func TestStructInsertInto(t *testing.T) {
	a := assert.New(t)
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
		a.Equal(sql, exceptedVerb+"INTO user (id, Name, status, created_at) VALUES (?, ?, ?, ?)")
		a.Equal(args, []interface{}{123, "Huan Du", 2, 1234567890})
	}

	for ib, exceptedVerb := range testMulitInsert {
		sql, args := ib.Build()
		a.Equal(sql, exceptedVerb+"INTO user (id, Name, status, created_at) VALUES (?, ?, ?, ?), (?, ?, ?, ?)")
		a.Equal(args, []interface{}{123, "Huan Du", 2, 1234567890, 456, "Du Huan", 2, 1234567890})
	}
}

func TestStructInsertIntoForTag(t *testing.T) {
	a := assert.New(t)
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
		a.Equal(sql, exceptedVerb+"INTO user (id, Name, status) VALUES (?, ?, ?)")
		a.Equal(args, []interface{}{123, "Huan Du", 2})
	}

	for ib, exceptedVerb := range testMulitInsertForTag {
		sql, args := ib.Build()
		a.Equal(sql, exceptedVerb+"INTO user (id, Name, status) VALUES (?, ?, ?), (?, ?, ?)")
		a.Equal(args, []interface{}{123, "Huan Du", 2, 456, "Du Huan", 2})
	}
}

func TestStructDeleteFrom(t *testing.T) {
	a := assert.New(t)
	db := userForTest.DeleteFrom("user")
	sql, args := db.Build()

	a.Equal(sql, "DELETE FROM user")
	a.Equal(args, nil)
}

func TestStructAddr(t *testing.T) {
	a := assert.New(t)
	user := new(structUserForTest)
	expected := &structUserForTest{
		ID:        123,
		Name:      "huandu",
		Status:    2,
		CreatedAt: 1234567890,
	}
	str := fmt.Sprintf("%v %v %v %v", expected.ID, expected.Name, expected.Status, expected.CreatedAt)
	_, _ = fmt.Sscanf(str, "%d%s%d%d", userForTest.Addr(user)...)

	a.Equal(user, expected)
}

func TestStructAddrForTag(t *testing.T) {
	a := assert.New(t)
	user := new(structUserForTest)
	expected := &structUserForTest{
		ID:        123,
		Name:      "huandu",
		Status:    2,
		CreatedAt: 1234567890,
	}
	user.CreatedAt = 9876543210
	str := fmt.Sprintf("%v %v %v %v", expected.ID, expected.Name, expected.Status, expected.CreatedAt)
	_, _ = fmt.Sscanf(str, "%d%s%d%d", userForTest.AddrForTag("important", user)...)
	expected.CreatedAt = 9876543210

	a.Equal(user, expected)
	a.Equal(userForTest.AddrForTag("invalid", user), nil)
}

func TestStructAddrWithCols(t *testing.T) {
	a := assert.New(t)
	user := new(structUserForTest)
	expected := &structUserForTest{
		ID:        123,
		Name:      "huandu",
		Status:    2,
		CreatedAt: 1234567890,
	}
	str := fmt.Sprintf("%v %v %v %v", expected.Name, expected.ID, expected.CreatedAt, expected.Status)
	_, _ = fmt.Sscanf(str, "%s%d%d%d", userForTest.AddrWithCols([]string{"Name", "id", "created_at", "status"}, user)...)

	a.Equal(user, expected)
	a.Equal(userForTest.AddrWithCols([]string{"invalid", "non-exist"}, user), nil)
}

func TestStructValues(t *testing.T) {
	a := assert.New(t)
	st := &structUserForTest{
		ID:        123,
		Name:      "huandu",
		Status:    2,
		CreatedAt: 1234567890,
	}
	expected := fmt.Sprintf("%v %v %v %v", st.ID, st.Name, st.Status, st.CreatedAt)
	actual := fmt.Sprintf("%v %v %v %v", userForTest.Values(st)...)

	a.Equal(actual, expected)
}

func TestStructValuesForTag(t *testing.T) {
	a := assert.New(t)
	st := &structUserForTest{
		ID:        123,
		Name:      "huandu",
		Status:    2,
		CreatedAt: 1234567890,
	}
	expected := fmt.Sprintf("%v %v %v", st.ID, st.Name, st.Status)
	actual := fmt.Sprintf("%v %v %v", userForTest.ValuesForTag("important", st)...)

	a.Equal(actual, expected)
	a.Equal(userForTest.ValuesForTag("invalid", st), nil)
}

func TestStructColumns(t *testing.T) {
	a := assert.New(t)
	a.Equal(userForTest.Columns(), []string{"id", "Name", "status", "created_at"})
	a.Equal(userForTest.ColumnsForTag("important"), []string{"id", "Name", "status"})
	a.Equal(userForTest.ColumnsForTag("invalid"), nil)
}

func TestWithAndWithoutTags(t *testing.T) {
	type Tags struct {
		A int `db:"a" fieldtag:"tag1"`
		B int `db:"b" fieldtag:"tag2"`
		C int `db:"c" fieldtag:"tag3"`
		D int `db:"d" fieldtag:"tag1,tag2"`
		E int `db:"e" fieldtag:"tag2,tag3"`
		F int `db:"f" fieldtag:"tag1,tag3"`
		G int `db:"g" fieldtag:"tag1,tag2,tag3"`
		H int `db:"h"`
	}
	structTags := NewStruct(Tags{})
	a := assert.New(t)

	a.Equal(structTags.Columns(), []string{"a", "b", "c", "d", "e", "f", "g", "h"})
	a.Equal(structTags.WithTag().Columns(), []string{"a", "b", "c", "d", "e", "f", "g", "h"})
	a.Equal(structTags.WithoutTag().Columns(), []string{"a", "b", "c", "d", "e", "f", "g", "h"})
	a.Equal(structTags.WithTag("").Columns(), []string{"a", "b", "c", "d", "e", "f", "g", "h"})
	a.Equal(structTags.WithoutTag("").Columns(), []string{"a", "b", "c", "d", "e", "f", "g", "h"})

	a.Equal(structTags.WithTag("tag1").Columns(), []string{"a", "d", "f", "g"})
	a.Equal(structTags.WithTag("tag2").Columns(), []string{"b", "d", "e", "g"})
	a.Equal(structTags.WithTag("tag3").Columns(), []string{"c", "e", "f", "g"})

	a.Equal(structTags.WithTag("tag1", "tag2").Columns(), []string{"a", "d", "f", "g", "b", "e"})
	a.Equal(structTags.WithTag("tag1", "tag3").Columns(), []string{"a", "d", "f", "g", "c", "e"})
	a.Equal(structTags.WithTag("tag2", "tag3").Columns(), []string{"b", "d", "e", "g", "c", "f"})
	a.Equal(structTags.WithTag("tag2", "tag3", "tag2", "", "tag3").Columns(), []string{"b", "d", "e", "g", "c", "f"})

	a.Equal(structTags.WithoutTag("tag3").Columns(), []string{"a", "b", "d", "h"})
	a.Equal(structTags.WithoutTag("tag3", "tag2").Columns(), []string{"a", "h"})
	a.Equal(structTags.WithoutTag("tag3", "tag2", "tag3", "", "tag2").Columns(), []string{"a", "h"})

	a.Equal(structTags.WithTag("tag1", "tag2").WithoutTag("tag3").Columns(), []string{"a", "d", "b"})
	a.Equal(structTags.WithoutTag("tag3").WithTag("tag1", "tag2").Columns(), []string{"a", "d", "b"})
	a.Equal(structTags.WithTag("tag1", "tag2", "tag3").WithoutTag("tag3").Columns(), []string{"a", "d", "b"})
	a.Equal(structTags.WithoutTag("tag3", "tag1").WithTag("tag1", "tag2", "tag3").Columns(), []string{"b"})

	a.Equal(structTags.WithTag("tag2").WithTag("tag1").Columns(), []string{"a", "d", "f", "g", "b", "e"})
	a.Equal(structTags.WithoutTag("tag3").WithTag("tag1").WithTag("tag3", "", "tag2").Columns(), []string{"a", "d", "b"})
	a.Equal(structTags.WithoutTag("tag3").WithTag("tag1").WithTag("tag3", "tag2").WithoutTag("tag1", "", "tag3").Columns(), []string{"b"})
}

func TestStructForeachRead(t *testing.T) {
	// a := assert.New(t)
	userForTest.ForeachRead(func(dbtag string, isQuoted bool, field reflect.StructField) {
		t.Logf("%s\n", dbtag)
	})
}

type State int
type testDB int
type testRows int

func (db *testDB) Query(string, ...interface{}) (testRows, error) {
	rows := testRows(*db)
	*db++
	return rows, nil
}

func (db *testDB) Exec(query string, args ...interface{}) {
}

func (rows testRows) Close() error {
	return nil
}

func (rows testRows) Scan(dest ...interface{}) error {
	if rows == 0 {
		fmt.Sscan("1234 huandu 1", dest...)
	} else if rows == 1 {
		fmt.Sscan("1234 34 huandu 1456725903636000000", dest...)
	} else if rows == 2 {
		fmt.Sscan("1 1456725903636000000", dest...)
	} else {
		panic("invalid rows")
	}

	return nil
}

var userDB testDB = 0

const (
	OrderStateInvalid State = iota
	OrderStateCreated
	OrderStatePaid
)

func ExampleStruct_useStructAsORM() {
	// Suppose we defined following type for user db.
	type User struct {
		ID     int64  `db:"id" fieldtag:"pk"`
		Name   string `db:"name"`
		Status int    `db:"status"`
	}

	// Parse user struct. The userStruct can be a global variable.
	// It's guraanteed to be thread-safe.
	var userStruct = NewStruct(new(User))

	// Prepare SELECT query.
	sb := userStruct.SelectFrom("user")
	sb.Where(sb.Equal("id", 1234))

	// Execute the query.
	sql, args := sb.Build()
	rows, _ := userDB.Query(sql, args...)
	defer func(rows testRows) {
		_ = rows.Close()
	}(rows)

	// Scan row data to user.
	var user User
	_ = rows.Scan(userStruct.Addr(&user)...)

	fmt.Println(sql)
	fmt.Println(args)
	fmt.Printf("%#v", user)

	// Output:
	// SELECT user.id, user.name, user.status FROM user WHERE id = ?
	// [1234]
	// sqlbuilder.User{ID:1234, Name:"huandu", Status:1}
}

var orderDB testDB = 1

func ExampleStruct_WithTag() {
	// Suppose we defined following type for an order.
	type Order struct {
		ID         int64  `db:"id"`
		State      State  `db:"state" fieldtag:"paid"`
		SkuID      int64  `db:"sku_id"`
		UserID     int64  `db:"user_id"`
		Price      int64  `db:"price" fieldtag:"update"`
		Discount   int64  `db:"discount" fieldtag:"update"`
		Desc       string `db:"desc" fieldtag:"new,update" fieldopt:"withquote"`
		CreatedAt  int64  `db:"created_at"`
		ModifiedAt int64  `db:"modified_at" fieldtag:"update,paid"`
	}

	// The orderStruct is a global variable for Order type.
	var orderStruct = NewStruct(new(Order))

	// Create an order with all fields set.
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
		orderDB.Exec(sql, args)
		fmt.Println(sql)
	}

	// Update order only with price related fields, which is tagged with "update".
	updatePrice := func(table string) {
		// Use tag "update" in all struct methods.
		st := orderStruct.WithTag("update")

		// Read order from database.
		var order Order
		sql, args := st.SelectFrom(table).Where("id = 1234").Build()
		rows, _ := orderDB.Query(sql, args...)
		defer func(rows testRows) {
			_ = rows.Close()
		}(rows)
		_ = rows.Scan(st.Addr(&order)...)
		fmt.Println(sql)

		// Discount for this user.
		// Use tag "update" to update necessary columns only.
		order.Discount += 100
		order.ModifiedAt = time.Now().Unix()

		// Save the order.
		b := st.Update(table, &order)
		b.Where(b.E("id", order.ID))
		sql, args = b.Build()
		orderDB.Exec(sql, args...)
		fmt.Println(sql)
	}

	// Update order only with payment related fields, which is tagged with "paid".
	updateState := func(table string) {
		st := orderStruct.WithTag("paid")

		// Read order from database.
		var order Order
		sql, args := st.SelectFrom(table).Where("id = 1234").Build()
		rows, _ := orderDB.Query(sql, args...)
		defer func(rows testRows) {
			_ = rows.Close()
		}(rows)
		_ = rows.Scan(st.Addr(&order)...)
		fmt.Println(sql)

		// Update state to paid when user has paid for the order.
		// Use tag "paid" to update necessary columns only.
		if order.State != OrderStateCreated {
			// Report state error here.
			panic(order.State)
			// return
		}

		// Update order state.
		order.State = OrderStatePaid
		order.ModifiedAt = time.Now().Unix()

		// Save the order.
		b := st.Update(table, &order)
		b.Where(b.E("id", order.ID))
		sql, args = b.Build()
		orderDB.Exec(sql, args...)
		fmt.Println(sql)
	}

	table := "order"
	createOrder(table)
	updatePrice(table)
	updateState(table)

	// Output:
	// INSERT INTO order (id, state, sku_id, user_id, price, discount, `desc`, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	// SELECT order.price, order.discount, order.`desc`, order.modified_at FROM order WHERE id = 1234
	// UPDATE order SET price = ?, discount = ?, `desc` = ?, modified_at = ? WHERE id = ?
	// SELECT order.state, order.modified_at FROM order WHERE id = 1234
	// UPDATE order SET state = ?, modified_at = ? WHERE id = ?
}

func ExampleStruct_WithoutTag() {
	// We can use WithoutTag to exclude fields with specific tag.
	// It's useful when we want to update all fields except some fields.

	type User struct {
		ID             int64     `db:"id" fieldtag:"pk"`
		FirstName      string    `db:"first_name"`
		LastName       string    `db:"last_name"`
		ModifiedAtTime time.Time `db:"modified_at_time"`
	}

	// The userStruct is a global variable for User type.
	var userStruct = NewStruct(new(User))

	// Update user with all fields except the user_id field which is tagged with "pk".
	user := &User{
		FirstName:      "Huan",
		LastName:       "Du",
		ModifiedAtTime: time.Now(),
	}
	sql, _ := userStruct.WithoutTag("pk").Update("user", user).Where("id = 1234").Build()
	fmt.Println(sql)

	// Output:
	// UPDATE user SET first_name = ?, last_name = ?, modified_at_time = ? WHERE id = 1234
}

func ExampleStruct_buildUPDATE() {
	// Suppose we defined following type for user db.
	type User struct {
		ID     int64  `db:"id" fieldtag:"pk"`
		Name   string `db:"name"`
		Status int    `db:"status"`
	}

	// Parse user struct. The userStruct can be a global variable.
	// It's guraanteed to be thread-safe.
	var userStruct = NewStruct(new(User))

	// Prepare UPDATE query.
	// We should not update the primary key field.
	user := &User{
		ID:     1234,
		Name:   "Huan Du",
		Status: 1,
	}
	ub := userStruct.WithoutTag("pk").Update("user", user)
	ub.Where(ub.Equal("id", user.ID))

	// Execute the query.
	sql, args := ub.Build()
	orderDB.Exec(sql, args...)

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// UPDATE user SET name = ?, status = ? WHERE id = ?
	// [Huan Du 1 1234]
}

func ExampleStruct_buildINSERT() {
	// Suppose we defined following type for user db.
	type User struct {
		ID     int64  `db:"id" fieldtag:"pk"`
		Name   string `db:"name"`
		Status int    `db:"status"`
	}

	// Parse user struct. The userStruct can be a global variable.
	// It's guraanteed to be thread-safe.
	var userStruct = NewStruct(new(User))

	// Prepare INSERT query.
	// Suppose that user id is generated by database.
	user := &User{
		Name:   "Huan Du",
		Status: 1,
	}
	ib := userStruct.WithoutTag("pk").InsertInto("user", user)

	// Execute the query.
	sql, args := ib.Build()
	orderDB.Exec(sql, args...)

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO user (name, status) VALUES (?, ?)
	// [Huan Du 1]
}

func ExampleStruct_buildDELETE() {
	// Suppose we defined following type for user db.
	type User struct {
		ID     int64  `db:"id" fieldtag:"pk"`
		Name   string `db:"name"`
		Status int    `db:"status"`
	}

	// Parse user struct. The userStruct can be a global variable.
	// It's guraanteed to be thread-safe.
	var userStruct = NewStruct(new(User))

	// Prepare DELETE query.
	user := &User{
		ID:     1234,
		Name:   "Huan Du",
		Status: 1,
	}
	b := userStruct.DeleteFrom("user")
	b.Where(b.Equal("id", user.ID))

	// Execute the query.
	sql, args := b.Build()
	orderDB.Exec(sql, args...)

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// DELETE FROM user WHERE id = ?
	// [1234]
}

func ExampleStruct_forPostgreSQL() {
	// Suppose we defined following type for user db.
	type User struct {
		ID     int64  `db:"id" fieldtag:"pk"`
		Name   string `db:"name"`
		Status int    `db:"status"`
	}

	// Parse user struct. The userStruct can be a global variable.
	// It's guraanteed to be thread-safe.
	var userStruct = NewStruct(new(User)).For(PostgreSQL)

	sb := userStruct.SelectFrom("user")
	sb.Where(sb.Equal("id", 1234))
	sql, args := sb.Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT user.id, user.name, user.status FROM user WHERE id = $1
	// [1234]
}

func ExampleStruct_forCQL() {
	// Suppose we defined following type for user db.
	type User struct {
		ID     int64  `db:"id" fieldtag:"pk"`
		Name   string `db:"name"`
		Status int    `db:"status"`
	}

	// Parse user struct. The userStruct can be a global variable.
	// It's guraanteed to be thread-safe.
	userStruct := NewStruct(new(User)).For(CQL)

	sb := userStruct.SelectFrom("user")
	sb.Where(sb.Equal("id", 1234))
	sql, args := sb.Build()

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT id, name, status FROM user WHERE id = ?
	// [1234]
}

type structWithQuote struct {
	A string  `db:"aa" fieldopt:"withquote"`
	B int     `db:"-" fieldopt:"withquote"` // fieldopt is ignored as db is "-".
	C float64 `db:"ccc"`
}

func TestStructWithQuote(t *testing.T) {
	a := assert.New(t)
	sb := NewStruct(new(structWithQuote)).For(MySQL).SelectFrom("foo")
	sql, _ := sb.Build()
	a.Equal(sql, "SELECT foo.`aa`, foo.ccc FROM foo")

	sb = NewStruct(new(structWithQuote)).For(PostgreSQL).SelectFrom("foo")
	sql, _ = sb.Build()
	a.Equal(sql, `SELECT foo."aa", foo.ccc FROM foo`)

	sb = NewStruct(new(structWithQuote)).For(CQL).SelectFrom("foo")
	sql, _ = sb.Build()
	a.Equal(sql, "SELECT 'aa', ccc FROM foo")

	ub := NewStruct(new(structWithQuote)).For(MySQL).Update("foo", structWithQuote{A: "aaa"})
	sql, _ = ub.Build()
	a.Equal(sql, "UPDATE foo SET `aa` = ?, ccc = ?")

	ub = NewStruct(new(structWithQuote)).For(PostgreSQL).Update("foo", structWithQuote{A: "aaa"})
	sql, _ = ub.Build()
	a.Equal(sql, `UPDATE foo SET "aa" = $1, ccc = $2`)

	ub = NewStruct(new(structWithQuote)).For(CQL).Update("foo", structWithQuote{A: "aaa"})
	sql, _ = ub.Build()
	a.Equal(sql, `UPDATE foo SET 'aa' = ?, ccc = ?`)

	ib := NewStruct(new(structWithQuote)).For(MySQL).InsertInto("foo", structWithQuote{A: "aaa"})
	sql, _ = ib.Build()
	a.Equal(sql, "INSERT INTO foo (`aa`, ccc) VALUES (?, ?)")

	ib = NewStruct(new(structWithQuote)).For(PostgreSQL).InsertInto("foo", structWithQuote{A: "aaa"})
	sql, _ = ib.Build()
	a.Equal(sql, `INSERT INTO foo ("aa", ccc) VALUES ($1, $2)`)

	ib = NewStruct(new(structWithQuote)).For(CQL).InsertInto("foo", structWithQuote{A: "aaa"})
	sql, _ = ib.Build()
	a.Equal(sql, "INSERT INTO foo ('aa', ccc) VALUES (?, ?)")
}

type structOmitEmpty struct {
	A int      `db:"aa" fieldopt:"omitempty,withquote"`
	B *string  `db:"bb" fieldopt:"omitempty"`
	C uint16   `db:"cc" fieldopt:",omitempty"`
	D *float64 `fieldopt:"omitempty"`
	E bool     `db:"ee"`
}

func TestStructOmitEmpty(t *testing.T) {
	a := assert.New(t)
	st := NewStruct(new(structOmitEmpty)).For(MySQL)
	sql1, _ := st.Update("foo", new(structOmitEmpty)).Build()

	a.Equal(sql1, "UPDATE foo SET ee = ?")

	i := 123
	b := "bbbb"
	c := uint16(234)
	d := 123.45
	e := true
	sql2, args2 := st.Update("foo", &structOmitEmpty{
		A: i,
		B: &b,
		C: c,
		D: &d,
		E: e,
	}).Build()

	a.Equal(sql2, "UPDATE foo SET `aa` = ?, bb = ?, cc = ?, D = ?, ee = ?")
	a.Equal(args2, []interface{}{i, b, c, d, e})
}

type structOmitEmptyForTag struct {
	A int      `db:"aa" fieldopt:"omitempty,withquote" fieldtag:"patch"`
	B *string  `db:"bb" fieldopt:"omitempty" fieldtag:"patch"`
	C uint16   `db:"cc" fieldopt:"omitempty()" fieldtag:"patch"`
	D *float64 `fieldopt:"omitempty(patch)" fieldtag:"patch"`
	E bool     `db:"ee" fieldtag:"patch"`
}

func TestStructOmitEmptyForTag(t *testing.T) {
	a := assert.New(t)
	st := NewStruct(new(structOmitEmptyForTag)).For(MySQL)
	sql1, _ := st.Update("foo", new(structOmitEmptyForTag)).Build()

	a.Equal(sql1, "UPDATE foo SET D = ?, ee = ?")

	i := 123
	b := "bbbb"
	c := uint16(234)
	e := true
	sql2, args2 := st.UpdateForTag("foo", "patch", &structOmitEmptyForTag{
		A: i,
		B: &b,
		C: c,
		D: nil,
		E: e,
	}).Build()

	a.Equal(sql2, "UPDATE foo SET `aa` = ?, bb = ?, cc = ?, ee = ?")
	a.Equal(args2, []interface{}{i, b, c, e})
}

type structOmitEmptyForMultipleTags struct {
	A int      `db:"aa" fieldopt:"omitempty, omitempty(patch, patch2),withquote" fieldtag:"patch, patch2 "`
	B *string  `db:"bb" fieldopt:"omitempty" fieldtag:"patch"`
	C uint16   `db:"cc" fieldopt:"omitempty, omitempty(patch2)" fieldtag:"patch2"`
	D *float64 `fieldopt:"omitempty(patch, patch2)" fieldtag:"patch,patch2"`
	E bool     `db:"ee" fieldtag:"patch"`
}

func TestStructOmitEmptyForMultipleTags(t *testing.T) {
	a := assert.New(t)
	st := NewStruct(new(structOmitEmptyForMultipleTags)).For(MySQL)
	sql1, _ := st.Update("foo", new(structOmitEmptyForMultipleTags)).Build()

	a.Equal(sql1, "UPDATE foo SET D = ?, ee = ?")

	i := 123
	b := "bbbb"
	c := uint16(2)
	e := true
	sql2, args2 := st.UpdateForTag("foo", "patch2", &structOmitEmptyForMultipleTags{
		A: i,
		B: &b,
		C: 0,
		D: nil,
		E: e,
	}).Build()

	a.Equal(sql2, "UPDATE foo SET `aa` = ?")
	a.Equal(args2, []interface{}{i})

	value1 := &structOmitEmptyForMultipleTags{
		A: i,
		B: &b,
		C: 0,
		D: nil,
		E: false, // should be false value.
	}
	value2 := &structOmitEmptyForMultipleTags{
		A: i,
		B: &b,
		C: c, // should not be omitted as C in value1 is not empty.
		D: nil,
		E: true,
	}
	sql3, args3 := st.InsertIntoForTag("foo", "patch2", value1, value2).Build()
	a.Equal(sql3, "INSERT INTO foo (`aa`, cc) VALUES (?, ?), (?, ?)")
	a.Equal(args3, []interface{}{i, uint16(0), i, c})
}

type structWithPointers struct {
	A int      `db:"aa" fieldopt:"omitempty"`
	B *string  `db:"bb"`
	C *float64 `db:"cc" fieldopt:"omitempty"`
}

func TestStructWithPointers(t *testing.T) {
	a := assert.New(t)
	st := NewStruct(new(structWithPointers)).For(MySQL)
	sql1, _ := st.Update("foo", new(structWithPointers)).Build()

	a.Equal(sql1, "UPDATE foo SET bb = ?")

	i := 123
	c := 123.45
	sql2, args2 := st.Update("foo", &structWithPointers{
		A: i,
		C: &c,
	}).Build()

	a.Equal(sql2, "UPDATE foo SET aa = ?, bb = ?, cc = ?")
	a.Equal(args2, []interface{}{i, (*string)(nil), c})
}

type structWithMapper struct {
	structWithMapperEmbedded

	FieldName1        string `fieldopt:"withquote"`
	FieldNameSetByTag int    `db:"set_by_tag"`
	FieldNameShadowed int    `db:"field_name1"` // Shadowed.
}

type structWithMapperEmbedded struct {
	structWithMapperEmbedded2

	FieldName1     int // Shadowed.
	EmbeddedField2 int
}

type structWithMapperEmbedded2 struct {
	EmbeddedAndEmbeddedField1 string
}

func TestStructFieldMapper(t *testing.T) {
	a := assert.New(t)

	old := DefaultFieldMapper
	defer func() {
		DefaultFieldMapper = old
	}()

	DefaultFieldMapper = SnakeCaseMapper
	s := NewStruct(new(structWithMapper))
	sWithoutMapper := s.WithFieldMapper(nil) // Columns in s will not be changed after this call.
	sql, _ := s.SelectFrom("t").Build()
	a.Equal(sql, "SELECT t.`field_name1`, t.set_by_tag, t.embedded_field2, t.embedded_and_embedded_field1 FROM t")

	expected := &structWithMapper{
		FieldName1:        "field",
		FieldNameSetByTag: 123,

		structWithMapperEmbedded: structWithMapperEmbedded{
			structWithMapperEmbedded2: structWithMapperEmbedded2{
				EmbeddedAndEmbeddedField1: "embedded",
			},
			EmbeddedField2: 456,
		},
	}
	var actual structWithMapper
	str := fmt.Sprintf("%v %v %v %v", expected.FieldName1, expected.FieldNameSetByTag, expected.EmbeddedField2, expected.EmbeddedAndEmbeddedField1)
	_, _ = fmt.Sscanf(str, "%d%s%d%d", s.Addr(&actual)...)

	sql, _ = sWithoutMapper.SelectFrom("t").Build()
	a.Equal(sql, "SELECT t.`FieldName1`, t.set_by_tag, t.field_name1, t.EmbeddedField2, t.EmbeddedAndEmbeddedField1 FROM t")
}

type structWithAs struct {
	T1 string `db:"t1" fieldas:"f1" fieldtag:"tag"`
	T2 string `db:"t2" fieldas:""`                  // Empty fieldas is the same as the tag is not set.
	T3 string `db:"t2" fieldas:"f3"`                // AS works without db tag.
	T4 string `db:"t4" fieldas:"f3" fieldtag:"tag"` // It's OK to set the same fieldas in different tags.
}

func TestStructFieldAs(t *testing.T) {
	a := assert.New(t)
	s := NewStruct(new(structWithAs))
	value := &structWithAs{
		T1: "t1",
		T2: "t2",
		T3: "t3",
		T4: "t4",
	}
	build := func(builder Builder) string {
		sql, _ := builder.Build()
		return sql
	}

	// Struct field T3 is not shadowed by T2.
	// Struct field T4 is shadowed by T3 due to same fieldas.
	sql := build(s.SelectFrom("t"))
	a.Equal(sql, `SELECT t.t1 AS f1, t.t2, t.t2 AS f3 FROM t`)

	// Struct field T4 is visible in the tag.
	sql = build(s.WithTag("tag").SelectFrom("t"))
	a.Equal(sql, `SELECT t.t1 AS f1, t.t4 AS f3 FROM t`)

	// Struct field T3 is shadowed by T2 due to same alias.
	sql = build(s.Update("t", value))
	a.Equal(sql, `UPDATE t SET t1 = ?, t2 = ?, t4 = ?`)
}

type structImplValuer int

func (v *structImplValuer) Value() (driver.Value, error) {
	return *v * 2, nil
}

type structContainsValuer struct {
	F1 string
	F2 *structImplValuer
}

func TestStructFieldsImplValuer(t *testing.T) {
	a := assert.New(t)
	st := NewStruct(new(structContainsValuer))
	f1 := "foo"
	f2 := structImplValuer(100)

	sql, args := st.Update("t", structContainsValuer{
		F1: f1,
		F2: &f2,
	}).BuildWithFlavor(MySQL)

	a.Equal(sql, "UPDATE t SET F1 = ?, F2 = ?")
	a.Equal(args[0], f1)
	a.Equal(args[1], &f2)

	result, err := MySQL.Interpolate(sql, args)
	a.NilError(err)
	a.Equal(result, "UPDATE t SET F1 = 'foo', F2 = 200")
}

func SomeOtherMapper(string) string {
	return ""
}

func ExampleFieldMapperFunc() {
	type Orders struct {
		ID            int64
		UserID        int64
		ProductName   string
		Status        int
		UserAddrLine1 string
		UserAddrLine2 string
		CreatedAt     time.Time
	}

	// Create a Struct for Orders.
	orders := NewStruct(new(Orders))

	// Set the default field mapper to snake_case mapper globally.
	DefaultFieldMapper = SnakeCaseMapper

	// Field names are converted to snake_case words.
	sql1, _ := orders.SelectFrom("orders").Limit(10).Build()

	fmt.Println(sql1)

	// Changing the default field mapper will *NOT* affect field names in orders.
	// Once field name conversion is done, they will not be changed again.
	DefaultFieldMapper = SomeOtherMapper
	sql2, _ := orders.SelectFrom("orders").Limit(10).Build()

	fmt.Println(sql1 == sql2)

	// Output:
	// SELECT orders.id, orders.user_id, orders.product_name, orders.status, orders.user_addr_line1, orders.user_addr_line2, orders.created_at FROM orders LIMIT 10
	// true
}
