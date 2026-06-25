package subscriptions

import (
	"fmt"
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type Date time.Time

func (d *Date) UnmarshalText(text []byte) error {
	t, err := parseDate(string(text))
	if err != nil {
		return err
	}
	*d = Date(t)
	return nil
}

func (d Date) MarshalText() ([]byte, error) {
	text := (time.Time(d)).Format("01-2006")
	return []byte(text), nil
}

func (d Date) Schema(r huma.Registry) *huma.Schema {
	s := huma.SchemaFromType(r, reflect.TypeFor[string]())
	s.Format = "month-year"
	s.Description = "Month and year (MM-YYYY)"
	s.Examples = []any{"02-2026"}
	return s
}

func parseDate(s string) (time.Time, error) {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, IndalidDateError{Date: s, Inner: err}
	}

	return t, nil
}

type IndalidDateError struct {
	Date  string
	Inner error
}

func (e IndalidDateError) Error() string {
	return fmt.Sprintf("date %s is invalid: %s", e.Date, e.Inner.Error())
}
