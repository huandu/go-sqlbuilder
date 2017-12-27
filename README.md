# SQL builder for Go #

[![Build Status](https://travis-ci.org/huandu/go-sqlbuilder.svg?branch=master)](https://travis-ci.org/huandu/go-sqlbuilder)
[![GoDoc](https://godoc.org/github.com/huandu/go-sqlbuilder?status.svg)](https://godoc.org/github.com/huandu/go-sqlbuilder)

Package `go-sqlbuilder` provides a set of flexible and powerful SQL string builders. The only goal of this package is to build SQL string with arguments which can be used in `DB#Query` defined in package `database/sql`.

## Install ##

Use `go get` to install this package.

    go get -u github.com/huandu/go-sqlbuilder

## Usage ##

Following builders are implemented right now. API document and examples are provided in the `godoc` document.

* [SelectBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#SelectBuilder)
* [InsertBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#InsertBuilder)
* [UpdateBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#UpdateBuilder)
* [DeleteBuilder](https://godoc.org/github.com/huandu/go-sqlbuilder#DeleteBuilder)

## License ##

This package is licensed under MIT license. See LICENSE for details.