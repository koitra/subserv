package cursor

import (
	"time"

	"github.com/koitra/subserv/internal/humaext"
)

type WebCursorIn struct {
	At    humaext.Opt[time.Time] `query:"cursor,omitempty"`
	Limit humaext.Opt[int]       `query:"limit,omitempty"`
}

func (c WebCursorIn) ToCursor() Cursor {
	var limit int
	if c.Limit.IsSet && c.Limit.Value > 0 {
		limit = c.Limit.Value
	}

	var at *time.Time
	if c.At.IsSet {
		at = &c.At.Value
	}

	return Cursor{
		At:    at,
		Limit: limit,
	}
}
