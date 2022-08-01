package proteus_test

import (
	"fmt"
	"os"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgenv"
	"github.com/simplesurance/proteus/sources/cfgflags"
)

func ExampleMustParse() {
	params := struct {
		Server string
		Port   uint16
	}{}

	parsed, err := proteus.MustParse(&params)
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Server: %s:%d\n", params.Server, params.Port)
}

func ExampleMustParse_withTags() {
	params := struct {
		Enabled bool   `param:"is_enabled,optional" param_desc:"Allows enabling or disabling the HTTP server"`
		Port    uint16 `param:",optional"           param_desc:"Port to bind for the HTTP server"`
		Token   string `param:",secret"             param_desc:"Client authentication token"`
	}{
		Enabled: true,
		Port:    8080,
	}

	parsed, err := proteus.MustParse(&params)
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	if params.Enabled {
		fmt.Printf("Starting HTTP server on :%d\n", params.Port)
	}
}

// ExampleMustParse_providers changes how and from where proteus reads
// configuration.
func ExampleMustParse_providers() {
	params := struct {
		Server string
		Port   uint16
	}{}

	parsed, err := proteus.MustParse(&params,
		proteus.WithProviders(
			cfgenv.New("CONFIG"), // change env var prefix to CONFIG
			cfgflags.New()))      // flags are used, but priority is to env vars
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Server: %s:%d\n", params.Server, params.Port)
}
