package mapx_test

import (
	"encoding"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/jszwec/mapx"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type C struct {
	A             Int    `custom:"a"`
	B             string `custom:"-"`
	V             any
	Ptr           *int
	TextMarshaler encoding.TextMarshaler
}

type D struct {
	CS      []C
	C       C
	Ints    []int
	PtrInts *[]int
	PInts   []*int
	Any     any
	Anys    []any
}

type (
	Int    int
	String string
	Bool   bool
)

func (n *Int) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return err
	}
	*n = Int(v)
	return nil
}

// the point is a value-receiver.
type wrappedInt struct {
	*Int
}

func (n wrappedInt) UnmarshalText(text []byte) error {
	return n.Int.UnmarshalText(text)
}

var stringIntDecConverter = func() mapx.DecoderFuncs {
	return mapx.RegisterDecoder(mapx.DecoderFuncs{}, func(s string, dst *int) error {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*dst = int(n)
		return nil
	})
}()

var textUnmarshalerDecConverter = func() mapx.DecoderFuncs {
	return mapx.RegisterDecoder(mapx.DecoderFuncs{}, func(s string, dst encoding.TextUnmarshaler) error {
		return dst.UnmarshalText([]byte(s))
	})
}()

var tm = time.Date(2022, 8, 4, 12, 0, 0, 0, time.UTC)

