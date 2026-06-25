package dba

import (
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/mods"
)

type (
	WithIn[Q psql.Filterable, I any] interface {
		In(slice ...I) mods.Where[Q]
	}
)

func InIfNotEmpty[Q psql.Filterable, I any](
	column WithIn[Q, I],
	vals ...I,
) bob.ModFunc[Q] {
	return func(sq Q) {
		if len(vals) == 0 {
			return
		}

		column.In(vals...).Apply(sq)
	}
}
