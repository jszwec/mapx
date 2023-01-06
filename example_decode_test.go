package mapx_test

import (
	"fmt"
	"strconv"

	"github.com/jszwec/mapx"
)

type User struct {
	Name         string
	Age          int `mapx:"-"`
	Address      Address
	PhoneNumbers []string
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
	}

	var user User
	if err := mapx.Decode(m, &user); err != nil {
		panic(err)
	}

	fmt.Print(user)

	// Output:
	// {Jacek 0 {Washington St 50} [123-456-7890]}
}

func ExampleDecoder_Decode() {
	// dfs should be a global variable.
	dfs := mapx.RegisterDecoder(mapx.DecoderFuncs{}, func(s string, dst *int) error {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*dst = int(n)
		return nil
	})

	// decoders should be global variables.
	dec := mapx.NewDecoder[*User](mapx.DecoderOpt{
		DecoderFuncs: dfs,
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
	}

	var user User
	if err := dec.Decode(m, &user); err != nil {
		panic(err)
	}

	fmt.Print(user)

	// Output:
	// {Jacek 0 {Washington St 50} [123-456-7890]}
}
