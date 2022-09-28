package proteus_test

import (
	"bytes"
	"testing"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/internal/assert"
	"github.com/simplesurance/proteus/plog"
	"github.com/simplesurance/proteus/sources/cfgtest"
	"github.com/simplesurance/proteus/types"
)

func TestDump(t *testing.T) {
	params := struct {
		Server string
		Port   uint16 `param:",optional"`
		Token  string `param:",optional,secret"`
		Key    string `param:",secret"`
	}{
		Port:  8080,
		Token: "secret-token",
	}

	provider := cfgtest.New(types.ParamValues{
		"": {
			"server": "localhost",
			"key":    "secret-key",
		},
	})

	parsed, err := proteus.MustParse(&params,
		proteus.WithLogger(plog.TestLogger(t)),
		proteus.WithProviders(provider))
	assert.NoErrorNow(t, err)

	usageBuffer := bytes.Buffer{}

	parsed.Dump(&usageBuffer)

	t.Log(usageBuffer.String())

	assert.Equal(t, `Parameter values:
- help = "false" (default)
- key = "<redacted>"
- port = "8080" (default)
- server = "localhost"
- token = "<redacted>" (default)
`, usageBuffer.String())
}
