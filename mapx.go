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
