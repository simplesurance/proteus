package proteus_test

import (
	"fmt"
	"os"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgenv"
)

func ExampleMustParse() {
	params := struct {
		Server string
		Port   uint16
	}{}

	parsed, err := proteus.MustParse(&params, proteus.WithSources(cfgenv.New("TEST")))
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Server: %s:%q", params.Server, params.Port)
}
