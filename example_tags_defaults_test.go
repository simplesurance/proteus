package proteus_test

import (
	"fmt"
	"os"

	"github.com/simplesurance/proteus"
)

func Example_tagsAndDefaults() {
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
