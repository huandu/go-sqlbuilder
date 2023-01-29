package benchmark

import (
	"testing"

	"github.com/huandu/go-sqlbuilder"
)

func Benchmark_SelectBasic_SQLBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sb := sqlbuilder.Select("*").
			From("users").
			Join("emails USING (email_id)")
		sb.Where(sb.IsNull("deleted_at"))
		sb.Build()
	}
}

func Benchmark_SelectComplex_SQLBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sb := sqlbuilder.Select("id", "name")
		innerSb := sqlbuilder.Select("*").From("banned")
		innerSb.Where(
			innerSb.NotIn("name", sqlbuilder.Flatten([]string{"Huan Du", "Charmy Liu"})...),
		)

		sb.From(
			sb.BuilderAs(innerSb, "user"),
		)
		sb.Where(
			sb.In("status", sqlbuilder.Flatten([]int{1, 2, 3})...),
			sb.Between("created_at", 1234567890, 1234599999),
		)
		sb.OrderBy("modified_at").Desc()

		sb.Build()
	}
}

func Benchmark_Update_SQLBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ub := sqlbuilder.Update("demo.user")
		ub.Set(
			ub.Assign("type", "sys"),
			ub.Incr("credit"),
			"modified_at = UNIX_TIMESTAMP(NOW())",
		)
		ub.Where(
			ub.GreaterThan("id", 1234),
			ub.Like("name", "%Du"),
			ub.Or(
				ub.IsNull("id_card"),
				ub.In("status", 1, 2, 5),
			),
			"modified_at > created_at + "+ub.Var(86400),
		)
		ub.OrderBy("id").Asc()

		ub.Build()
	}
}

func Benchmark_Delete_SqlBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		del := sqlbuilder.DeleteFrom("demo.user")
		del.Where(
			del.GreaterThan("id", 1234),
			del.Like("name", "%Du"),
			del.Or(
				del.IsNull("id_card"),
				del.In("status", 1, 2, 5),
			),
			"modified_at > created_at + "+del.Var(86400),
		)

		del.Build()
	}
}
