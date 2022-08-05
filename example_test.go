package proteus_test

import (
	"fmt"
	"os"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgenv"
	"github.com/simplesurance/proteus/sources/cfgflags"
)

func Example() {
	params := struct {
		Server string
		Port   uint16 `param:",optional"`
	}{
		Port: 5432,
	}

	parsed, err := proteus.MustParse(&params, proteus.WithProviders(
		cfgflags.New(),
		cfgenv.New("CFG"),
	))
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Server: %s:%d\n", params.Server, params.Port)
}
