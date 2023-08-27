package sqlbuilder

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/huandu/go-assert"
)

type errorValuer int

var ErrErrorValuer = errors.New("error valuer")

func (v errorValuer) Value() (driver.Value, error) {
	return 0, ErrErrorValuer
}

func TestFlavorInterpolate(t *testing.T) {
	dt := time.Date(2019, 4, 24, 12, 23, 34, 123456789, time.FixedZone("CST", 8*60*60)) // 2019-04-24 12:23:34.987654321 CST
	_, errOutOfRange := strconv.ParseInt("12345678901234567890", 10, 32)
	byteArr := [...]byte{'f', 'o', 'o'}
	cases := []struct {
		Flavor Flavor
		SQL    string
		Args   []interface{}
		Query  string
		Err    error
	}{
		{
			MySQL,
			"SELECT * FROM a WHERE name = ? AND state IN (?, ?, ?, ?, ?)", []interface{}{"I'm fine", 42, int8(8), int16(-16), int32(32), int64(64)},
			"SELECT * FROM a WHERE name = 'I\\'m fine' AND state IN (42, 8, -16, 32, 64)", nil,
		},
		{
			MySQL,
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN (?, '?', ?, ?, ?, ?, ?)", []interface{}{"\r\n\b\t\x1a\x00\\\"'", uint(42), uint8(8), uint16(16), uint32(32), uint64(64), "useless"},
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN ('\\r\\n\\b\\t\\Z\\0\\\\\\\"\\'', '?', 42, 8, 16, 32, 64)", nil,
		},
		{
			MySQL,
			"SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?", []interface{}{true, false, float32(1.234567), float64(9.87654321), []byte(nil), []byte("I'm bytes"), dt, time.Time{}, nil},
			"SELECT TRUE, FALSE, 1.234567, 9.87654321, NULL, _binary'I\\'m bytes', '2019-04-24 12:23:34.123457', '0000-00-00', NULL", nil,
		},
		{
			MySQL,
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\?", []interface{}{MySQL},
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\'MySQL'", nil,
		},
		{
			MySQL,
			"SELECT ?", []interface{}{byteArr},
			"SELECT _binary'foo'", nil,
		},
		{
			MySQL,
			"SELECT ?", nil,
			"", ErrInterpolateMissingArgs,
		},
		{
			MySQL,
			"SELECT ?", []interface{}{complex(1, 2)},
			"", ErrInterpolateUnsupportedArgs,
		},
		{
			MySQL,
			"SELECT ?", []interface{}{[]complex128{complex(1, 2)}},
			"", ErrInterpolateUnsupportedArgs,
		},
		{
			MySQL,
			"SELECT ?", []interface{}{errorValuer(1)},
			"", ErrErrorValuer,
		},

		{
			PostgreSQL,
			"SELECT * FROM a WHERE name = $3 AND state IN ($2, $4, $1, $6, $5)", []interface{}{"I'm fine", 42, int8(8), int16(-16), int32(32), int64(64)},
			"SELECT * FROM a WHERE name = 8 AND state IN (42, -16, E'I\\'m fine', 64, 32)", nil,
		},
		{
			PostgreSQL,
			"SELECT * FROM $abc$$1$abc$1$1 WHERE name = \"$1\" AND state IN ($2, '$1', $3, $6, $5, $4, $2) $3", []interface{}{"\r\n\b\t\x1a\x00\\\"'", uint(42), uint8(8), uint16(16), uint32(32), uint64(64), "useless"},
			"SELECT * FROM $abc$$1$abc$1E'\\r\\n\\b\\t\\Z\\0\\\\\\\"\\'' WHERE name = \"$1\" AND state IN (42, '$1', 8, 64, 32, 16, 42) 8", nil,
		},
		{
			PostgreSQL,
			"SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $11, $a", []interface{}{true, false, float32(1.234567), float64(9.87654321), []byte(nil), []byte("I'm bytes"), dt, time.Time{}, nil, 10, 11, 12},
			"SELECT TRUE, FALSE, 1.234567, 9.87654321, NULL, E'\\\\x49276D206279746573'::bytea, '2019-04-24 12:23:34.123457 CST', '0000-00-00', NULL, 11, $a", nil,
		},
		{
			PostgreSQL,
			"SELECT '\\'$1', \"\\\"$1\", `$1`, \\$1a, $$1$$, $a $b$ $a $ $1$b$1$1 $a$ $", []interface{}{MySQL},
			"SELECT '\\'$1', \"\\\"$1\", `E'MySQL'`, \\E'MySQL'a, $$1$$, $a $b$ $a $ $1$b$1E'MySQL' $a$ $", nil,
		},
		{
			PostgreSQL,
			"SELECT * FROM a WHERE name = 'Huan''Du''$1' AND desc = $1", []interface{}{"c'mon"},
			"SELECT * FROM a WHERE name = 'Huan''Du''$1' AND desc = E'c\\'mon'", nil,
		},
		{
			PostgreSQL,
			"SELECT $1", nil,
			"", ErrInterpolateMissingArgs,
		},
		{
			PostgreSQL,
			"SELECT $1", []interface{}{complex(1, 2)},
			"", ErrInterpolateUnsupportedArgs,
		},
		{
			PostgreSQL,
			"SELECT $12345678901234567890", nil,
			"", errOutOfRange,
		},

		{
			SQLite,
			"SELECT * FROM a WHERE name = ? AND state IN (?, ?, ?, ?, ?)", []interface{}{"I'm fine", 42, int8(8), int16(-16), int32(32), int64(64)},
			"SELECT * FROM a WHERE name = 'I\\'m fine' AND state IN (42, 8, -16, 32, 64)", nil,
		},
		{
			SQLite,
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN (?, '?', ?, ?, ?, ?, ?)", []interface{}{"\r\n\b\t\x1a\x00\\\"'", uint(42), uint8(8), uint16(16), uint32(32), uint64(64), "useless"},
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN ('\\r\\n\\b\\t\\Z\\0\\\\\\\"\\'', '?', 42, 8, 16, 32, 64)", nil,
		},
		{
			SQLite,
			"SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?", []interface{}{true, false, float32(1.234567), float64(9.87654321), []byte(nil), []byte("I'm bytes"), dt, time.Time{}, nil},
			"SELECT TRUE, FALSE, 1.234567, 9.87654321, NULL, X'49276D206279746573', '2019-04-24 12:23:34.123', '0000-00-00', NULL", nil,
		},
		{
			SQLite,
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\?", []interface{}{SQLite},
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\'SQLite'", nil,
		},

		{
			SQLServer,
			"SELECT * FROM a WHERE name = @p1 AND state IN (@p3, @P2, @p4, @P6, @p5)", []interface{}{"I'm fine", 42, int8(8), int16(-16), int32(32), int64(64)},
			"SELECT * FROM a WHERE name = N'I\\'m fine' AND state IN (8, 42, -16, 64, 32)", nil,
		},
		{
			SQLServer,
			"SELECT * FROM \"a@p1\" WHERE name = '@p1' AND state IN (@p2, '@p1', @p1, @p3, @p4, @p5, @p6)", []interface{}{"\r\n\b\t\x1a\x00\\\"'", uint(42), uint8(8), uint16(16), uint32(32), uint64(64), "useless"},
			"SELECT * FROM \"a@p1\" WHERE name = '@p1' AND state IN (42, '@p1', N'\\r\\n\\b\\t\\Z\\0\\\\\\\"\\'', 8, 16, 32, 64)", nil,
		},
		{
			SQLServer,
			"SELECT @p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9", []interface{}{true, false, float32(1.234567), float64(9.87654321), []byte(nil), []byte("I'm bytes"), dt, time.Time{}, nil},
			"SELECT TRUE, FALSE, 1.234567, 9.87654321, NULL, 0x49276D206279746573, '2019-04-24 12:23:34.123457 +08:00', '0000-00-00', NULL", nil,
		},
		{
			SQLServer,
			"SELECT '\\'@p1', \"\\\"@p1\", \\@p1, @abc", []interface{}{SQLServer},
			"SELECT '\\'@p1', \"\\\"@p1\", \\N'SQLServer', @abc", nil,
		},
		{
			SQLServer,
			"SELECT @p1", nil,
			"", ErrInterpolateMissingArgs,
		},
		{
			CQL,
			"SELECT * FROM a WHERE name = ? AND state IN (?, ?, ?, ?, ?)", []interface{}{"I'm fine", 42, int8(8), int16(-16), int32(32), int64(64)},
			"SELECT * FROM a WHERE name = 'I''m fine' AND state IN (42, 8, -16, 32, 64)", nil,
		},
		{
			CQL,
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN (?, '?', ?, ?, ?, ?, ?)", []interface{}{"\r\n\b\t\x1a\x00\\\"'", uint(42), uint8(8), uint16(16), uint32(32), uint64(64), "useless"},
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN ('\\r\\n\\b\\t\\Z\\0\\\\\\\"''', '?', 42, 8, 16, 32, 64)", nil,
		},
		{
			CQL,
			"SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?", []interface{}{true, false, float32(1.234567), float64(9.87654321), []byte(nil), []byte("I'm bytes"), dt, time.Time{}, nil},
			"SELECT TRUE, FALSE, 1.234567, 9.87654321, NULL, 0x49276D206279746573, '2019-04-24 12:23:34.123457+0800', '0000-00-00', NULL", nil,
		},
		{
			CQL,
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\?", []interface{}{CQL},
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\'CQL'", nil,
		},
		{
			CQL,
			"SELECT ?", nil,
			"", ErrInterpolateMissingArgs,
		},
		{
			CQL,
			"SELECT ?", []interface{}{complex(1, 2)},
			"", ErrInterpolateUnsupportedArgs,
		},
		{
			ClickHouse,
			"SELECT * FROM a WHERE name = ? AND state IN (?, ?, ?, ?, ?)", []interface{}{"I'm fine", 42, int8(8), int16(-16), int32(32), int64(64)},
			"SELECT * FROM a WHERE name = 'I\\'m fine' AND state IN (42, 8, -16, 32, 64)", nil,
		},
		{
			ClickHouse,
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN (?, '?', ?, ?, ?, ?, ?)", []interface{}{"\r\n\b\t\x1a\x00\\\"'", uint(42), uint8(8), uint16(16), uint32(32), uint64(64), "useless"},
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN ('\\r\\n\\b\\t\\Z\\0\\\\\\\"\\'', '?', 42, 8, 16, 32, 64)", nil,
		},
		{
			ClickHouse,
			"SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?", []interface{}{true, false, float32(1.234567), 9.87654321, []byte(nil), []byte("I'm bytes"), dt, time.Time{}, nil},
			"SELECT TRUE, FALSE, 1.234567, 9.87654321, NULL, unhex('49276D206279746573'), '2019-04-24 12:23:34.123457', '0000-00-00', NULL", nil,
		},
		{
			ClickHouse,
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\?", []interface{}{MySQL},
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\'MySQL'", nil,
		},
		{
			ClickHouse,
			"SELECT ?", []interface{}{byteArr},
			"SELECT unhex('666F6F')", nil,
		},
		{
			ClickHouse,
			"SELECT ?", nil,
			"", ErrInterpolateMissingArgs,
		},
		{
			ClickHouse,
			"SELECT ?", []interface{}{complex(1, 2)},
			"", ErrInterpolateUnsupportedArgs,
		},
		{
			ClickHouse,
			"SELECT ?", []interface{}{[]complex128{complex(1, 2)}},
			"", ErrInterpolateUnsupportedArgs,
		},
		{
			ClickHouse,
			"SELECT ?", []interface{}{errorValuer(1)},
			"", ErrErrorValuer,
		},
		{
			Presto,
			"SELECT * FROM a WHERE name = ? AND state IN (?, ?, ?, ?, ?)", []interface{}{"I'm fine", 42, int8(8), int16(-16), int32(32), int64(64)},
			"SELECT * FROM a WHERE name = 'I\\'m fine' AND state IN (42, 8, -16, 32, 64)", nil,
		},
		{
			Presto,
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN (?, '?', ?, ?, ?, ?, ?)", []interface{}{"\r\n\b\t\x1a\x00\\\"'", uint(42), uint8(8), uint16(16), uint32(32), uint64(64), "useless"},
			"SELECT * FROM `a?` WHERE name = \"?\" AND state IN ('\\r\\n\\b\\t\\Z\\0\\\\\\\"\\'', '?', 42, 8, 16, 32, 64)", nil,
		},
		{
			Presto,
			"SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?", []interface{}{true, false, float32(1.234567), 9.87654321, []byte(nil), []byte("I'm bytes"), dt, time.Time{}, nil},
			"SELECT TRUE, FALSE, 1.234567, 9.87654321, NULL, from_hex('49276D206279746573'), '2019-04-24 12:23:34.123', '0000-00-00', NULL", nil,
		},
		{
			Presto,
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\?", []interface{}{MySQL},
			"SELECT '\\'?', \"\\\"?\", `\\`?`, \\'MySQL'", nil,
		},
		{
			Presto,
			"SELECT ?", []interface{}{byteArr},
			"SELECT from_hex('666F6F')", nil,
		},
		{
			Presto,
			"SELECT ?", nil,
			"", ErrInterpolateMissingArgs,
		},
		{
			Presto,
			"SELECT ?", []interface{}{complex(1, 2)},
			"", ErrInterpolateUnsupportedArgs,
		},
		{
			Presto,
			"SELECT ?", []interface{}{[]complex128{complex(1, 2)}},
			"", ErrInterpolateUnsupportedArgs,
		},
		{
			Presto,
			"SELECT ?", []interface{}{errorValuer(1)},
			"", ErrErrorValuer,
		},

		{
			Oracle,
			"SELECT * FROM a WHERE name = :3 AND state IN (:2, :4, :1, :6, :5)", []interface{}{"I'm fine", 42, int8(8), int16(-16), int32(32), int64(64)},
			"SELECT * FROM a WHERE name = 8 AND state IN (42, -16, 'I\\'m fine', 64, 32)", nil,
		},
		{
			Oracle,
			"SELECT * FROM :abc::1:abc:1:1 WHERE name = \":1\" AND state IN (:2, ':1', :3, :6, :5, :4, :2) :3", []interface{}{"\r\n\b\t\x1a\x00\\\"'", uint(42), uint8(8), uint16(16), uint32(32), uint64(64), "useless"},
			"SELECT * FROM :abc::1:abc:1'\\r\\n\\b\\t\\Z\\0\\\\\\\"\\'' WHERE name = \":1\" AND state IN (42, ':1', 8, 64, 32, 16, 42) 8", nil,
		},
		{
			Oracle,
			"SELECT :1, :2, :3, :4, :5, :6, :7, :8, :9, :11, :a", []interface{}{true, false, float32(1.234567), float64(9.87654321), []byte(nil), []byte("I'm bytes"), dt, time.Time{}, nil, 10, 11, 12},
			"SELECT 1, 0, 1.234567, 9.87654321, NULL, hextoraw('49276D206279746573'), to_timestamp('2019-04-24 12:23:34.123457', 'YYYY-MM-DD HH24:MI:SS.FF'), '0000-00-00', NULL, 11, :a", nil,
		},
		{
			Oracle,
			"SELECT '\\':1', \"\\\":1\", `:1`, \\:1a, ::1::, :a :b: :a : :1:b:1:1 :a: :", []interface{}{Oracle},
			"SELECT '\\':1', \"\\\":1\", `'Oracle'`, \\'Oracle'a, ::1::, :a :b: :a : :1:b:1'Oracle' :a: :", nil,
		},
		{
			Oracle,
			"SELECT * FROM a WHERE name = 'Huan''Du'':1' AND desc = :1", []interface{}{"c'mon"},
			"SELECT * FROM a WHERE name = 'Huan''Du'':1' AND desc = 'c\\'mon'", nil,
		},
		{
			Oracle,
			"SELECT :1", nil,
			"", ErrInterpolateMissingArgs,
		},
		{
			Oracle,
			"SELECT :1", []interface{}{complex(1, 2)},
			"", ErrInterpolateUnsupportedArgs,
		},
		{
			Oracle,
			"SELECT :12345678901234567890", nil,
			"", errOutOfRange,
		},
	}

	for idx, c := range cases {
		t.Run(fmt.Sprintf("%s: %s", c.Flavor.String(), c.Query), func(t *testing.T) {
			a := assert.New(t)
			a.Use(&idx, &c)
			query, err := c.Flavor.Interpolate(c.SQL, c.Args)

			a.Equal(query, c.Query)
			a.Assert(err == c.Err || err.Error() == c.Err.Error())
		})
	}
}
