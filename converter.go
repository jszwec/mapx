package mapx

import "reflect"

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
	m map[reflect.Type]decodingFunc
}

type decodingFunc struct {
	dst reflect.Type
	f   func(any, any) error
}

func RegisterDecoder[T, V any](c *DecodingConverter, f func(T, *V) error) {
	ftyp := reflect.TypeOf(f)

	if c.m == nil {
		c.m = make(map[reflect.Type]decodingFunc)
	}

	c.m[ftyp.In(0)] = decodingFunc{
		dst: reflect.TypeOf((*V)(nil)).Elem(),
		f:   func(v, dst any) error { return f(v.(T), dst.(*V)) },
	}
}
