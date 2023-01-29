package benchmark

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func Benchmark_SelectBasic_Squirrel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		users := sq.Select("*").
			From("users").
			Join("emails USING (email_id)").
			Where(sq.Eq{"deleted_at": nil})
		users.ToSql()
	}
}

func Benchmark_SelectComplex_Squirrel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		innerSb := sq.Select("*").
			From("banned").
			Where("name NOT IN (?, ?)", "Huan Du", "Charmy Liu")
		sb := sq.Select("id", "name").
			FromSelect(innerSb, "user").
			Where("status IN (?, ?, ?)", 1, 2, 3).
			Where("created_at BETWEEN ? AND ?", 1234567890, 1234599999)

		sb.ToSql()
	}
}

func Benchmark_Update_Squirrel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ub := sq.Update("demo.user").
			Set("type", "sys").
			Set("credit", sq.Expr("credit + 1")).
			Set("modified_at", sq.Expr("UNIX_TIMESTAMP(NOW())")).
			Where(sq.Gt{"id": 1234}).
			Where(sq.Like{"name": "%Du"}).
			Where(sq.Or{
				sq.Eq{"id_card": nil},
				sq.Expr("status IN (?, ?, ?)", 1, 2, 5),
			}).
			Where(sq.Expr("modified_at > created_at + ?", 86400)).
			OrderBy("id ASC")

		ub.ToSql()
	}
}

func Benchmark_Delete_Squirrel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		del := sq.Delete("demo.user").
			Where(sq.Gt{"id": 1234}).
			Where(sq.Like{"name": "%Du"}).
			Where(sq.Or{
				sq.Eq{"id_card": nil},
				sq.Expr("status IN (?, ?, ?)", 1, 2, 5),
			}).
			Where(sq.Expr("modified_at > created_at + ?", 86400))

		del.ToSql()
	}
}
