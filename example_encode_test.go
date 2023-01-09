package mapx_test

import (
	"fmt"
	"strconv"

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
	}

	m, err := mapx.Encode(u)
	if err != nil {
		panic(err)
	}

	fmt.Print(m)

	// Output:
	// map[Address:map[Unit:50 street:Washington St] Name:Jacek PhoneNumbers:[123-456-7890]]
}

func ExampleEncoder_Encode() {
	// encoderFuncs should be a global variable.
	encoderFuncs := mapx.RegisterEncoder(mapx.EncoderFuncs{}, func(n int) (string, error) {
		return strconv.Itoa(n), nil
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
	}

	m, err := enc.Encode(&u)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", m)

	// Output:
	// map[string]interface {}{"Address":map[string]interface {}{"Unit":"50", "street":"Washington St"}, "Name":"Jacek", "PhoneNumbers":[]string{"123-456-7890"}}
}
