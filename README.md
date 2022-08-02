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

Note that one parameter was provided by environment variable and the others
with command-line flags.

### Struct Tags and Defaults

Some struct tags are supported to allow specifying some details about the
parameter. Default values can also be provided:

```go
func main() {
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
```

Only token has to be specified:

```bash
go run *.go -token my-secret-token

# Output: Starting HTTP server on :8080
```

The token is marked as a secret, which is important to avoid leaking its value.

// TODO: link with doc for `Dump()`, `Usage()` and `ErrUsage()`.

### XTypes

_XTypes_ are types provided by _proteus_ to handle complex types and to provide
more additional functionality.

```go
package main

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

// TODO: Document `Dump()`, `Usage()` and `ErrUsage()`.