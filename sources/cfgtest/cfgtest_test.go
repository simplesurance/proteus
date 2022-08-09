package cfgtest_test

import (
	"os"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgtest"
	"github.com/simplesurance/proteus/types"
)

func Example() {
	params := struct {
		Server string
		Port   uint16 `param:",optional"`
	}{}

	parsed, err := proteus.MustParse(&params, proteus.WithProviders(cfgtest.New(types.ParamValues{
		"": map[string]string{
			"server": "localhost.localdomain",
			"port":   "42",
		},
	})))
	if err != nil {
		panic(err)
	}

	parsed.Dump(os.Stdout)

	// Output:
	// Parameter values:
	// - port = "42"
	// - server = "localhost.localdomain"
}
