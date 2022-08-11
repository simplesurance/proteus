package proteus_test

import (
	"os"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgenv"
)

func ExampleParsed_Usage() {
	params := struct {
		Server string `param_desc:"Name of the server to connect"`
		Port   uint16 `param:",optional" param_desc:"Port to conect"`
	}{
		Port: 5432,
	}

	parsed, _ := proteus.MustParse(&params, proteus.WithProviders(cfgenv.New("TEST")))

	parsed.Usage(os.Stdout)

	// Output:
	// Usage:
	// proteus.test \
	//     [-help] \
	//     -server <string> \
	//     [-port <uint16>]
	//
	// PARAMETERS
	// - help default=false
	//   Prints information about how to use this application
	// - server
	//   Name of the server to connect
	// - port default=5432
	//   Port to conect
}
