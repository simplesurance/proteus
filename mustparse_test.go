package proteus_test

import (
	"fmt"
	"os"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgenv"
	"github.com/simplesurance/proteus/sources/cfgflags"
	"github.com/simplesurance/proteus/xtypes"
)

func ExampleMustParse() {
	params := struct {
		Server string
		Port   uint16
	}{}

	parsed, err := proteus.MustParse(&params, proteus.WithSources(
		cfgflags.New(),
		cfgenv.New("CFG"),
	))
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Server: %s:%d\n", params.Server, params.Port)
}

func ExampleMustParse_defaultValues() {
	params := struct {
		Enabled bool   `param:"is_enabled,optional" param_desc:"Allows enabling or disabling the HTTP server"`
		Port    uint16 `param:",optional"           param_desc:"Port to bind for the HTTP server"`
		Token   string `param:",secret"             param_desc:"Token clients must provide for authentication"`
	}{
		Enabled: true,
		Port:    8080,
	}

	parsed, err := proteus.MustParse(&params, proteus.WithSources(
		cfgflags.New(),
		cfgenv.New("CFG"),
	))
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	if params.Enabled {
		fmt.Printf("Starting HTTP server on :%d\n", params.Port)
	}
}

func ExampleMustParse_xtypes() {
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

func ExampleMustParse_parameterSets() {
	type httpParams struct {
		BindPort uint16 `param:"bind_port"`
		Password string `param:"pwd,secret"`
	}

	type dbParams struct {
		Server   string
		Username string `param:"user"`
		Password string `param:"pwd,secret"`
		Database string
	}

	params := struct {
		HTTP httpParams
		DB   dbParams
	}{}

	parsed, err := proteus.MustParse(&params, proteus.WithSources(
		cfgflags.New(),
		cfgenv.New("CFG"),
	))
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	connectDB(params.DB)
	startHTTP(params.HTTP)
}

func connectDB(_ any) {
}

func startHTTP(_ any) {
}
