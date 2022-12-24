package mapx

import (
	"fmt"
	"reflect"
)

var mapType = reflect.TypeOf((*map[string]any)(nil)).Elem()

type DecodeError struct {
	Value any
	Type  reflect.Type
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("mapx: cannot decode value of type %T into %s", e.Value, e.Type)
}

func Decode(m map[string]any, v any) error {
	dst := reflect.ValueOf(v)

	if dst.Kind() == reflect.Pointer {
		dst = dst.Elem()
	}

	return decode(m, dst)
}

func decode(m map[string]any, dst reflect.Value) error {
	fields := cachedFields(typeKey{
		tag:  "mapx",
		Type: dst.Type(),
	})

	for _, f := range fields {
		v, ok := m[f.name]
		if !ok {
			continue
		}

		val := reflect.ValueOf(v)
		typ := val.Type()

		fv := dst.Field(f.index[0])
		switch {
		case typ == f.typ:
			fv.Set(val)
		case val.CanConvert(f.typ):
			fv.Set(val.Convert(f.typ))
		case f.typ.Kind() == reflect.Slice && val.Type().Kind() == reflect.Slice:
			l := val.Len()
			slice := reflect.MakeSlice(f.typ, l, l)

			for i := 0; i < l; i++ {
				val := val.Index(i)
				if val.Type().Kind() == reflect.Interface {
					val = val.Elem()
				}

				switch {
				case val.CanConvert(f.typ.Elem()):
					slice.Index(i).Set(val.Convert(f.typ.Elem()))
				case val.Type().ConvertibleTo(mapType):
					if err := decode(val.Interface().(map[string]any), slice.Index(i)); err != nil {
						return err
					}
				default:
					return &DecodeError{
						Value: val.Interface(),
						Type:  f.typ.Elem(),
					}
				}

			}
			fv.Set(slice)
		case f.typ.Kind() == reflect.Struct && typ.ConvertibleTo(mapType):
			if err := decode(val.Interface().(map[string]any), fv); err != nil {
				return err
			}
		default:
			return &DecodeError{
				Value: v,
				Type:  fv.Type(),
			}
		}
	}

	return nil
}
