package humaext

import (
	"reflect"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type UUIDs struct {
	Raw []string
	IDs uuid.UUIDs
}

func (u UUIDs) Schema(r huma.Registry) *huma.Schema {
	uuidType := huma.SchemaFromType(r, reflect.TypeFor[[]uuid.UUID]())
	uuidType.Items = huma.SchemaFromType(r, reflect.TypeFor[uuid.UUID]())
	uuidType.Items.Format = "uuid"
	return uuidType
}

func (u *UUIDs) Receiver() reflect.Value {
	return reflect.ValueOf(u).Elem().Field(0)
}

func (u *UUIDs) OnParamSet(isSet bool, parsed any) {
	items, ok := parsed.([]string)
	if !ok {
		return
	}
	u.IDs = make(uuid.UUIDs, len(items))

	for i, raw := range items {
		uid, err := uuid.Parse(raw)
		if err != nil {
			continue
		}

		u.IDs[i] = uid
	}
}
