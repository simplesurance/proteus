package proteus_test

import (
	"fmt"
	"os"

	"github.com/simplesurance/proteus"
)

func Example_paramSet() {
	params := struct {
		HTTP httpParams
		DB   dbParams
	}{}

	parsed, err := proteus.MustParse(&params)
	if err != nil {
		parsed.WriteError(os.Stderr, err)
		os.Exit(1)
	}

	connectDB(params.DB)
	startHTTP(params.HTTP)
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

func connectDB(dbP dbParams) {
	fmt.Printf("Connecting to DB server %s (db=%q, user=%q)",
		dbP.Server, dbP.Database, dbP.Username)
}

func startHTTP(httpP httpParams) {
	fmt.Printf("Starting HTTP server on :%d\n", httpP.BindPort)
}
