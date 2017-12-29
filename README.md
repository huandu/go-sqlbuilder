# SQL builder for Go #

[![Build Status](https://travis-ci.org/huandu/go-sqlbuilder.svg?branch=master)](https://travis-ci.org/huandu/go-sqlbuilder)
[![GoDoc](https://godoc.org/github.com/huandu/go-sqlbuilder?status.svg)](https://godoc.org/github.com/huandu/go-sqlbuilder)

Package `sqlbuilder` provides a set of flexible and powerful SQL string builders. The only goal of this package is to build SQL string with arguments which can be used in `DB#Query` defined in package `database/sql`.

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

* [SelectBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#SelectBuilder)
* [InsertBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#InsertBuilder)
* [UpdateBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#UpdateBuilder)
* [DeleteBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#DeleteBuilder)
* [Build](https://godoc.org/github.com/huandu/go-sqlbuilder#Build)
* [Buildf](https://godoc.org/github.com/huandu/go-sqlbuilder#Buildf)

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

### Using a special syntax to build SQL ###

Package `sqlbuilder` defines a special syntax to represent an uncompiled SQL internally. If we want to take advantage of the syntax to build customized tools, we can use `Build` to compile it with arguments.

The format string uses a special syntax to represent arguments.

* `$?` uses successive arguments passed in the call. It works similar as `%v` in `fmt.Sprintf`.
* `$0` `$1` ... `$n` uses nth-argument passed in the call. Next `$?` will use arguments n+1.
* `${name}` uses a named argument created by `Named` with `name`.
* `$$` represents a `"$"` string.

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