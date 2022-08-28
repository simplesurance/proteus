//go:build unittest || !integrationtest
// +build unittest !integrationtest

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

func TestSignedInt(t *testing.T) {
	params := struct {
		I8_1 *xtypes.Integer[int8]
		I8_2 *xtypes.Integer[int8]
	}{}

	providedParameters := types.ParamValues{
		"": map[string]string{
			"i8_1": "127",
			"i8_2": "-128",
		},
	}

	parsed, err := proteus.MustParse(&params, proteus.WithProviders(
		cfgtest.New(providedParameters)))
	assert.NoErrorNow(t, err)

	buffer := bytes.Buffer{}
	parsed.Dump(&buffer)
	t.Log("DUMP OF PROVIDED PARAMETERS\n" + buffer.String())

	buffer = bytes.Buffer{}
	parsed.Usage(&buffer)
	t.Log("USAGE INFORMATION\n" + buffer.String())

	x := params.I8_1.Value()

	t.Logf("%T %T", params.I8_1.Value(), params.I8_2.Value())

	assert.Equal(t, int8(127), x)
	assert.Equal(t, int8(-128), params.I8_2.Value())
}
