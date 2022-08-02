# Proteus

## About
Proteus is a library for structured dynamic configuration for go service. It
can be used to specify parameters a go service needs, and the parameters can
be provided as command-line flags or environment variables.

## How to Use

Specify the parameters and from what sources they can be provided:
```go
func main() {
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
```

Provide the parameters using one of the configured providers:
```sh
go run main.go -server a -port 123
# Output: Server: a:123

CFG__SERVER=b CFG__PORT=42 go run *.go
# Output: Server: b:42
```

When the same parameter is provided by two providers, the one listed first
have priority.

### Sets of Parameters
Parameters can be namespaces into _sets_ of parameters. This allows using
simple names like _server_ for parameters in one namespace without clashing
with a parameter on other namespace.

```go
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
```

```bash
CFG__DB__SERVER=localhost go run *.go \
  db \ 
    -database library \
    -user sa \
    -pwd sa \
  http \  
    -bind_port 5432 \
    -pwd secret-token

# Output:
# Connecting to localhost on db library with user=sa
# Starting HTTP server on :5432
```

Node that one parameter was provided by environment variable and the others
with command-line flags.

### XTypes

_XTypes_ are types provided by _proteus_ to handle complex types and to provide
more additional functionality.

```go
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
```
