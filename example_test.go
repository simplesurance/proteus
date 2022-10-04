package proteus_test

import (
	"fmt"
	"os"

	"github.com/simplesurance/proteus"
)

func Example() {
	params := struct {
		Server string
		Port   uint16 `param:",optional"`
	}{
		Port: 5432,
	}

	parsed, err := proteus.MustParse(&params)
	if err != nil {
		parsed.WriteError(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Server: %s:%d\n", params.Server, params.Port)
}
