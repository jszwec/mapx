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

// Error implements error interface.
func (e *DecodeError) Error() string {
	return fmt.Sprintf("mapx: cannot decode value of type %T into %s", e.Value, e.Type)
}

// Is implements errors.Is interface.
func (e *DecodeError) Is(err error) bool {
	var derr *DecodeError
	return err != nil &&
		e != nil &&
		errors.As(err, &derr) &&
		derr.Type == e.Type &&
		derr.Value == e.Value
}

type Decoder[T any] struct {
	opt    DecoderOpt
	fields fields
}

func NewDecoder[T any](opts DecoderOpt) *Decoder[T] {
	return &Decoder[T]{
		opt:    opts,
		fields: structFields[T](opts.Tag),
	}
}

func (dec *Decoder[T]) Decode(m map[string]any, v T) error {
	dst := reflect.ValueOf(v)

	if dst.Kind() != reflect.Pointer {
		return ErrNotAPointer
	}

	dst = dst.Elem()
	if dst.Kind() != reflect.Struct {
		return ErrNotAStruct
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

loop:
	for _, f := range fields {
		v, ok := m[f.name]
		if !ok {
			continue
		}

		val := reflect.ValueOf(v)
		if !val.IsValid() {
			if f.baseType.Kind() != reflect.Interface && f.baseType.Kind() != reflect.Pointer {
				return &DecodeError{
					Value: v,
					Type:  f.baseType,
				}
			}
			continue
		}

		typ := val.Type()

		fv := dst.Field(f.index[0])

		switch {
		case dec.opt.Converter.m != nil:
			if conv, ok := dec.opt.Converter.m[typ]; ok && reflect.PointerTo(fv.Type()) == conv.dst {
				if err := conv.f(v, fv.Addr().Interface()); err != nil {
					return err
				}
				continue
			}
		case dec.opt.Converter.ifaceFuncs != nil:
			for _, fn := range dec.opt.Converter.ifaceFuncs[typ] {
				if f.typ.AssignableTo(fn.typ) {
					if err := fn.f(v, fv.Interface()); err != nil {
						return err
					}
					continue loop
				}
				if reflect.PtrTo(f.typ).AssignableTo(fn.typ) {
					if err := fn.f(v, fv.Addr().Interface()); err != nil {
						return err
					}
					continue loop
				}
			}
		}

		if f.baseType.Kind() == reflect.Pointer && typ.Kind() != reflect.Pointer && fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
			fv = fv.Elem()
		}

		switch {
		case typ == fv.Type() || fastCanConvert(fv.Type(), typ):
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
		case val.CanConvert(fv.Type()):
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
				fv.Set(val.Convert(fv.Type()))
			}
		case fv.Type().Kind() == reflect.Slice && val.Type().Kind() == reflect.Slice:
			l := val.Len()
			slice := reflect.MakeSlice(fv.Type(), l, l)

			shouldInit := fv.Type().Elem().Kind() == reflect.Pointer

		sliceLoop:
			for i := 0; i < l; i++ {
				val := val.Index(i)
				if val.Type().Kind() == reflect.Interface {
					val = val.Elem()
				}

				if !val.IsValid() {
					if typ, k := fv.Type().Elem(), fv.Type().Elem().Kind(); k != reflect.Interface && k != reflect.Pointer {
						return &DecodeError{
							Value: nil,
							Type:  typ,
						}
					}
					continue
				}

				dst := slice.Index(i)
				if shouldInit {
					dst.Set(reflect.New(fv.Type().Elem().Elem()))
				}

				switch {
				case dec.opt.Converter.m != nil:
					if conv, ok := dec.opt.Converter.m[val.Type()]; ok && reflect.PointerTo(f.typ.Elem()) == conv.dst {
						if err := conv.f(val.Interface(), dst.Addr().Interface()); err != nil {
							return err
						}
						continue
					}
				case dec.opt.Converter.ifaceFuncs != nil:
					for _, fn := range dec.opt.Converter.ifaceFuncs[val.Type()] {
						if f.typ.Elem().AssignableTo(fn.typ) {
							if err := fn.f(val.Interface(), dst.Interface()); err != nil {
								return err
							}
							continue sliceLoop
						}

						if reflect.PtrTo(f.typ.Elem()).AssignableTo(fn.typ) {
							if err := fn.f(val.Interface(), dst.Addr().Interface()); err != nil {
								return err
							}
							continue sliceLoop
						}
					}
				}

				switch {
				case val.CanConvert(fv.Type().Elem()):
					dst.Set(val.Convert(fv.Type().Elem()))
				case shouldInit && val.CanConvert(fv.Type().Elem().Elem()):
					dst.Elem().Set(val.Convert(dst.Type().Elem()))
				case val.Type().ConvertibleTo(mapType):
					if err := dec.decode(val.Interface().(map[string]any), dst, f.fields); err != nil {
						return err
					}
				default:
					return &DecodeError{
						Value: val.Interface(),
						Type:  f.baseType.Elem(),
					}
				}
			}
			fv.Set(slice)
		case fv.Type().Kind() == reflect.Struct && typ.ConvertibleTo(mapType):
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
