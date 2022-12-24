package mapx

import (
	"errors"
	"reflect"
)

type EncoderOpt func(*Encoder)

type Encoder struct {
	converter Converter
}

func NewEncoder(opts ...EncoderOpt) *Encoder {
	var enc Encoder
	for _, opt := range opts {
		opt(&enc)
	}
	return &enc
}

func (e *Encoder) Encode(val any) (map[string]any, error) {
	return e.encode(reflect.ValueOf(val))
}

func (e *Encoder) WithConverter(c Converter) {
	e.converter = c
}

func (e *Encoder) encode(v reflect.Value) (_ map[string]any, err error) {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, errors.New("not a struct")
	}

	fs := cachedFields(typeKey{
		tag:  "mapx",
		Type: v.Type(),
	})

	m := make(map[string]any, len(fs))
	for _, f := range fs {
		if f.typ.Kind() == reflect.Struct {
			sub, err := e.encode(v.Field(f.index[0]))
			if err != nil {
				return nil, err
			}
			m[f.name] = sub
			continue
		}

		if e.converter.m != nil {
			if fn, ok := e.converter.m[f.typ]; ok {
				m[f.name], err = fn(v.Field(f.index[0]).Interface())
				if err != nil {
					return nil, err
				}
				continue
			}
		}

		fv := v.Field(f.index[0]).Interface()

		if e.converter.anyConv != nil {
			v, err := e.converter.anyConv(fv)
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

func Encode(val any, opts ...EncoderOpt) (map[string]any, error) {
	return NewEncoder(opts...).encode(reflect.ValueOf(val))
}
