// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"testing"
	"time"

	"github.com/huandu/go-assert"
)

type structUserForTest struct {
	ID        int    `db:"id" fieldtag:"important"`
	Name      string `fieldtag:"important"`
	Status    int    `db:"status" fieldtag:"important"`
	CreatedAt int    `db:"created_at"`
}

var userForTest = NewStruct(new(structUserForTest))

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

func (db testDB) Query(string, ...interface{}) (testRows, error) {
	return 0, nil
}

func (db testDB) Exec(string, ...interface{}) {
	return
}

func (rows testRows) Close() error {
	return nil
}

func (rows testRows) Scan(dest ...interface{}) error {
	_, _ = fmt.Sscan("1234 huandu 1", dest...)
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
		defer func(rows testRows) {
			_ = rows.Close()
		}(rows)
		_ = rows.Scan(orderStruct.AddrForTag(tag, &order)...)

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
		defer func(rows testRows) {
			_ = rows.Close()
		}(rows)
		_ = rows.Scan(orderStruct.AddrForTag(tag, &order)...)

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

func ExampleStruct_forCQL() {
	userStruct := NewStruct(new(User)).For(CQL)

	sb := userStruct.SelectFrom("user")
	sb.Where(sb.E("id", 1234))
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
