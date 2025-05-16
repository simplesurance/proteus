//go:build unittest || !integrationtest
// +build unittest !integrationtest

package cfgenv_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/simplesurance/proteus/internal/assert"
	"github.com/simplesurance/proteus/plog"
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
			assert.NoErrorNow(t, err)
		}
	}()

	os.Clearenv()
	assert.NoErrorNow(t, os.Setenv("TEST__A", "1"))
	assert.NoErrorNow(t, os.Setenv("TEST__B", "2"))
	assert.NoErrorNow(t, os.Setenv("TEST__ENABLED_BOOL", "true"))

	assert.NoErrorNow(t, os.Setenv("TEST__PARAMSET1__A", "11"))
	assert.NoErrorNow(t, os.Setenv("TEST__PARAMSET1__B", "12"))
	assert.NoErrorNow(t, os.Setenv("TEST__PARAMSET1__ENABLED_BOOL", "true"))

	assert.NoErrorNow(t, os.Setenv("TEST__PARAMSET2__A", "21"))
	assert.NoErrorNow(t, os.Setenv("TEST__PARAMSET2__B", "22"))
	assert.NoErrorNow(t, os.Setenv("TEST__PARAMSET2__ENABLED_BOOL", "false"))

	assert.NoErrorNow(t, os.Setenv("MUST_IGNORE_THIS", "1"))

	paramSource := cfgenv.New("TEST")
	values, err := paramSource.Watch(sources.Parameters{
		"":          map[string]sources.ParameterInfo{"a": {}, "b": {}, "c": {}, "enabled_bool": {IsBool: true}, "other_bool": {IsBool: true}},
		"paramset1": map[string]sources.ParameterInfo{"a": {}, "b": {}, "c": {}, "enabled_bool": {IsBool: true}, "other_bool": {IsBool: true}},
		"paramset2": map[string]sources.ParameterInfo{"a": {}, "b": {}, "c": {}, "enabled_bool": {IsBool: true}, "other_bool": {IsBool: true}},
		"paramset3": map[string]sources.ParameterInfo{"a": {}, "b": {}, "c": {}, "enabled_bool": {IsBool: true}, "other_bool": {IsBool: true}},
	}, &testUpdater{
		LogFn: plog.TestLogger(t),
		IsBooleanFn: func(_, paramName string) bool {
			return strings.HasSuffix(paramName, "bool")
		},
	})
	assert.NoErrorNow(t, err)

	want := types.ParamValues{
		"": map[string]string{
			"a":            "1",
			"b":            "2",
			"enabled_bool": "true",
		},
		"paramset1": map[string]string{
			"a":            "11",
			"b":            "12",
			"enabled_bool": "true",
		},
		"paramset2": map[string]string{
			"a":            "21",
			"b":            "22",
			"enabled_bool": "false",
		},
	}

	if !reflect.DeepEqual(want, values) {
		jhave, _ := json.MarshalIndent(values, "", " ")
		jwant, _ := json.MarshalIndent(want, "", " ")
		t.Errorf("Not the expected configuration.\nWANT\n%s\n\nHAVE\n%s", jwant, jhave)
	}
}

func TestUnexpectedEnvVar(t *testing.T) {
	envCopy := os.Environ()
	defer func() {
		os.Clearenv()
		for _, v := range envCopy {
			key, value, _ := strings.Cut(v, "=")
			assert.NoErrorNow(t, os.Setenv(key, value))
		}
	}()

	os.Clearenv()
	assert.NoErrorNow(t, os.Setenv("TEST__UNEXPECTED", "1"))

	paramSource := cfgenv.New("TEST")
	_, err := paramSource.Watch(sources.Parameters{
		"": map[string]sources.ParameterInfo{"expected": {}},
	}, &testUpdater{
		LogFn: plog.TestLogger(t),
		IsBooleanFn: func(_, _ string) bool {
			return false
		},
	})
	assert.ErrorNow(t, err)
}

type testUpdater struct {
	UpdateFn    func(types.ParamValues)
	LogFn       plog.Logger
	IsBooleanFn func(setName, paramName string) bool
}

var _ sources.Updater = &testUpdater{}

func (t *testUpdater) Update(p types.ParamValues) {
	t.UpdateFn(p)
}

func (t *testUpdater) Log(entry plog.Entry) {
	t.LogFn(entry)
}

func (*testUpdater) Peek(_, _ string) (*string, error) {
	// environment variables do not read values from another providers
	return nil, nil
}
