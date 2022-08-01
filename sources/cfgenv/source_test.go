//go:build unittest || !integrationtest
// +build unittest !integrationtest

package cfgenv_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/simplesurance/proteus/sources"
	"github.com/simplesurance/proteus/sources/cfgenv"
	"github.com/simplesurance/proteus/types"
)

func TestCfgEnv(t *testing.T) {
	envCopy := os.Environ()
	defer func() {
		os.Clearenv()
		for _, v := range envCopy {
			key, value, _ := strings.Cut(v, "=")
			err := os.Setenv(key, value)
			require.NoError(t, err)
		}
	}()

	os.Clearenv()
	require.NoError(t, os.Setenv("TEST__A", "1"))
	require.NoError(t, os.Setenv("TEST__B", "2"))
	require.NoError(t, os.Setenv("TEST__ENABLED_BOOL", ""))

	require.NoError(t, os.Setenv("TEST__FLAGSET1__A", "11"))
	require.NoError(t, os.Setenv("TEST__FLAGSET1__B", "12"))
	require.NoError(t, os.Setenv("TEST__FLAGSET1__ENABLED_BOOL", "true"))

	require.NoError(t, os.Setenv("TEST__FLAGSET2__A", "21"))
	require.NoError(t, os.Setenv("TEST__FLAGSET2__B", "22"))
	require.NoError(t, os.Setenv("TEST__FLAGSET2__ENABLED_BOOL", "false"))

	require.NoError(t, os.Setenv("MUST_IGNORE_THIS", "1"))

	flagSource := cfgenv.New("TEST")
	values, err := flagSource.Watch(sources.Parameters{
		"":         map[string]sources.ParameterInfo{"a": {}, "b": {}, "c": {}, "enabled_bool": {IsBool: true}, "other_bool": {IsBool: true}},
		"flagset1": map[string]sources.ParameterInfo{"a": {}, "b": {}, "c": {}, "enabled_bool": {IsBool: true}, "other_bool": {IsBool: true}},
		"flagset2": map[string]sources.ParameterInfo{"a": {}, "b": {}, "c": {}, "enabled_bool": {IsBool: true}, "other_bool": {IsBool: true}},
		"flagset3": map[string]sources.ParameterInfo{"a": {}, "b": {}, "c": {}, "enabled_bool": {IsBool: true}, "other_bool": {IsBool: true}},
	}, &testUpdater{
		t: t,
		IsBooleanFn: func(setName, paramName string) bool {
			return strings.HasSuffix(paramName, "bool")
		},
	})
	require.NoError(t, err)

	require.Equal(t, types.ParamValues{
		"": map[string]string{
			"a":            "1",
			"b":            "2",
			"enabled_bool": "true",
		},
		"flagset1": map[string]string{
			"a":            "11",
			"b":            "12",
			"enabled_bool": "true",
		},
		"flagset2": map[string]string{
			"a":            "21",
			"b":            "22",
			"enabled_bool": "false",
		},
	}, values)
}

func TestUnexpectedEnvVar(t *testing.T) {
	envCopy := os.Environ()
	defer func() {
		os.Clearenv()
		for _, v := range envCopy {
			key, value, _ := strings.Cut(v, "=")
			require.NoError(t, os.Setenv(key, value))
		}
	}()

	os.Clearenv()
	require.NoError(t, os.Setenv("TEST__UNEXPECTED", "1"))

	flagSource := cfgenv.New("TEST")
	_, err := flagSource.Watch(sources.Parameters{
		"": map[string]sources.ParameterInfo{"expected": {}},
	}, &testUpdater{
		t: t,
		IsBooleanFn: func(setName, paramName string) bool {
			return false
		},
	})
	require.Error(t, err)
}

type testUpdater struct {
	t           *testing.T
	UpdateFn    func(types.ParamValues)
	IsBooleanFn func(setName, paramName string) bool
}

var _ sources.Updater = &testUpdater{}

func (t *testUpdater) Update(p types.ParamValues) {
	t.UpdateFn(p)
}

func (t *testUpdater) Log(msg string) {
	t.t.Helper()
	t.t.Log(msg)
}
