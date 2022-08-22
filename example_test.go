package proteus_test

import (
	"fmt"
	"log"
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

	parsed, err := proteus.MustParse(&params, proteus.WithPrintfLogger(log.Printf))
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Server: %s:%d\n", params.Server, params.Port)
}
