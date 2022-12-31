package mapx

import (
	"fmt"
	"reflect"
)

var mapType = reflect.TypeOf((*map[string]any)(nil)).Elem()

type DecoderOpt func(*Decoder)

type DecodeError struct {
	Value any
	Type  reflect.Type
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("mapx: cannot decode value of type %T into %s", e.Value, e.Type)
}

type Decoder struct {
	converter DecodingConverter
	tag       string
}

func NewDecoder(opts ...DecoderOpt) *Decoder {
	var dec Decoder
	for _, opt := range opts {
		opt(&dec)
	}
	return &dec
}

func (dec *Decoder) Decode(m map[string]any, v any) error {
	dst := reflect.ValueOf(v)

	if dst.Kind() == reflect.Pointer {
		dst = dst.Elem()
	}

	return dec.decode(m, dst)
}

func (e *Decoder) WithConverter(c DecodingConverter) { e.converter = c }
func (e *Decoder) WithTag(s string)                  { e.tag = s }

func (dec *Decoder) decode(m map[string]any, dst reflect.Value) error {
	fields := cachedFields(typeKey{
		tag:  defaultTag(dec.tag),
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

		if dec.converter.m != nil {
			if conv, ok := dec.converter.m[typ]; ok && fv.Type() == conv.dst {
				if err := conv.f(v, fv.Addr().Interface()); err != nil {
					return err
				}
				continue
			}
		}

		switch {
		case typ == f.typ:
			switch typ.Kind() {
			case reflect.String:
				fv.SetString(val.String())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fv.SetInt(val.Int())
			case reflect.Float64, reflect.Float32:
				fv.SetFloat(val.Float())
			default:
				fv.Set(val)
			}
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

				if conv, ok := dec.converter.m[val.Type()]; ok && f.typ.Elem() == conv.dst {
					if err := conv.f(val.Interface(), slice.Index(i).Addr().Interface()); err != nil {
						return err
					}
					continue
				}

				switch {
				case val.CanConvert(f.typ.Elem()):
					slice.Index(i).Set(val.Convert(f.typ.Elem()))
				case val.Type().ConvertibleTo(mapType):
					if err := dec.decode(val.Interface().(map[string]any), slice.Index(i)); err != nil {
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
			if err := dec.decode(val.Interface().(map[string]any), fv); err != nil {
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

func Decode(m map[string]any, v any, opts ...DecoderOpt) error {
	return NewDecoder(opts...).Decode(m, v)
}
