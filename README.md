# SQL builder for Go #

[![Build Status](https://travis-ci.org/huandu/go-sqlbuilder.svg?branch=master)](https://travis-ci.org/huandu/go-sqlbuilder)
[![GoDoc](https://godoc.org/github.com/huandu/go-sqlbuilder?status.svg)](https://godoc.org/github.com/huandu/go-sqlbuilder)

Package `sqlbuilder` provides a set of flexible and powerful SQL string builders. The only goal of this package is to build SQL string with arguments which can be used in `DB#Query` or `DB#Exec` defined in package `database/sql`.

## Install ##

Use `go get` to install this package.

    go get -u github.com/huandu/go-sqlbuilder

## Usage ##

### Basic usage ###

Here is a sample to demonstrate how to build a SELECT query.

```go
sb := sqlbuilder.NewSelectBuilder()

sb.Select("id", "name", sb.As("COUNT(*)", c))
sb.From("user")
sb.Where(sb.In("status", 1, 2, 5))

sql, args := sb.Build()
fmt.Println(sql)
fmt.Println(args)

// Output:
// SELECT id, name, COUNT(*) AS c FROM user WHERE status IN (?, ?, ?)
// [1 2 5]
```

Following builders are implemented right now. API document and examples are provided in the `godoc` document.

* [Struct](https://godoc.org/github.com/huandu/go-sqlbuilder#Struct): Builder factory for a struct.
* [SelectBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#SelectBuilder): Builder for SELECT.
* [InsertBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#InsertBuilder): Builder for INSERT.
* [UpdateBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#UpdateBuilder): Builder for UPDATE.
* [DeleteBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#DeleteBuilder): Builder for DELETE.
* [Build](https://godoc.org/github.com/huandu/go-sqlbuilder#Build): Freestyle builder using `fmt.Sprintf`-like syntax.
* [Buildf](https://godoc.org/github.com/huandu/go-sqlbuilder#Buildf): Advanced freestyle builder using special syntax defined in [Args#Compile](https://godoc.org/github.com/huandu/go-sqlbuilder#Args.Compile).

### Using `Struct` as a light weight ORM ###

`Struct` stores information of type and struct fields of a struct. It works like a factory of builders. We can use `Struct` methods to create initialized SELECT/INSERT/UPDATE/DELETE builders. It can help us to save time and avoid human-error when writing column names in query for a table.

We can define a struct type and set some field tags to let `Struct` know how to use each field.

```go
type ATable struct {
    Field1     string                                  // If a field doesn't has a tag, use "Field1" as column name in SQL.
    Field2     int    `db:"field2"`                    // Use "db" in field tag to set column name used in SQL.
    Field3     int64  `db:"field3" fieldtag:"foo,bar"` // Set fieldtag to a field. We can use methods like `Struct#SelectForTag` to use it.
    Field4     int64  `db:"field4" fieldtag:"foo"`     // If we use `s.SelectForTag(table, "foo")`, columnes of SELECT are field3 and field3.
    Ignored    int32  `db:"-"`                         // If we set field name as "-", Struct will ignore it.
    unexported int                                     // Unexported field is not visible to Struct.
}
```

Read [examples](https://godoc.org/github.com/huandu/go-sqlbuilder#Struct) for `Struct` to learn how to use it.

What's more, we can use `Struct` as a kind of ORM. It's quite light weight and clean without any magic comparing with other ORM package. It just creates necessary SQL to query rows and takes address of all selected fields of a struct to let it work well with `Rows#Scan` or `Row#Scan` defined in `database/sql`.

```go
type User struct {
    ID     int64  `db:"id"`
    Name   string `db:"name"`
    Status int    `db:"status"`
}

var userStruct = NewStruct(new(User))

func ExampleStruct() {
    // Prepare SELECT query.
    //     SELECT id, name, status FROM user WHERE id = 1234 LIMIT 1
    sb := userStruct.SelectFrom("user")
    sb.Where(sb.E("id", 1234))

    // Execute the query.
    sql, args := sb.Build()
    rows, _ := db.Query(sql, args...)
    defer rows.Close()

    // Scan row data and set value to user.
    // Suppose we get following data.
    //
    //     |  id  |  name  | status |
    //     |------|--------|--------|
    //     | 1234 | huandu | 1      |
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
```

### Nested SQL ###

It's quite straight forward to create a nested SQL: use a builder as an argument to nest it.

Here is a simple sample.

```go
sb := sqlbuilder.NewSelectBuilder()
fromSb := sqlbuilder.NewSelectBuilder()
statusSb := sqlbuilder.NewSelectBuilder()

sb.Select("id")
sb.From(sb.As(fmt.Sprintf("(%v)", sb.Var(fromSb), "user")))
sb.Where(sb.In("status", statusSb))

fromSb.Select("id")
fromSb.From("user")
fromSb.Where(sb.G("level", 4))

statusSb.Select("status")
statusSb.From("config")
statusSb.Where(sb.E("state", 1))

sql, args := sb.Build()
fmt.Println(sql)
fmt.Println(args)

// Output:
// SELECT id FROM (SELECT id FROM user WHERE level > ?) AS user WHERE status IN (SELECT status FROM config WHERE state = ?)
// [4 1]
```

### Use `sql.Named` in a builder ###

The function `sql.Named` defined in `database/sql` can create a named argument in SQL. It's necessary if we want to reuse an argument several times in one SQL. It's still quite simple to use named arguments in a builder: use it as an argument.

Here is a sample.

```go
now := time.Now().Unix()
start := sql.Named("start", now-86400)
end := sql.Named("end", now+86400)
sb := sqlbuilder.NewSelectBuilder()

sb.Select("name")
sb.From("user")
sb.Where(
    sb.Between("created_at", start, end),
    sb.GE("modified_at", start),
)

sql, args := sb.Build()
fmt.Println(sql)
fmt.Println(args)

// Output:
// SELECT name FROM user WHERE created_at BETWEEN @start AND @end AND modified_at >= @start
// [{{} start 1514458225} {{} end 1514544625}]
```

### Special argument types ###

There are several special argument types.

* `List(arg)` represents a list of arguments. If `arg` is a slice or array, e.g. a slice with 3 ints, it will be compiled to `?, ?, ?` and flattened in the final arguments as 3 ints. It's a tool for convenience. We can use it in the `IN` expression or `VALUES` of `INSERT INTO`.
* `Named(name, arg)` represents a named argument. It only works with `Build` to define a named placeholder using syntax `${name}`.
* `Raw(expr)` marks an `expr` as a plain string rather than an argument. The `expr` will not be included in the final arguments after `Compile`.

### Freestyle builder ###

A builder is only a way to record arguments. If we want to build a long SQL with lots of special syntax (e.g. special comments for a database proxy), simply use `Buildf` to format a SQL string using a `fmt.Sprintf`-like syntax.

```go
sb := sqlbuilder.NewSelectBuilder()
sb.Select("id").From("user")

explain := sqlbuilder.Buildf("EXPLAIN %v LEFT JOIN SELECT * FROM banned WHERE state IN (%v, %v)", sb, 1, 2)
sql, args := explain.Build()
fmt.Println(sql)
fmt.Println(args)

// Output:
// EXPLAIN SELECT id FROM user LEFT JOIN SELECT * FROM banned WHERE state IN (?, ?)
// [1 2]
```

### Using special syntax to build SQL ###

Package `sqlbuilder` defines special syntax to represent an uncompiled SQL internally. If we want to take advantage of the syntax to build customized tools, we can use `Build` to compile it with arguments.

The format string uses special syntax to represent arguments.

* `$?` refers successive arguments passed in the call. It works similar as `%v` in `fmt.Sprintf`.
* `$0` `$1` ... `$n` refers nth-argument passed in the call. Next `$?` will use arguments n+1.
* `${name}` refers a named argument created by `Named` with `name`.
* `$$` is a `"$"` string.

```go
sb := sqlbuilder.NewSelectBuilder()
sb.Select("id").From("user").Where(sb.In("status", 1, 2))

b := sqlbuilder.Build("EXPLAIN $? LEFT JOIN SELECT * FROM $? WHERE created_at > $? AND state IN (${states}) AND modified_at BETWEEN $2 AND $?",
    sb, sqlbuilder.Raw("banned"), 1514458225, 1514544625, sqlbuilder.Named("states", sqlbuilder.List([]int{3, 4, 5})))
sql, args := b.Build()

fmt.Println(sql)
fmt.Println(args)

// Output:
// EXPLAIN SELECT id FROM user WHERE status IN (?, ?) LEFT JOIN SELECT * FROM banned WHERE created_at > ? AND state IN (?, ?, ?) AND modified_at BETWEEN ? AND ?
// [1 2 1514458225 3 4 5 1514458225 1514544625]
```

## License ##

This package is licensed under MIT license. See LICENSE for details.