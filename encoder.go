package mapx

import (
	"errors"
	"reflect"
)

var defaultEncoder = NewEncoder[any](EncoderOpt{})

type EncoderOpt struct {
	Converter EncodingConverter
	Tag       string
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

		if e.opts.Converter.m != nil {
			if fn, ok := e.opts.Converter.m[f.typ]; ok {
				m[f.name], err = fn(v.Field(f.index[0]).Interface())
				if err != nil {
					return nil, err
				}
				continue
			}
		}

		fv := v.Field(f.index[0]).Interface()

		if e.opts.Converter.anyConv != nil {
			v, err := e.opts.Converter.anyConv(fv)
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
