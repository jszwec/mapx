package mapx_test

import (
	"fmt"

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
