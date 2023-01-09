package mapx

import (
	"errors"
	"reflect"
	"time"
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
loop:
	for _, f := range fields {
		fv := fieldByIndex(v, f.index)
		if !fv.IsValid() {
			m[f.name] = nil
			continue
		}

		dst := fv.Interface()

		if e.opts.EncoderFuncs.m != nil {
			if fn, ok := e.opts.EncoderFuncs.m[f.baseType]; ok {
				m[f.name], err = fn(fv.Interface())
				if err != nil {
					return nil, err
				}
				continue
			}
		}

		if e.opts.EncoderFuncs.ifaceFuncs != nil {
			for _, fn := range e.opts.EncoderFuncs.ifaceFuncs {
				if f.baseType.Implements(fn.argType) {
					if f.baseType.Kind() == reflect.Pointer && fv.IsNil() {
						m[f.name] = nil
						continue loop
					}
					m[f.name], err = fn.f(fv.Interface())
					if err != nil {
						return nil, err
					}
					continue loop
				}

				if reflect.PointerTo(f.baseType).Implements(fn.argType) && fv.CanAddr() {
					m[f.name], err = fn.f(fv.Addr().Interface())
					if err != nil {
						return nil, err
					}
					continue loop
				}
			}
		}

		if e.opts.EncoderFuncs.anyConv != nil {
			v, err := e.opts.EncoderFuncs.anyConv(dst)
			if err != nil {
				return nil, err
			}

			switch v {
			case SkipValue{}:
				continue
			case NoChange{}:
			default:
				dst = v
			}
		}

		if f.typ.Kind() == reflect.Struct && (!f.tag.raw && !isKnownStruct(f.typ)) {
			sub, err := e.encode(fv, f.fields)
			if err != nil {
				return nil, err
			}
			m[f.name] = sub
			continue
		}

		m[f.name] = dst
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

type EncoderFuncs struct {
	anyConv    func(any) (any, error)
	m          map[reflect.Type]func(any) (any, error)
	ifaceFuncs []encodingFunc
}

type encodingFunc struct {
	argType reflect.Type
	f       func(any) (any, error)
}

func (ef EncoderFuncs) clone() EncoderFuncs {
	var (
		m          map[reflect.Type]func(any) (any, error)
		ifaceFuncs []encodingFunc
	)

	if ef.m != nil {
		m = make(map[reflect.Type]func(any) (any, error), len(ef.m)+1)
		for k, v := range ef.m {
			m[k] = v
		}
	}

	if ef.ifaceFuncs != nil {
		cp := make([]encodingFunc, len(ef.ifaceFuncs), len(ef.ifaceFuncs)+1)
		copy(cp, ef.ifaceFuncs)
		ifaceFuncs = cp
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
	if ftyp.In(0).Kind() == reflect.Interface {
		if ftyp.In(0).NumMethod() == 0 {
			out.anyConv = reflect.ValueOf(f).Interface().(func(any) (any, error))
			return out
		}

		out.ifaceFuncs = append(out.ifaceFuncs,
			encodingFunc{
				argType: ftyp.In(0),
				f:       func(v any) (any, error) { return f(v.(T)) },
			},
		)
		return out
	}

	if out.m == nil {
		out.m = make(map[reflect.Type]func(any) (any, error))
	}

	out.m[ftyp.In(0)] = func(v any) (any, error) { return f(v.(T)) }
	return out
}

var timeType = reflect.TypeOf((*time.Time)(nil)).Elem()

func isKnownStruct(typ reflect.Type) bool {
	// we will possibly list more here.
	return timeType == typ
}
