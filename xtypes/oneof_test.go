package xtypes_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/sources/cfgtest"
	"github.com/simplesurance/proteus/types"
	"github.com/simplesurance/proteus/xtypes"
	"github.com/stretchr/testify/require"
)

func TestOneOfValid(t *testing.T) {
	notes := []string{"do", "re", "mi"}

	params := struct {
		Normal   *xtypes.OneOf
		Secret   *xtypes.OneOf `param:",secret"`
		Optional *xtypes.OneOf `param:",optional"`
	}{
		Normal: &xtypes.OneOf{
			Choices: notes,
		},
		Secret: &xtypes.OneOf{
			Choices: notes,
		},
		Optional: &xtypes.OneOf{
			Choices:      notes,
			DefaultValue: "mi",
		},
	}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{
			"normal": "do",
			"secret": "re",
		},
	})

	parsed, err := proteus.MustParse(&params, proteus.WithProviders(testProvider))
	require.NoError(t, err)

	buffer := bytes.Buffer{}
	parsed.Dump(&buffer)

	require.Equal(t, `Parameter values:
- normal = "do"
- secret = "<redacted>"
`, buffer.String())

	require.Equal(t, "do", params.Normal.Value())
	require.Equal(t, "re", params.Secret.Value())
	require.Equal(t, "mi", params.Optional.Value())

	buffer.Reset()
	parsed.Usage(&buffer)
	require.Equal(t, `Syntax:
./xtypes.test \
    <-normal (do|re|mi)> \
    <-secret (do|re|mi)> \
    [-optional (do|re|mi)]

PARAMETERS
- normal:(do|re|mi)
- secret:(do|re|mi)
- optional:(do|re|mi) default=mi

`, buffer.String())
}

func TestOneOfInvalid(t *testing.T) {
	notes := []string{"do", "re", "mi"}

	params := struct {
		V *xtypes.OneOf
	}{
		V: &xtypes.OneOf{
			Choices: notes,
		},
	}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{
			"v": "sol",
		},
	})

	_, err := proteus.MustParse(&params, proteus.WithProviders(testProvider))
	require.Error(t, err)

	violations := types.ErrViolations{}
	require.True(t, errors.As(err, &violations))

	require.Equal(t, 1, len(violations))

	violation := violations[0]
	require.Equal(t, "", violation.SetName)
	require.Equal(t, "v", violation.ParamName)
	t.Logf("got error, as expected: %v", violation)
}
