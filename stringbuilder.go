// Copyright 2023 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"io"
	"strings"
)

type stringBuilder struct {
	builder *strings.Builder
}

var _ io.Writer = new(stringBuilder)

func newStringBuilder() *stringBuilder {
	return &stringBuilder{
		builder: &strings.Builder{},
	}
}

// WriteLeadingString writes s to internal buffer.
// If it's not the first time to write the string, a blank (" ") will be written before s.
func (sb *stringBuilder) WriteLeadingString(s string) {
	if sb.builder.Len() > 0 {
		sb.builder.WriteString(" ")
	}

	sb.builder.WriteString(s)
}

func (sb *stringBuilder) WriteString(s string) {
	sb.builder.WriteString(s)
}

func (sb *stringBuilder) WriteRune(r rune) {
	sb.builder.WriteRune(r)
}

func (sb *stringBuilder) Write(data []byte) (int, error) {
	return sb.builder.Write(data)
}

func (sb *stringBuilder) String() string {
	return sb.builder.String()
}

func (sb *stringBuilder) Reset() {
	sb.builder.Reset()
}
