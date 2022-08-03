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
		Database    dbParams
		Environment *xtypes.OneOf `param_desc:"What environment the app is running on"`
		Port        uint16
	}{
		Database: defaultDBParams(),
		Environment: &xtypes.OneOf{
			Choices: []string{"dev", "stg", "prd"},
		},
	}

	parsed, err := proteus.MustParse(&params,
		proteus.WithAutoUsage(os.Stderr, "Demo App", func() { os.Exit(0) }),
		proteus.WithSources(
			cfgflags.New(),
			cfgenv.New("CFG"),
		))
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Running in %s mode\n", params.Environment.Value())
}

type dbParams struct {
	Server   string `param:",optional"    param_desc:"Name of the database server"`
	Port     uint16 `param:",optional"    param_desc:"TCP port number of database server"`
	User     string `                     param_desc:"Username for authentication"`
	Password string `                     param_desc:"Password for authentication"`
}

func defaultDBParams() dbParams {
	return dbParams{
		Server: "localhost",
		Port:   5432,
	}
}
