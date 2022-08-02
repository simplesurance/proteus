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
