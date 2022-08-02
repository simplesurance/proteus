package main

import (
	"fmt"
	"os"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgenv"
	"github.com/simplesurance/proteus/sources/cfgflags"
)

func main() {
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

func connectDB(dbP dbParams) {
	fmt.Printf("Connecting to %s on db %s with user=%s\n",
		dbP.Server, dbP.Database, dbP.Username)
}

func startHTTP(httpP httpParams) {
	fmt.Printf("Starting HTTP server on :%d\n", httpP.BindPort)
}

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
