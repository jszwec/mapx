package mapx_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/jszwec/mapx"

	"github.com/google/go-cmp/cmp"
)

type A struct {
	A1 int
	B  B
}

type B struct {
	B1   int `custom:"b1"`
	Ints []int
	Map  map[string]int `custom:"-"`
}

type Embedded struct {
	A int
	*B
}

var stringerEncoder = mapx.RegisterEncoder(mapx.EncoderFuncs{}, func(s fmt.Stringer) (string, error) {
	return s.String(), nil
})

func (n Int) String() string {
	return strconv.Itoa(int(n))
}

// ptr receiver testing.
func (b *Bool) String() string {
	return strconv.FormatBool(bool(*b))
}

func TestStruct(t *testing.T) {
	fixtures := []struct {
		desc string
		in   any
		opts mapx.EncoderOpt
		out  map[string]any
	}{
		{
			desc: "basic map",
			in: struct {
				String string
				Int    uint
			}{
				String: "hello",
				Int:    100,
			},
			out: map[string]any{
				"String": "hello",
				"Int":    uint(100),
			},
		},
		{
			desc: "inner map",
			in: A{
				A1: 100,
				B: B{
					B1:   200,
					Ints: []int{1, 2},
					Map:  map[string]int{"1": 1},
				},
			},
			out: map[string]any{
				"A1": 100,
				"B": map[string]any{
					"B1":   200,
					"Ints": []int{1, 2},
					"Map":  map[string]int{"1": 1},
				},
			},
		},
		{
			desc: "json converter",
			in: A{
				A1: 100,
				B: B{
					B1:   200,
					Ints: []int{1, 2},
					Map:  map[string]int{"1": 1},
				},
			},
			opts: mapx.EncoderOpt{
				EncoderFuncs: mapx.JSONEncoderFuncs,
			},
			out: map[string]any{
				"A1": float64(100),
				"B": map[string]any{
					"B1":   float64(200),
					"Ints": []any{float64(1), float64(2)},
					"Map":  map[string]any{"1": float64(1)},
				},
			},
		},
		{
			desc: "with tags",
			in: B{
				B1:   200,
				Ints: []int{1, 2},
				Map:  map[string]int{"1": 1},
			},
			opts: mapx.EncoderOpt{
				Tag: "custom",
			},
			out: map[string]any{
				"b1":   200,
				"Ints": []int{1, 2},
			},
		},
		{
			desc: "with interface encoder",
			in: &struct {
				Int     Int
				PtrInt  *Int
				NilInt  *Int
				Bool    Bool
				PtrBool *Bool
				NilBool *Bool
			}{
				Int:     5,
				PtrInt:  ptr(Int(10)),
				Bool:    true,
				PtrBool: ptr(Bool(true)),
			},
			opts: mapx.EncoderOpt{
				EncoderFuncs: stringerEncoder,
			},
			out: map[string]any{
				"Int":     "5",
				"PtrInt":  "10",
				"NilInt":  nil,
				"Bool":    "true",
				"PtrBool": "true",
				"NilBool": nil,
			},
		},
		{
			desc: "embedded",
			in: &struct {
				Int Int
				B
			}{
				Int: 5,
				B: B{
					B1:   100,
					Ints: []int{10},
					Map:  map[string]int{"foo": 1},
				},
			},
			out: map[string]any{
				"Int":  Int(5),
				"B1":   100,
				"Ints": []int{10},
				"Map":  map[string]int{"foo": 1},
			},
		},
		{
			desc: "nil embedded",
			in: &struct {
				Int Int
				*B
			}{
				Int: 5,
				B:   nil,
			},
			out: map[string]any{
				"Int":  Int(5),
				"B1":   nil,
				"Ints": nil,
				"Map":  nil,
			},
		},
		{
			desc: "deep embedded",
			in: &struct {
				Int Int
				*Embedded
			}{
				Int:      5,
				Embedded: &Embedded{A: 100, B: nil},
			},
			out: map[string]any{
				"Int":  Int(5),
				"A":    100,
				"B1":   nil,
				"Ints": nil,
				"Map":  nil,
			},
		},
		{
			desc: "embedded field conflict",
			in: &struct {
				A
				B int
			}{
				A: A{
					A1: 100,
					B:  B{B1: 99},
				},
				B: 999,
			},
			out: map[string]any{
				"A1": 100,
				"B":  999,
			},
		},
		{
			desc: "inline",
			in: &struct {
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
			out: map[string]any{
				"Int":  Int(5),
				"B1":   100,
				"Ints": []int{10},
				"Map":  map[string]int{"foo": 1},
			},
		},
		{
			desc: "inline field conflict",
			in: &struct {
				A A `mapx:",inline"`
				B int
			}{
				A: A{
					A1: 100,
					B:  B{B1: 99},
				},
				B: 999,
			},
			out: map[string]any{
				"A1": 100,
				"B":  999,
			},
		},
		{
			desc: "raw tag",
			in: &struct {
				A A `mapx:",raw"`
				B int
			}{
				A: A{
					A1: 100,
					B:  B{B1: 99},
				},
				B: 999,
			},
			out: map[string]any{
				"A": A{
					A1: 100,
					B:  B{B1: 99},
				},
				"B": 999,
			},
		},
	}

	for _, f := range fixtures {
		t.Run(f.desc, func(t *testing.T) {
			enc := mapx.NewEncoder[any](f.opts)
			out, err := enc.Encode(f.in)
			if err != nil {
				t.Fatal("expected err=nil; got ", err)
			}

			if d := cmp.Diff(f.out, out); d != "" {
				t.Error(d)
			}
		})
	}
}
