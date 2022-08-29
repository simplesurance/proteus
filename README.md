# Proteus
![Coverage](https://img.shields.io/badge/Coverage-58.4%25-yellow)
[![Go Report Card](https://goreportcard.com/badge/github.com/simplesurance/proteus)](https://goreportcard.com/report/github.com/simplesurance/proteus)

## About

Proteus is a package for defining the configuration of an Go application in a
struct and loading it from different sources. Application can also opt-in to
getting updates when the configuration changes.

## Project Status

This project is in pre-release stage and backwards compatibility is not
guaranteed.

## How to Get

```bash
go get github.com/simplesurance/proteus@latest
```

## How to Use

Specify the parameters and from what sources they can be provided:
```go
func main() {
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
```

Provide the parameters using one of the configured providers:
```sh
go run main.go -server a -port 123
# Output: Server: a:123

CFG__SERVER=b CFG__PORT=42 go run *.go
# Output: Server: b:42
```

In the default configuration, as shown above, command-line flags have priority
over environment variables.

### Sets of Parameters

Parameters can be organized in _parameter sets_. Two parameters with the same
name can exist, as long as they belong to different sets.

```go
package main

func main() {
	params := struct {
		HTTP httpParams
		DB   dbParams
	}{}

	parsed, err := proteus.MustParse(&params)
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

Note that one parameter was provided as environment variable and the others
as command-line flags.

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

	parsed, err := proteus.MustParse(&params)
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	if params.Enabled {
		fmt.Printf("Starting HTTP server on :%d\n", params.Port)
	}
}
```

Only `-token` is mandatory:

```bash
go run *.go -token my-secret-token

# Output: Starting HTTP server on :8080
```

The token is marked as a secret, which is important to avoid leaking its value.

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

	parsed, err := proteus.MustParse(&params)
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

### Auto-Generated Usage (a.k.a --help)

To have usage information include the `WithAutoUsage` option:

```go
func main() {
	params := struct {
		Database    dbParams
		Environment *xtypes.OneOf `param_desc:"Which environment the app is running on"`
		Port        uint16
	}{
		Database: defaultDBParams(),
		Environment: &xtypes.OneOf{
			Choices: []string{"dev", "stg", "prd"},
		},
	}

	parsed, err := proteus.MustParse(&params,
		proteus.WithAutoUsage(os.Stderr, "Demo App", func() { os.Exit(0) }))
	if err != nil {
		parsed.ErrUsage(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Running in %s mode\n", params.Environment.Value())
}

type dbParams struct {
	Server   string `param:",optional"    param_desc:"Name of the database server"`
	Port     uint16 `param:",optional"    param_desc:"TCP port number of the database server"`
	User     string `                     param_desc:"Username for authentication"`
	Password string `                     param_desc:"Password for authentication"`
}

func defaultDBParams() dbParams {
	return dbParams{
		Server: "localhost",
		Port:   5432,
	}
}
```

```bash
go run main.go --help

# Output:
# app: Multiple errors:
# - "environment": parameter is required but was not specified
# - "port": parameter is required but was not specified
# - "database.password": parameter is required but was not specified
# - "database.user": parameter is required but was not specified
#
#
# Usage:
# app \
#     [-help] \
#     -environment <dev|stg|prd> \
#     -port <uint16> \
#   database \
#     -password <string> \
#     -user <string> \
#     [-port <uint16>] \
#     [-server <string>]
#
# PARAMETERS
# - help default=false
#   Prints information about how to use this application
# - environment
#   Which environment the app is running on
# - port
#
# PARAMETER SET: DATABASE
# - password
#   Password for authentication
# - user
#   Username for authentication
# - port default=5432
#   TCP port number of the database server
# - server default=localhost
#   Name of the database server
```

## Supported Providers

- [cfgenv](sources/cfgenv/): For environ variables
- [cfgflags](sources/cfgflags/): For command-line flags
- [cfgtest](sources/cfgtest/): For tests
- [cfgconsul](https://github.com/simplesurance/proteus-consul): For HashiCorp
  Consul
