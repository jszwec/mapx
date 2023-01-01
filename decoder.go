package mapx

import (
	"errors"
	"fmt"
	"reflect"
)

var defaultDecoder = NewDecoder[any](DecoderOpt{})

var mapType = reflect.TypeOf((*map[string]any)(nil)).Elem()

type DecoderOpt struct {
	Converter DecodingConverter
	Tag       string
}

type DecodeError struct {
	Value any
	Type  reflect.Type
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("mapx: cannot decode value of type %T into %s", e.Value, e.Type)
}

type Decoder[T any] struct {
	opt    DecoderOpt
	fields fields
}

func NewDecoder[T any](opts DecoderOpt) *Decoder[T] {
	var fields fields
	typ := reflect.TypeOf((*T)(nil)).Elem()
	if typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.Struct {
		fields = cachedFields(typeKey{
			tag:  defaultTag(opts.Tag),
			Type: typ.Elem(),
		})
	}

	return &Decoder[T]{
		opt:    opts,
		fields: fields,
	}
}

func (dec *Decoder[T]) Decode(m map[string]any, v T) error {
	dst := reflect.ValueOf(v)

	if dst.Kind() == reflect.Pointer {
		dst = dst.Elem()
	}

	if dst.Kind() != reflect.Struct {
		return errors.New("not a struct")
	}

	return dec.decode(m, dst, dec.fields)
}

func (dec *Decoder[T]) decode(m map[string]any, dst reflect.Value, fields fields) error {
	if fields == nil {
		fields = cachedFields(typeKey{
			tag:  defaultTag(dec.opt.Tag),
			Type: dst.Type(),
		})
	}

	for _, f := range fields {
		v, ok := m[f.name]
		if !ok {
			continue
		}

		val := reflect.ValueOf(v)
		typ := val.Type()

		fv := dst.Field(f.index[0])

		if dec.opt.Converter.m != nil {
			if conv, ok := dec.opt.Converter.m[typ]; ok && fv.Type() == conv.dst {
				if err := conv.f(v, fv.Addr().Interface()); err != nil {
					return err
				}
				continue
			}
		}

		switch {
		case typ == f.typ || fastCanConvert(f.typ, typ):
			switch typ.Kind() {
			case reflect.String:
				fv.SetString(val.String())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fv.SetInt(val.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fv.SetUint(val.Uint())
			case reflect.Float64, reflect.Float32:
				fv.SetFloat(val.Float())
			default:
				fv.Set(val)
			}
		case val.CanConvert(f.typ):
			switch f.typ.Kind() {
			case reflect.String:
				fv.SetString(val.Convert(f.typ).String())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fv.SetInt(val.Convert(f.typ).Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fv.SetUint(val.Convert(f.typ).Uint())
			case reflect.Float64, reflect.Float32:
				fv.SetFloat(val.Convert(f.typ).Float())
			default:
				fv.Set(val.Convert(f.typ))
			}
		case f.typ.Kind() == reflect.Slice && val.Type().Kind() == reflect.Slice:
			l := val.Len()
			slice := reflect.MakeSlice(f.typ, l, l)

			for i := 0; i < l; i++ {
				val := val.Index(i)
				if val.Type().Kind() == reflect.Interface {
					val = val.Elem()
				}

				if conv, ok := dec.opt.Converter.m[val.Type()]; ok && f.typ.Elem() == conv.dst {
					if err := conv.f(val.Interface(), slice.Index(i).Addr().Interface()); err != nil {
						return err
					}
					continue
				}

				switch {
				case val.CanConvert(f.typ.Elem()):
					slice.Index(i).Set(val.Convert(f.typ.Elem()))
				case val.Type().ConvertibleTo(mapType):
					if err := dec.decode(val.Interface().(map[string]any), slice.Index(i), f.fields); err != nil {
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
			if err := dec.decode(val.Interface().(map[string]any), fv, f.fields); err != nil {
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

func Decode[T any](m map[string]any, v *T) error {
	return defaultDecoder.Decode(m, v)
}

func fastCanConvert(typ, dst reflect.Type) bool {
	switch typ.Kind() {
	case reflect.String:
		return dst.Kind() == reflect.String
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch dst.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return true
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch dst.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return true
		}
	case reflect.Float32, reflect.Float64:
		switch dst.Kind() {
		case reflect.Float32, reflect.Float64:
			return true
		}
	}
	return false
}
