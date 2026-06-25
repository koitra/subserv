package humaext

import (
	"reflect"

	"github.com/danielgtaylor/huma/v2"
)

type Opt[T any] struct {
	Value T
	IsSet bool
}

func (o Opt[T]) Schema(r huma.Registry) *huma.Schema {
	s := huma.SchemaFromType(r, reflect.TypeOf(o.Value))
	s.Nullable = true
	return s
}

func (o *Opt[T]) Receiver() reflect.Value {
	return reflect.ValueOf(o).Elem().Field(0)
}

func (o *Opt[T]) OnParamSet(isSet bool, parsed any) {
	o.IsSet = isSet
}

func (o *Opt[T]) Ptr() *T {
	if !o.IsSet {
		return nil
	}
	return &o.Value
}
