package cursor

import (
	"time"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/mods"
)

type (
	Cursor struct {
		At    *time.Time
		Limit int
		Order Order
	}
	Order int

	CursorQueryMod[T bob.Expression] struct {
		at     *time.Time
		limit  int
		order  Order
		column Column[T]
	}

	Column[T bob.Expression] interface {
		GT(bob.Expression) T
		LT(bob.Expression) T
	}

	CursorExpression = func(time.Time) mods.Where[*dialect.SelectQuery]
)

func (c Cursor) WithMaxLimit(limit int) Cursor {
	if c.Limit <= 0 || c.Limit > limit {
		c.Limit = limit
	}
	return c
}

func SelectMod[T bob.Expression](c Cursor, column Column[T]) CursorQueryMod[T] {
	return CursorQueryMod[T]{
		at:     c.At,
		limit:  c.Limit,
		order:  c.Order,
		column: column,
	}
}

func (c CursorQueryMod[T]) Apply(q *dialect.SelectQuery) {
	switch c.order {
	case Desc:
		sm.OrderBy(c.column).Desc().Apply(q)
		if c.at != nil {
			sm.Where(c.column.LT(psql.Arg(*c.at))).Apply(q)
		}
	case Asc:
		sm.OrderBy(c.column).Asc().Apply(q)
		if c.at != nil {
			sm.Where(c.column.GT(psql.Arg(*c.at))).Apply(q)
		}
	}

	if c.limit > 0 {
		sm.Limit(c.limit).Apply(q)
	}
}

const (
	Desc Order = iota
	Asc
)
