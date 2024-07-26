# SQL builder for Go

[![Go](https://github.com/huandu/go-sqlbuilder/workflows/Go/badge.svg)](https://github.com/huandu/go-sqlbuilder/actions)
[![GoDoc](https://godoc.org/github.com/huandu/go-sqlbuilder?status.svg)](https://pkg.go.dev/github.com/huandu/go-sqlbuilder)
[![Go Report](https://goreportcard.com/badge/github.com/huandu/go-sqlbuilder)](https://goreportcard.com/report/github.com/huandu/go-sqlbuilder)
[![Coverage Status](https://coveralls.io/repos/github/huandu/go-sqlbuilder/badge.svg?branch=master)](https://coveralls.io/github/huandu/go-sqlbuilder?branch=master) [![Join the chat at https://gitter.im/go-sqlbuilder/community](https://badges.gitter.im/go-sqlbuilder/community.svg)](https://gitter.im/go-sqlbuilder/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

- [Install](#install)
- [Usage](#usage)
  - [Basic usage](#basic-usage)
  - [Pre-defined SQL builders](#pre-defined-sql-builders)
  - [Build `WHERE` clause](#build-where-clause)
  - [Share `WHERE` clause among builders](#share-where-clause-among-builders)
  - [Build SQL for different systems](#build-sql-for-different-systems)
  - [Using `Struct` as a light weight ORM](#using-struct-as-a-light-weight-orm)
  - [Nested SQL](#nested-sql)
  - [Use `sql.Named` in a builder](#use-sqlnamed-in-a-builder)
  - [Argument modifiers](#argument-modifiers)
  - [Freestyle builder](#freestyle-builder)
  - [Using special syntax to build SQL](#using-special-syntax-to-build-sql)
  - [Interpolate `args` in the `sql`](#interpolate-args-in-the-sql)
- [License](#license)

The `sqlbuilder` package implements a series of flexible and powerful SQL string concatenation builders. This package focuses on constructing SQL strings for direct use with the Go standard library's `sql.DB` and `sql.Stmt` related interfaces, and strives to optimize the performance of building SQL and reduce memory consumption.

The initial goal in designing this package was to create a pure SQL construction library that is independent of specific database drivers and business logic. It is designed to meet the needs of enterprise-level scenarios that require various customized database drivers, special operation and maintenance standards, heterogeneous systems, and non-standard SQL in complex situations. Since its open-source inception, this package has been tested in a large enterprise-level application scenario, enduring the pressure of hundreds of millions of orders daily and nearly ten million transactions per day, demonstrating good performance and scalability.

This package does not bind to any specific database driver, nor does it automatically connect to any database. It does not even assume the use of the generated SQL, making it suitable for any application scenario that constructs SQL-like statements. It is also very suitable for secondary development on this basis, to implement more business-related database access packages, ORMs, and so on.

## Install

Use `go get` to install this package.

```shell
go get github.com/huandu/go-sqlbuilder
```

## Usage

### Basic usage

We can build a SQL really quick with this package.

```go
sql := sqlbuilder.Select("id", "name").From("demo.user").
    Where("status = 1").Limit(10).
    String()

fmt.Println(sql)

// Output:
// SELECT id, name FROM demo.user WHERE status = 1 LIMIT 10
```

In the most common cases, we need to escape all input from user. In this case, create a builder before starting.

```go
sb := sqlbuilder.NewSelectBuilder()

sb.Select("id", "name", sb.As("COUNT(*)", "c"))
sb.From("user")
sb.Where(sb.In("status", 1, 2, 5))

sql, args := sb.Build()
fmt.Println(sql)
fmt.Println(args)

// Output:
// SELECT id, name, COUNT(*) AS c FROM user WHERE status IN (?, ?, ?)
// [1 2 5]
```

### Pre-defined SQL builders

This package includes following pre-defined builders so far. API document and examples can be found in the `godoc` online document.

- [Struct](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Struct): Builder factory for a struct.
- [CreateTableBuilder](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#CreateTableBuilder): Builder for CREATE TABLE.
- [SelectBuilder](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#SelectBuilder): Builder for SELECT.
- [InsertBuilder](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#InsertBuilder): Builder for INSERT.
- [UpdateBuilder](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#UpdateBuilder): Builder for UPDATE.
- [DeleteBuilder](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#DeleteBuilder): Builder for DELETE.
- [UnionBuilder](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#UnionBuilder): Builder for UNION and UNION ALL.
- [CTEBuilder](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#CTEBuilder): Builder for Common Table Expression (CTE), e.g. `WITH name (col1, col2) AS (SELECT ...)`.
- [Buildf](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Buildf): Freestyle builder using `fmt.Sprintf`-like syntax.
- [Build](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Build): Advanced freestyle builder using special syntax defined in [Args#Compile](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Args.Compile).
- [BuildNamed](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#BuildNamed): Advanced freestyle builder using `${key}` to refer the value of a map by key.

There is a special method `SQL(sql string)` implemented by all statement builders. We can use this method to insert any arbitrary SQL fragment into a builder when building a SQL. It's quite useful to build SQL containing non-standard syntax supported by a OLTP or OLAP system.

```go
// Build a SQL to create a HIVE table.
sql := sqlbuilder.CreateTable("users").
    SQL("PARTITION BY (year)").
    SQL("AS").
    SQL(
        sqlbuilder.Select("columns[0] id", "columns[1] name", "columns[2] year").
            From("`all-users.csv`").
            Limit(100).
            String(),
    ).
    String()

fmt.Println(sql)

// Output:
// CREATE TABLE users PARTITION BY (year) AS SELECT columns[0] id, columns[1] name, columns[2] year FROM `all-users.csv` LIMIT 100
```

Following are some utility methods to deal with special cases.

- [Flatten](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Flatten) can convert an array-like variable to a flat slice of `[]interface{}` recursively. For instance, calling `Flatten([]interface{"foo", []int{2, 3}})` returns `[]interface{}{"foo", 2, 3}`. This method can work with builder methods like `In`/`NotIn`/`Values`/etc to convert a typed array to `[]interface{}` or merge inputs.
- [List](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#List) works similar to `Flatten` except that its return value is dedecated for builder args. For instance, calling `Buildf("my_func(%v)", List([]int{1, 2, 3})).Build()` returns SQL `my_func(?, ?, ?)` and args `[]interface{}{1, 2, 3}`.
- [Raw](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Raw) marks a string as "raw string" in args. For instance, calling `Buildf("SELECT %v", Raw("NOW()")).Build()` returns SQL `SELECT NOW()`.

To learn how to use builders, check out [examples on GoDoc](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#pkg-examples).

### Build `WHERE` clause

`WHERE` clause is the most important part of a SQL. We can use `Where` method to add one or more conditions to a builder.

To make building `WHERE` clause easier, there is an utility type called `Cond` to build condition. All builders which support `WHERE` clause have an anonymous `Cond` field so that we can call methods implemented by `Cond` on these builders.

```go
sb := sqlbuilder.Select("id").From("user")
sb.Where(
    sb.In("status", 1, 2, 5),
    sb.Or(
        sb.Equal("name", "foo"),
        sb.Like("email", "foo@%"),
    ),
)

sql, args := sb.Build()
fmt.Println(sql)
fmt.Println(args)

// Output:
// SELECT id FROM user WHERE status IN (?, ?, ?) AND (name = ? OR email LIKE ?)
// [1 2 5 foo foo@%]
```

There are many methods for building conditions.

- [Cond.Equal](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.Equal)/[Cond.E](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.E)/[Cond.EQ](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.EQ): `field = value`.
- [Cond.NotEqual](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.NotEqual)/[Cond.NE](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.NE)/[Cond.NEQ](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.NEQ): `field <> value`.
- [Cond.GreaterThan](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.GreaterThan)/[Cond.G](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.G)/[Cond.GT](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.GT): `field > value`.
- [Cond.GreaterEqualThan](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.GreaterEqualThan)/[Cond.GE](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.GE)/[Cond.GTE](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.GTE): `field >= value`.
- [Cond.LessThan](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.LessThan)/[Cond.L](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.L)/[Cond.LT](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.LT): `field < value`.
- [Cond.LessEqualThan](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.LessEqualThan)/[Cond.LE](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.LE)/[Cond.LTE](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.LTE): `field <= value`.
- [Cond.In](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.In): `field IN (value1, value2, ...)`.
- [Cond.NotIn](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.NotIn): `field NOT IN (value1, value2, ...)`.
- [Cond.Like](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.Like): `field LIKE value`.
- [Cond.NotLike](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.NotLike): `field NOT LIKE value`.
- [Cond.Between](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.Between): `field BETWEEN lower AND upper`.
- [Cond.NotBetween](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.NotBetween): `field NOT BETWEEN lower AND upper`.
- [Cond.IsNull](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.IsNull): `field IS NULL`.
- [Cond.IsNotNull](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.IsNotNull): `field IS NOT NULL`.
- [Cond.Exists](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.Exists): `EXISTS (subquery)`.
- [Cond.NotExists](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.NotExists): `NOT EXISTS (subquery)`.
- [Cond.Any](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.Any): `field op ANY (value1, value2, ...)`.
- [Cond.All](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.All): `field op ALL (value1, value2, ...)`.
- [Cond.Some](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.Some): `field op SOME (value1, value2, ...)`.
- [Cond.Var](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.Var): A placeholder for any value.

There are also some methods to combine conditions.

- [Cond.And](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.And): Combine conditions with `AND` operator.
- [Cond.Or](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Cond.Or): Combine conditions with `OR` operator.

### Share `WHERE` clause among builders

Due to the importance of the `WHERE` statement in SQL, we often need to continuously append conditions and even share some common `WHERE` conditions among different builders. Therefore, we abstract the `WHERE` statement into a `WhereClause` struct, which can be used to create reusable `WHERE` conditions.

Here is a sample to show how to copy `WHERE` clause from a `SelectBuilder` to an `UpdateBuilder`.

```go
// Build a SQL to select a user from database.
sb := Select("name", "level").From("users")
sb.Where(
    sb.Equal("id", 1234),
)
fmt.Println(sb)

ub := Update("users")
ub.Set(
    ub.Add("level", 10),
)

// Set the WHERE clause of UPDATE to the WHERE clause of SELECT.
ub.WhereClause = sb.WhereClause
fmt.Println(ub)

// Output:
// SELECT name, level FROM users WHERE id = ?
// UPDATE users SET level = level + ? WHERE id = ?
```

Read samples for [WhereClause](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#WhereClause) to learn how to use it.

### Build SQL for different systems

SQL syntax and parameter marks vary in different systems. In this package, we introduce a concept called "flavor" to smooth out these difference.

Right now, `MySQL`, `PostgreSQL`, `SQLServer`, `SQLite`, `CQL`, `ClickHouse`, `Presto` and `Oracle` are defined in flavor list. Feel free to open issue or send pull request if anyone asks for a new flavor.

By default, all builders uses `DefaultFlavor` to build SQL. The default value is `MySQL`.

There is a `BuildWithFlavor` method in `Builder` interface. We can use it to build a SQL with provided flavor.

We can wrap any `Builder` with a default flavor through `WithFlavor`.

To be more verbose, we can use `PostgreSQL.NewSelectBuilder()` to create a `SelectBuilder` with the `PostgreSQL` flavor. All builders can be created in this way.

### Using `Struct` as a light weight ORM

`Struct` stores type information and struct fields of a struct. It's a factory of builders. We can use `Struct` methods to create initialized SELECT/INSERT/UPDATE/DELETE builders to work with the struct. It can help us to save time and avoid human-error on writing column names.

We can define a struct type and use field tags to let `Struct` know how to create right builders for us.

```go
type ATable struct {
    Field1     string                                    // If a field doesn't has a tag, use "Field1" as column name in SQL.
    Field2     int    `db:"field2"`                      // Use "db" in field tag to set column name used in SQL.
    Field3     int64  `db:"field3" fieldtag:"foo,bar"`   // Set fieldtag to a field. We can call `WithTag` to include fields with tag or `WithoutTag` to exclude fields with tag.
    Field4     int64  `db:"field4" fieldtag:"foo"`       // If we use `s.WithTag("foo").Select("t")`, columnes of SELECT are "t.field3" and "t.field4".
    Field5     string `db:"field5" fieldas:"f5_alias"`   // Use "fieldas" in field tag to set a column alias (AS) used in SELECT.
    Ignored    int32  `db:"-"`                           // If we set field name as "-", Struct will ignore it.
    unexported int                                       // Unexported field is not visible to Struct.
    Quoted     string `db:"quoted" fieldopt:"withquote"` // Add quote to the field using back quote or double quote. See `Flavor#Quote`.
    Empty      uint   `db:"empty" fieldopt:"omitempty"`  // Omit the field in UPDATE if it is a nil or zero value.

    // The `omitempty` can be written as a function.
    // In this case, omit empty field `Tagged` when UPDATE for tag `tag1` and `tag3` but not `tag2`.
    Tagged     string `db:"tagged" fieldopt:"omitempty(tag1,tag3)" fieldtag:"tag1,tag2,tag3"`

    // By default, the `SelectFrom("t")` will add the "t." to all names of fields matched tag.
    // We can add dot to field name to disable this behavior.
    FieldWithTableAlias string `db:"m.field"`
}
```

Read [examples](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#Struct) for `Struct` to learn details of how to use it.

What's more, we can use `Struct` as a kind of zero-config ORM. While most ORM implementations requires several prerequisite configs to work with database connections, `Struct` doesn't require any config and work well with any SQL driver which works with `database/sql`. `Struct` doesn't call any `database/sql` API; It just creates right SQL with arguments for `DB#Query`/`DB#Exec` or a slice of address of struct fields for `Rows#Scan`/`Row#Scan`.

Here is a sample to use `Struct` as ORM. It should be quite straight forward for developers who are familiar with `database/sql` APIs.

```go
type User struct {
    ID     int64  `db:"id" fieldtag:"pk"`
    Name   string `db:"name"`
    Status int    `db:"status"`
}

// A global variable to create SQL builders.
// All methods of userStruct are thread-safe.
var userStruct = NewStruct(new(User))

func ExampleStruct() {
    // Prepare SELECT query.
    //     SELECT user.id, user.name, user.status FROM user WHERE id = 1234
    sb := userStruct.SelectFrom("user")
    sb.Where(sb.Equal("id", 1234))

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
    // SELECT user.id, user.name, user.status FROM user WHERE id = ?
    // [1234]
    // sqlbuilder.User{ID:1234, Name:"huandu", Status:1}
}
```

In many production environments, table column names are usually snake_case words, e.g. `user_id`, while we have to use CamelCase in struct types to make struct fields public and `golint` happy. It's a bit redundant to use the `db` tag in every struct field. If there is a certain rule to map field names to table column names, We can use field mapper function to make code simpler.

The `DefaultFieldMapper` is a global field mapper function to convert field name to new style. By default, it sets to `nil` and does nothing. If we know that most table column names are snake_case words, we can set `DefaultFieldMapper` to `sqlbuilder.SnakeCaseMapper`. If we have some special cases, we can set custom mapper to a `Struct` by calling `WithFieldMapper`.

Following are special notes regarding to field mapper.

- Field tag has precedence over field mapper function - thus, mapper is ignored if the `db` tag is set;
- Field mapper is called only once on a Struct when the Struct is used to create builder for the first time.

See [field mapper function sample](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#FieldMapperFunc) as a demo.

### Nested SQL

It's quite straight forward to create a nested SQL: use a builder as an argument to nest it.

Here is a sample.

```go
sb := sqlbuilder.NewSelectBuilder()
fromSb := sqlbuilder.NewSelectBuilder()
statusSb := sqlbuilder.NewSelectBuilder()

sb.Select("id")
sb.From(sb.BuilderAs(fromSb, "user")))
sb.Where(sb.In("status", statusSb))

fromSb.Select("id").From("user").Where(fromSb.GreaterThan("level", 4))
statusSb.Select("status").From("config").Where(statusSb.Equal("state", 1))

sql, args := sb.Build()
fmt.Println(sql)
fmt.Println(args)

// Output:
// SELECT id FROM (SELECT id FROM user WHERE level > ?) AS user WHERE status IN (SELECT status FROM config WHERE state = ?)
// [4 1]
```

### Use `sql.Named` in a builder

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

### Argument modifiers

There are several modifiers for arguments.

- `List(arg)` represents a list of arguments. If `arg` is a slice or array, e.g. a slice with 3 ints, it will be compiled to `?, ?, ?` and flattened in the final arguments as 3 ints. It's a tool for convenience. We can use it in the `IN` expression or `VALUES` of `INSERT INTO`.
- `TupleNames(names)` and `Tuple(values)` represent the tuple syntax in SQL. See [Tuple](https://pkg.go.dev/github.com/huandu/go-sqlbuilder#example-Tuple) for usage.
- `Named(name, arg)` represents a named argument. It only works with `Build` or `BuildNamed` to define a named placeholder using syntax `${name}`.
- `Raw(expr)` marks an `expr` as a plain string in SQL rather than an argument. When we build a builder, the value of raw expressions are copied in SQL string directly without leaving any `?` in SQL.

### Freestyle builder

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

### Using special syntax to build SQL

Package `sqlbuilder` defines special syntax to represent an uncompiled SQL internally. If we want to take advantage of the syntax to build customized tools, we can use `Build` to compile it with arguments.

The format string uses special syntax to represent arguments.

- `$?` refers successive arguments passed in the call. It works similar as `%v` in `fmt.Sprintf`.
- `$0` `$1` ... `$n` refers nth-argument passed in the call. Next `$?` will use arguments n+1.
- `${name}` refers a named argument created by `Named` with `name`.
- `$$` is a `"$"` string.

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

If we just want to use `${name}` syntax to refer named arguments, use `BuildNamed` instead. It disables all special syntax but `${name}` and `$$`.

### Interpolate `args` in the `sql`

Some SQL-like drivers, e.g. SQL for Redis, SQL for ES, etc., doesn't actually implement `StmtExecContext#ExecContext`. They will fail when `len(args) > 0`. The only solution is to interpolate `args` in the `sql`, and execute the interpolated query with the driver.

The design goal of the interpolation feature in this package is to implement a "basically sufficient" capability, rather than a feature that is on par with various SQL drivers and DBMS systems.

_Security warning_: I try my best to escape special characters in interpolate methods, but it's still less secure than `Stmt` implemented by SQL servers.

This feature is inspired by interpolation feature in package `github.com/go-sql-driver/mysql`.

Here is a sample for MySQL.

```go
sb := MySQL.NewSelectBuilder()
sb.Select("name").From("user").Where(
    sb.NE("id", 1234),
    sb.E("name", "Charmy Liu"),
    sb.Like("desc", "%mother's day%"),
)
sql, args := sb.Build()
query, err := MySQL.Interpolate(sql, args)

fmt.Println(query)
fmt.Println(err)

// Output:
// SELECT name FROM user WHERE id <> 1234 AND name = 'Charmy Liu' AND desc LIKE '%mother\'s day%'
// <nil>
```

Here is a sample for PostgreSQL. Note that the dollar quote is supported.

```go
// Only the last `$1` is interpolated.
// Others are not interpolated as they are inside dollar quote (the `$$`).
query, err := PostgreSQL.Interpolate(`
CREATE FUNCTION dup(in int, out f1 int, out f2 text) AS $$
    SELECT $1, CAST($1 AS text) || ' is text'
$$
LANGUAGE SQL;

SELECT * FROM dup($1);`, []interface{}{42})

fmt.Println(query)
fmt.Println(err)

// Output:
//
// CREATE FUNCTION dup(in int, out f1 int, out f2 text) AS $$
//     SELECT $1, CAST($1 AS text) || ' is text'
// $$
// LANGUAGE SQL;
//
// SELECT * FROM dup(42);
// <nil>
```

## License

This package is licensed under MIT license. See LICENSE for details.
