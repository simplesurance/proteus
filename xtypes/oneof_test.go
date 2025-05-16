package xtypes_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/simplesurance/proteus"
	"github.com/simplesurance/proteus/internal/assert"
	"github.com/simplesurance/proteus/plog"
	"github.com/simplesurance/proteus/sources/cfgtest"
	"github.com/simplesurance/proteus/types"
	"github.com/simplesurance/proteus/xtypes"
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

	_, err := proteus.MustParse(&params, proteus.WithProviders(testProvider))
	assert.NoErrorNow(t, err)

	assert.Equal(t, "do", params.Normal.Value())
	assert.Equal(t, "re", params.Secret.Value())
	assert.Equal(t, "mi", params.Optional.Value())
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
	assert.ErrorNow(t, err)

	violations := types.ErrViolations{}
	assert.TrueNow(t, errors.As(err, &violations), fmt.Sprintf(
		"Returned error should be types.ErrViolations, but is %T", err))

	assert.EqualNow(t, 1, len(violations))

	violation := violations[0]
	assert.Equal(t, "", violation.SetName)
	assert.Equal(t, "v", violation.ParamName)
	t.Logf("got error, as expected: %v", violation)
}

func TestOneOfBadDefault(t *testing.T) {
	params := struct {
		P *xtypes.OneOf `param:",optional"`
	}{
		P: &xtypes.OneOf{
			DefaultValue: "fa",
			Choices:      []string{"do", "re", "mi"},
		},
	}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{},
	})

	assert.PanicsNow(t, func() {
		_, _ = proteus.MustParse(&params, proteus.WithProviders(testProvider))
	})

}

func TestOneOfCallbackProvidedParameter(t *testing.T) {
	invoked := false

	params := struct {
		P *xtypes.OneOf
	}{
		P: &xtypes.OneOf{
			Choices:  []string{"do", "re", "mi"},
			UpdateFn: func(_ string) { invoked = true },
		},
	}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{
			"p": "do",
		},
	})

	_, err := proteus.MustParse(&params, proteus.WithProviders(testProvider))
	assert.NoErrorNow(t, err)
	assert.True(t, invoked, "UpdateFn was not invoked")
}

func TestOneOfRevertToDefault(t *testing.T) {
	var setUpdatedValue *string

	params := struct {
		P *xtypes.OneOf `param:",optional"`
	}{
		P: &xtypes.OneOf{
			Choices:      []string{"do", "re", "mi"},
			DefaultValue: "do",
			UpdateFn: func(s string) {
				setUpdatedValue = &s
			},
		},
	}

	testProvider := cfgtest.New(types.ParamValues{
		"": map[string]string{"p": "mi"},
	})

	_, err := proteus.MustParse(&params,
		proteus.WithLogger(plog.TestLogger(t)),
		proteus.WithProviders(testProvider))
	assert.NoErrorNow(t, err)

	assert.NotNilNow(t, setUpdatedValue)
	assert.Equal(t, "mi", *setUpdatedValue)
	assert.Equal(t, "mi", params.P.Value())

	testProvider.Update("", "p", nil)
	assert.NotNilNow(t, setUpdatedValue)
	assert.Equal(t, "do", *setUpdatedValue)
	assert.Equal(t, "do", params.P.Value())
}
