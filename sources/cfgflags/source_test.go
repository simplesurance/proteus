//go:build unittest || !integrationtest
// +build unittest !integrationtest

package cfgflags_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/simplesurance/proteus/sources"
	"github.com/simplesurance/proteus/sources/cfgflags"
	"github.com/simplesurance/proteus/types"
)

func TestCfgEnv(t *testing.T) {
	argCopy := make([]string, len(os.Args))
	copy(argCopy, os.Args)
	defer func() {
		os.Args = argCopy
	}()

	os.Args = []string{
		"./the/binary/name",

		"-a=1",
		"-b", "2",
		"-enabled_bool",

		"flagset1",
		"-a=11",
		"-b", "12",
		"-enabled_bool=true",

		"flagset2",
		"-a=21",
		"-b", "22",
		"-enabled_bool=false"}

	flagSource := cfgflags.New()
	values, err := flagSource.Watch(sources.Parameters{
		"": map[string]sources.ParameterInfo{
			"a":            {},
			"b":            {},
			"c":            {},
			"enabled_bool": {IsBool: true},
			"other_bool":   {IsBool: true}},
		"flagset1": map[string]sources.ParameterInfo{
			"a":            {},
			"b":            {},
			"c":            {},
			"enabled_bool": {IsBool: true},
			"other_bool":   {IsBool: true}},
		"flagset2": map[string]sources.ParameterInfo{
			"a":            {},
			"b":            {},
			"c":            {},
			"enabled_bool": {IsBool: true},
			"other_bool":   {IsBool: true}},
		"flagset3": map[string]sources.ParameterInfo{
			"a":            {},
			"b":            {},
			"c":            {},
			"enabled_bool": {IsBool: true},
			"other_bool":   {IsBool: true}},
	}, &testUpdater{t: t})
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
			"enabled_bool": "false"}},
		values)
}

func TestUnexpectedParameter(t *testing.T) {
	argCopy := make([]string, len(os.Args))
	copy(argCopy, os.Args)
	defer func() {
		os.Args = argCopy
	}()

	os.Args = []string{
		"./the/binary/name",
		"-unexpected"}

	flagSource := cfgflags.New()
	_, err := flagSource.Watch(sources.Parameters{
		"": map[string]sources.ParameterInfo{
			"expected": {}},
	}, &testUpdater{t: t})
	assert.Error(t, err)
}

type testUpdater struct {
	t        *testing.T
	UpdateFn func(types.ParamValues)
}

var _ sources.Updater = &testUpdater{}

func (t *testUpdater) Update(p types.ParamValues) {
	t.UpdateFn(p)
}

func (t *testUpdater) Log(msg string) {
	t.t.Helper()
	t.t.Log(msg)
}
