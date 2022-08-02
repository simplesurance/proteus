package main

import (
	"fmt"
	"os"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgenv"
	"github.com/simplesurance/proteus/sources/cfgflags"
	"github.com/simplesurance/proteus/xtypes"
)

func main() {
	params := struct {
		BindPort   uint16
		PrivKey    *xtypes.RSAPrivateKey
		RequestLog *xtypes.OneOf
	}{
		RequestLog: &xtypes.OneOf{
			Choices: []string{"none", "basic", "full"},
			UpdateFn: func(s string) {
				fmt.Printf("Log level changed to %s", s)
			},
		},
	}

	parsed, err := proteus.MustParse(&params, proteus.WithSources(
		cfgflags.New(),
		cfgenv.New("CFG"),
	))
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Starting HTTP server on :%d with an RSA key of size %d, log level %s\n",
		params.BindPort,
		params.PrivKey.Value().Size(),
		params.RequestLog.Value())
}
