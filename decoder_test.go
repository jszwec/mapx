package mapx_test

import (
	"encoding"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/jszwec/mapx"

	"github.com/google/go-cmp/cmp"
)

type C struct {
	A             Int    `custom:"a"`
	B             string `custom:"-"`
	V             any
	TextMarshaler encoding.TextMarshaler
}

type D struct {
	CS   []C
	C    C
	Ints []int
	Any  any
	Anys []any
}

type (
	Int    int
	String string
	Bool   bool
)

var stringIntDecConverter = func() mapx.DecodingConverter {
	var c mapx.DecodingConverter
	mapx.RegisterDecoder(&c, func(s string, dst *int) error {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*dst = int(n)
		return nil
	})
	return c
}()

func TestDecode(t *testing.T) {
	tm := time.Date(2022, 8, 4, 12, 0, 0, 0, time.UTC)

	fixtures := []struct {
		desc     string
		m        map[string]any
		opts     []mapx.DecoderOpt
		expected any
		err      error
	}{
		{
			desc: "simple",
			m: map[string]any{
				"A":             1,
				"B":             "hello",
				"V":             false,
				"TextMarshaler": tm,
			},
			expected: &C{
				A:             1,
				B:             "hello",
				V:             false,
				TextMarshaler: tm,
			},
			err: nil,
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
			err: nil,
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
			opts: []mapx.DecoderOpt{
				mapx.WithTag[*mapx.Decoder]("custom"),
			},
			err: nil,
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
			err: nil,
		},
		{
			desc: "with arrays - iface array to type",
			m: map[string]any{
				"Ints": []any{1, 2, 3},
			},
			expected: &D{
				Ints: []int{1, 2, 3},
			},
			err: nil,
		},
		{
			desc: "with arrays - cast array to type",
			m: map[string]any{
				"Ints": []float64{1, 2, 3},
			},
			expected: &D{
				Ints: []int{1, 2, 3},
			},
			err: nil,
		},
		{
			desc: "with custom decoder",
			m: map[string]any{
				"Int":  "1",
				"Ints": []string{"1", "2", "3"},
			},
			opts: []mapx.DecoderOpt{
				mapx.WithConverter[*mapx.Decoder](stringIntDecConverter),
			},
			expected: &struct {
				Int  int
				Ints []int
			}{
				Int:  1,
				Ints: []int{1, 2, 3},
			},
			err: nil,
		},
	}

	for _, f := range fixtures {
		t.Run(f.desc, func(t *testing.T) {
			dst := reflect.New(reflect.TypeOf(f.expected).Elem())

			if err := mapx.Decode(f.m, dst.Interface(), f.opts...); err != nil {
				t.Error(err)
			}

			if d := cmp.Diff(f.expected, dst.Interface()); d != "" {
				t.Error(d)
			}
		})
	}
}