func TestDecode(t *testing.T) {

	fixtures := []struct {
		desc     string
		m        map[string]any
		opts     mapx.DecoderOpt
		dst      any
		expected any
		err      error
	}{
		{
			desc: "simple",
			m: map[string]any{
				"A":             1,
				"B":             "hello",
				"V":             false,
				"Ptr":           ptr(10),
				"TextMarshaler": tm,
			},
			expected: &C{
				A:             1,
				B:             "hello",
				V:             false,
				Ptr:           ptr(10),
				TextMarshaler: tm,
			},
		},
		{
			desc: "to ptr",
			m: map[string]any{
				"Ptr": 10,
			},
			expected: &C{
				Ptr: ptr(10),
			},
		},
		{
			desc: "to ptr - nil",
			m: map[string]any{
				"Ptr": nil,
			},
			expected: &C{},
			err:      nil,
		},
		{
			desc: "to ptr - fast conv",
			m: map[string]any{
				"Ptr": int8(10),
			},
			expected: &C{
				Ptr: ptr(10),
			},
		},
		{
			desc: "to ptr - conv",
			m: map[string]any{
				"Ptr": uint8(10),
			},
			expected: &C{
				Ptr: ptr(10),
			},
		},
		{
			desc: "type aliases",
			m: map[string]any{
				"A":             1.0,
				"B":             String("hello"),
				"V":             Bool(false),
				"TextMarshaler": tm,
			},
			expected: &C{
				A:             1,
				B:             "hello",
				V:             Bool(false),
				TextMarshaler: tm,
			},
		},
		{
			desc: "custom tag",
			m: map[string]any{
				"a": 1,
				"B": "hello",
			},
			expected: &C{
				A: 1,
			},
			opts: mapx.DecoderOpt{
				Tag: "custom",
			},
		},
		{
			desc: "with arrays",
			m: map[string]any{
				"CS": []any{
					map[string]any{
						"A":             1,
						"B":             "hello",
						"V":             false,
						"TextMarshaler": tm,
					},
				},
				"C": map[string]any{
					"A":             1,
					"B":             "hello",
					"V":             false,
					"TextMarshaler": tm,
				},
				"Ints": []int{1, 2, 3},
				"Any":  10,
				"Anys": []any{10, "lol"},
			},
			expected: &D{
				CS: []C{
					{
						A:             1,
						B:             "hello",
						V:             false,
						TextMarshaler: tm,
					},
				},
				C: C{
					A:             1,
					B:             "hello",
					V:             false,
					TextMarshaler: tm,
				},
				Ints: []int{1, 2, 3},
				Any:  10,
				Anys: []any{10, "lol"},
			},
		},
		{
			desc: "with arrays - iface array to type",
			m: map[string]any{
				"Ints":    []any{1, 2, 3},
				"PtrInts": []any{1, 2, 3},
				"PInts":   []any{ptr(1), 2, nil},
			},
			expected: &D{
				Ints:    []int{1, 2, 3},
				PtrInts: &[]int{1, 2, 3},
				PInts:   []*int{ptr(1), ptr(2), nil},
			},
		},
		{
			desc: "with slices - cast slice to type",
			m: map[string]any{
				"Ints":    []float64{1, 2, 3},
				"PtrInts": []float64{1, 2, 3},
				"PInts":   []float64{1, 2, 3},
			},
			expected: &D{
				Ints:    []int{1, 2, 3},
				PtrInts: &[]int{1, 2, 3},
				PInts:   []*int{ptr(1), ptr(2), ptr(3)},
			},
		},
		{
			desc: "with custom decoder",
			m: map[string]any{
				"Int":  "1",
				"Ints": []string{"1", "2", "3"},
			},
			opts: mapx.DecoderOpt{
				DecoderFuncs: stringIntDecConverter,
			},
			expected: &struct {
				Int  int
				Ints []int
			}{
				Int:  1,
				Ints: []int{1, 2, 3},
			},
		},
		{
			desc: "with custom decoder - interface - ptr receiver",
			m: map[string]any{
				"Int":   "1",
				"Ints":  []string{"1", "2", "3"},
				"PInts": []string{"1", "2", "3"},
			},
			opts: mapx.DecoderOpt{
				DecoderFuncs: textUnmarshalerDecConverter,
			},
			expected: &struct {
				Int   Int
				Ints  []Int
				PInts []*Int
			}{
				Int:   1,
				Ints:  []Int{1, 2, 3},
				PInts: []*Int{ptr(Int(1)), ptr(Int(2)), ptr(Int(3))},
			},
		},
		{
			desc: "embedded",
			m: map[string]any{
				"A":    10,
				"B1":   100,
				"Ints": []int{1},
				"Map":  map[string]int{"foo": 1},
			},
			expected: &Embedded{
				A: 10,
				B: &B{
					B1:   100,
					Ints: []int{1},
					Map:  map[string]int{"foo": 1},
				},
			},
		},
		{
			desc: "embedded field conflict",
			m: map[string]any{
				"A1": 100,
				"B":  999,
			},
			expected: &struct {
				A
				B int
			}{
				A: A{
					A1: 100,
					B:  B{},
				},
				B: 999,
			},
		},
		{
			desc: "inline",
			m: map[string]any{
				"Int":  Int(5),
				"B1":   100,
				"Ints": []int{10},
				"Map":  map[string]int{"foo": 1},
			},
			expected: &struct {
				Int Int
				B   B `mapx:",inline"`
			}{
				Int: 5,
				B: B{
					B1:   100,
					Ints: []int{10},
					Map:  map[string]int{"foo": 1},
				},
			},
		},
		{
			desc: "inline field conflict",
			m: map[string]any{
				"A1": 100,
				"B":  999,
			},
			expected: &struct {
				A A `mapx:",inline"`
				B int
			}{
				A: A{
					A1: 100,
					B:  B{},
				},
				B: 999,
			},
		},
		{
			desc: "with custom decoder - interface - value receiver",
			m: map[string]any{
				"Int":  "1",
				"Ints": []string{"1", "2", "3"},
			},
			opts: mapx.DecoderOpt{
				DecoderFuncs: textUnmarshalerDecConverter,
			},
			dst: &struct {
				Int wrappedInt
			}{
				Int: wrappedInt{new(Int)},
			},
			expected: &struct {
				Int wrappedInt
			}{
				Int: wrappedInt{ptr(Int(1))},
			},
		},
		{
			desc: "error - string to int",
			m: map[string]any{
				"A": "1",
			},
			dst: &C{},
			err: &mapx.DecodeError{
				Value: "1",
				Type:  reflect.TypeOf((*Int)(nil)).Elem(),
			},
		},
		{
			desc: "error - nil value",
			m: map[string]any{
				"A": nil,
			},
			dst: &C{},
			err: &mapx.DecodeError{
				Value: nil,
				Type:  reflect.TypeOf((*Int)(nil)).Elem(),
			},
		},
		{
			desc: "error - int to ptr int",
			m: map[string]any{
				"Ptr": struct{}{},
			},
			dst: &C{},
			err: &mapx.DecodeError{
				Value: struct{}{},
				Type:  reflect.TypeOf((*int)(nil)).Elem(),
			},
		},
		{
			desc: "error - slice with invalid type",
			m: map[string]any{
				"Ptrs": []any{struct{}{}},
			},
			dst: &struct {
				Ptrs []int
			}{},
			err: &mapx.DecodeError{
				Value: struct{}{},
				Type:  reflect.TypeOf((*int)(nil)).Elem(),
			},
		},
		{
			desc: "error - slice with nil values",
			m: map[string]any{
				"Ptrs": []any{nil},
			},
			dst: &struct {
				Ptrs []int
			}{},
			err: &mapx.DecodeError{
				Value: nil,
				Type:  reflect.TypeOf((*int)(nil)).Elem(),
			},
		},
		{
			desc: "error - not a struct",
			m:    map[string]any{},
			dst:  ptr(5),
			err:  mapx.ErrNotAStruct,
		},
		{
			desc: "error - not a pointer",
			m:    map[string]any{},
			dst:  C{},
			err:  mapx.ErrNotAPointer,
		},
	}

	for _, f := range fixtures {
		t.Run(f.desc, func(t *testing.T) {
			var dst reflect.Value
			switch {
			case f.dst != nil:
				dst = reflect.ValueOf(f.dst)
			case f.expected != nil:
				dst = reflect.New(reflect.TypeOf(f.expected).Elem())
			default:
				dst = reflect.ValueOf(struct{}{})
			}

			dec := mapx.NewDecoder[any](f.opts)
			err := dec.Decode(f.m, dst.Interface())
			if f.err != nil {
				if d := cmp.Diff(f.err, err, cmpopts.EquateErrors()); d != "" {
					t.Error(d)
				}
				return
			} else if err != nil {
				t.Error(err)
				return
			}

			if d := cmp.Diff(f.expected, dst.Interface()); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestDecodeTypedD(t *testing.T) {
	in := map[string]any{
		"CS": []any{
			map[string]any{
				"A":             1,
				"B":             "hello",
				"V":             false,
				"TextMarshaler": tm,
			},
		},
		"C": map[string]any{
			"A":             1,
			"B":             "hello",
			"V":             false,
			"TextMarshaler": tm,
		},
		"Ints": []int{1, 2, 3},
		"Any":  10,
		"Anys": []any{10, "lol"},
	}

	var d D
	if err := mapx.NewDecoder[*D](mapx.DecoderOpt{}).Decode(in, &d); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(D{
		CS: []C{
			{
				A:             1,
				B:             "hello",
				V:             false,
				TextMarshaler: tm,
			},
		},
		C: C{
			A:             1,
			B:             "hello",
			V:             false,
			TextMarshaler: tm,
		},
		Ints: []int{1, 2, 3},
		Any:  10,
		Anys: []any{10, "lol"},
	}, d); diff != "" {
		t.Error(diff)
	}
}

func ptr[T any](v T) *T { return &v }
