package xtypes_test

import (
	"bytes"
	"testing"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/internal/assert"
	"github.com/simplesurance/proteus/sources/cfgtest"
	"github.com/simplesurance/proteus/types"
	"github.com/simplesurance/proteus/xtypes"
)

func TestEd25519Priv(t *testing.T) {
	const testPrivED25519Key = `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEILeVMy9KxALhIuev5dTLmtb8u9weRofKqd+n7Vifb8G0
-----END PRIVATE KEY-----`

	params := struct {
		Key *xtypes.Ed25519PrivateKey
	}{}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{
			"key": testPrivED25519Key,
		},
	})

	parsed, err := proteus.MustParse(&params, proteus.WithProviders(testProvider))
	assert.NoErrorNow(t, err)

	buffer := bytes.Buffer{}
	parsed.Dump(&buffer)
	t.Logf("Dumped: %s", buffer.String())
	t.Logf("Priv: %v", params.Key.Value())
	t.Logf("Public: %v", params.Key.Value().Public())
}
