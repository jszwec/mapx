package mapx_test

import (
	"encoding"
	"fmt"
	"strconv"
	"time"

	"github.com/jszwec/mapx"
)

type User struct {
	Name         string
	Age          int `mapx:"-"`
	Address      Address
	PhoneNumbers []string
	CreatedAt    time.Time
}

type Address struct {
	Street string `mapx:"street"`
	Unit   int
}

func ExampleDecode() {
	m := map[string]any{
		"Name": "Jacek",
		"Age":  100,
		"Address": map[string]any{
			"street": "Washington St",
			"Unit":   50,
		},
		"PhoneNumbers": []any{"123-456-7890"},
		"CreatedAt":    time.Date(2023, 1, 8, 12, 0, 0, 0, time.UTC),
	}

	var user User
	if err := mapx.Decode(m, &user); err != nil {
		panic(err)
	}

	fmt.Print(user)

	// Output:
	// {Jacek 0 {Washington St 50} [123-456-7890] 2023-01-08 12:00:00 +0000 UTC}
}

func ExampleDecoder_Decode() {
	// decoderFuncs should be a global variable.
	decoderFuncs := mapx.RegisterDecoder(mapx.DecoderFuncs{}, func(s string, dst *int) error {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*dst = int(n)
		return nil
	})

	decoderFuncs = mapx.RegisterDecoder(decoderFuncs, func(s string, tu encoding.TextUnmarshaler) error {
		return tu.UnmarshalText([]byte(s))
	})

	// decoders should be global variables.
	dec := mapx.NewDecoder[*User](mapx.DecoderOpt{
		DecoderFuncs: decoderFuncs,
		Tag:          "mapx",
	})

	m := map[string]any{
		"Name": "Jacek",
		"Age":  "100",
		"Address": map[string]any{
			"street": "Washington St",
			"Unit":   "50", // our registered decoder function will parse this to int.
		},
		"PhoneNumbers": []any{"123-456-7890"},
		"CreatedAt":    "2023-01-08T12:00:00Z",
	}

	var user User
	if err := dec.Decode(m, &user); err != nil {
		panic(err)
	}

	fmt.Print(user)

	// Output:
	// {Jacek 0 {Washington St 50} [123-456-7890] 2023-01-08 12:00:00 +0000 UTC}
}
