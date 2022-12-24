package mapx_test

import (
	"mapx"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type A struct {
	A1 int
	B  B
}

type B struct {
	B1   int
	Ints []int
	Map  map[string]int
}

func TestStruct(t *testing.T) {
	fixtures := []struct {
		desc string
		in   any
		opts []mapx.EncoderOpt
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
			opts: []mapx.EncoderOpt{mapx.WithConverter[*mapx.Encoder](mapx.JSONConverter)},
			out: map[string]any{
				"A1": float64(100),
				"B": map[string]any{
					"B1":   float64(200),
					"Ints": []any{float64(1), float64(2)},
					"Map":  map[string]any{"1": float64(1)},
				},
			},
		},
	}

	for _, f := range fixtures {
		t.Run(f.desc, func(t *testing.T) {
			out, err := mapx.Encode(f.in, f.opts...)
			if err != nil {
				t.Fatal("expected err=nil; got ", err)
			}

			if d := cmp.Diff(f.out, out); d != "" {
				t.Error(d)
			}
		})
	}
}
