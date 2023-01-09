package mapx

import (
	"errors"
	"reflect"
)

var (
	ErrNotAStruct  = errors.New("mapx: provided value is not a struct")
	ErrNotAPointer = errors.New("mapx: provided value is not a pointer")
)

func defaultTag(s string) string {
	if s == "" {
		return "mapx"
	}
	return s
}

func structFields[T any](tag string) fields {
	typ := walkType(reflect.TypeOf((*T)(nil)).Elem())
	if typ.Kind() == reflect.Struct {
		return cachedFields(typeKey{
			tag:  defaultTag(tag),
			Type: typ,
		})
	}
	return nil
}

func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	if len(index) == 1 {
		return v.Field(index[0])
	}

	for i, x := range index {
		if i > 0 {
			if v.Kind() == reflect.Pointer && v.Type().Elem().Kind() == reflect.Struct {
				if v.IsNil() {
					return reflect.Value{}
				}
				v = v.Elem()
			}
		}
		v = v.Field(x)
	}
	return v
}
