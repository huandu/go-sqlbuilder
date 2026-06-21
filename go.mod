module github.com/huandu/go-sqlbuilder

go 1.18

require (
	github.com/huandu/go-assert v1.1.6
	github.com/huandu/go-clone v1.7.3
	github.com/huandu/xstrings v1.4.0
)

require github.com/davecgh/go-spew v1.1.1 // indirect

// v1.42.0 was published with an incorrect module path
// (github.com/webdaad/go-sqlbuilder). Do not use.
retract v1.42.0
