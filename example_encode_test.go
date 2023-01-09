package mapx_test

import (
	"encoding"
	"fmt"
	"strconv"
	"time"

	"github.com/jszwec/mapx"
)

func ExampleEncode() {
	u := User{
		Name: "Jacek",
		Age:  100, // ignored
		Address: Address{
			Street: "Washington St",
			Unit:   50,
		},
		PhoneNumbers: []string{"123-456-7890"},
		CreatedAt:    time.Date(2023, 1, 8, 12, 0, 0, 0, time.UTC),
	}

	m, err := mapx.Encode(u)
	if err != nil {
		panic(err)
	}

	fmt.Print(m)

	// Output:
	// map[Address:map[Unit:50 street:Washington St] CreatedAt:2023-01-08 12:00:00 +0000 UTC Name:Jacek PhoneNumbers:[123-456-7890]]
}

func ExampleEncoder_Encode() {
	// encoderFuncs should be a global variable.
	encoderFuncs := mapx.RegisterEncoder(mapx.EncoderFuncs{}, func(n int) (string, error) {
		return strconv.Itoa(n), nil
	})

	encoderFuncs = mapx.RegisterEncoder(encoderFuncs, func(tm encoding.TextMarshaler) (string, error) {
		text, err := tm.MarshalText()
		return string(text), err
	})

	// encoders should be global variables.
	enc := mapx.NewEncoder[*User](mapx.EncoderOpt{
		EncoderFuncs: encoderFuncs,
		Tag:          "mapx",
	})

	u := User{
		Name: "Jacek",
		Age:  100, // ignored
		Address: Address{
			Street: "Washington St",
			Unit:   50,
		},
		PhoneNumbers: []string{"123-456-7890"},
		CreatedAt:    time.Date(2023, 1, 8, 12, 0, 0, 0, time.UTC),
	}

	m, err := enc.Encode(&u)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", m)

	// Output:
	// map[string]interface {}{"Address":map[string]interface {}{"Unit":"50", "street":"Washington St"}, "CreatedAt":"2023-01-08T12:00:00Z", "Name":"Jacek", "PhoneNumbers":[]string{"123-456-7890"}}
}
