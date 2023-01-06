package mapx

import (
	"reflect"
)

type (
	SkipValue struct{}
	NoChange  struct{}
)

type EncodingConverter struct {
	m       map[reflect.Type]func(any) (any, error)
	anyConv func(any) (any, error)
}

var anyConvFuncType = reflect.TypeOf((func(any) (any, error))(nil))

func RegisterEncoder[T, V any](c *EncodingConverter, f func(T) (V, error)) {
	ftyp := reflect.TypeOf(f)
	if ftyp == anyConvFuncType {
		c.anyConv = reflect.ValueOf(f).Interface().(func(any) (any, error))
		return
	}

	if c.m == nil {
		c.m = make(map[reflect.Type]func(any) (any, error))
	}

	c.m[ftyp.In(0)] = func(v any) (any, error) { return f(v.(T)) }
}

type DecodingConverter struct {
	m          map[reflect.Type]decodingFunc
	ifaceFuncs map[reflect.Type][]ifaceDecodingFunc
}

type decodingFunc struct {
	dst reflect.Type
	f   func(any, any) error
}

type ifaceDecodingFunc struct {
	f   func(any, any) error
	typ reflect.Type
}

func RegisterDecoder[T, V any](c *DecodingConverter, f func(T, V) error) {
	ftyp := reflect.TypeOf(f)

	if ftyp.In(1).Kind() == reflect.Interface {
		if c.ifaceFuncs == nil {
			c.ifaceFuncs = make(map[reflect.Type][]ifaceDecodingFunc)
		}

		if ftyp.In(1).NumMethod() == 0 {
			panic("mapx: empty interface not allowed as destination type for RegisterDecoder")
		}

		c.ifaceFuncs[ftyp.In(0)] = append(c.ifaceFuncs[ftyp.In(0)],
			ifaceDecodingFunc{
				f:   func(v, dst any) error { return f(v.(T), dst.(V)) },
				typ: ftyp.In(1),
			},
		)
		return
	}

	if c.m == nil {
		c.m = make(map[reflect.Type]decodingFunc)
	}

	c.m[ftyp.In(0)] = decodingFunc{
		dst: ftyp.In(1),
		f:   func(v, dst any) error { return f(v.(T), dst.(V)) },
	}
}
