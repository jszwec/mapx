package mapx

import "reflect"

type (
	SkipValue struct{}
	NoChange  struct{}
)

type Converter struct {
	m       map[reflect.Type]func(any) (any, error)
	anyConv func(any) (any, error)
}

var anyConvFuncType = reflect.TypeOf((func(any) (any, error))(nil))

func Register[T, V any](c *Converter, f func(T) (V, error)) {
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

func WithConverter[T interface{ WithConverter(V) }, V any](c V) func(T) {
	return func(v T) { v.WithConverter(c) }
}

func WithTag[T interface{ WithTag(string) }](s string) func(T) {
	return func(v T) { v.WithTag(s) }
}
