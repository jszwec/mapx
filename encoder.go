package mapx

import (
	"errors"
	"reflect"
)

var defaultEncoder = NewEncoder[any](EncoderOpt{})

type EncoderOpt struct {
	EncoderFuncs EncoderFuncs
	Tag          string
}

type Encoder[T any] struct {
	opts   EncoderOpt
	fields fields
}

func NewEncoder[T any](opts EncoderOpt) *Encoder[T] {
	return &Encoder[T]{
		opts:   opts,
		fields: structFields[T](opts.Tag),
	}
}

func (e *Encoder[T]) Encode(val T) (map[string]any, error) {
	return e.encode(reflect.ValueOf(val), e.fields)
}

func (e *Encoder[T]) encode(v reflect.Value, fields fields) (_ map[string]any, err error) {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, errors.New("not a struct")
	}

	if fields == nil {
		fields = cachedFields(typeKey{
			tag:  defaultTag(e.opts.Tag),
			Type: v.Type(),
		})
	}

	m := make(map[string]any, len(fields))
	for _, f := range fields {
		if f.typ.Kind() == reflect.Struct {
			sub, err := e.encode(v.Field(f.index[0]), f.fields)
			if err != nil {
				return nil, err
			}
			m[f.name] = sub
			continue
		}

		if e.opts.EncoderFuncs.m != nil {
			if fn, ok := e.opts.EncoderFuncs.m[f.typ]; ok {
				m[f.name], err = fn(v.Field(f.index[0]).Interface())
				if err != nil {
					return nil, err
				}
				continue
			}
		}

		fv := v.Field(f.index[0]).Interface()

		if e.opts.EncoderFuncs.anyConv != nil {
			v, err := e.opts.EncoderFuncs.anyConv(fv)
			if err != nil {
				return nil, err
			}

			switch v {
			case SkipValue{}:
				continue
			case NoChange{}:
			default:
				fv = v
			}
		}

		m[f.name] = fv
	}

	return m, nil
}

func Encode[T any](val T) (map[string]any, error) {
	return defaultEncoder.Encode(val)
}

type (
	SkipValue struct{}
	NoChange  struct{}
)

var anyConvFuncType = reflect.TypeOf((func(any) (any, error))(nil))

type EncoderFuncs struct {
	anyConv    func(any) (any, error)
	m          map[reflect.Type]func(any) (any, error)
	ifaceFuncs map[reflect.Type][]func(any) (any, error)
}

func (ef EncoderFuncs) clone() EncoderFuncs {
	var (
		m          map[reflect.Type]func(any) (any, error)
		ifaceFuncs map[reflect.Type][]func(any) (any, error)
	)

	if ef.m != nil {
		m = make(map[reflect.Type]func(any) (any, error), len(ef.m)+1)
		for k, v := range ef.m {
			m[k] = v
		}
	}

	if ef.ifaceFuncs != nil {
		ifaceFuncs = make(map[reflect.Type][]func(any) (any, error), len(ef.ifaceFuncs)+1)
		for k, v := range ef.ifaceFuncs {
			cp := make([]func(any) (any, error), len(v))
			copy(cp, v)
			ifaceFuncs[k] = v
		}
	}

	return EncoderFuncs{
		anyConv:    ef.anyConv,
		m:          m,
		ifaceFuncs: ifaceFuncs,
	}
}

func RegisterEncoder[T, V any](ef EncoderFuncs, f func(T) (V, error)) EncoderFuncs {
	out := ef.clone()

	ftyp := reflect.TypeOf(f)
	if ftyp == anyConvFuncType {
		out.anyConv = reflect.ValueOf(f).Interface().(func(any) (any, error))
		return out
	}

	if out.m == nil {
		out.m = make(map[reflect.Type]func(any) (any, error))
	}

	out.m[ftyp.In(0)] = func(v any) (any, error) { return f(v.(T)) }
	return out
}
