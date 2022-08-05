// Package cfgtest implements a configuration reader that can be used in tests.
package cfgtest

import (
	"testing"

	"github.com/simplesurance/proteus/sources"
	"github.com/simplesurance/proteus/types"
)

// New creates a new parameter source that can be used on tests. An initial
// value for the parameters can be provided on construction. The returned
// object can be used to change the values, allowing for tests on parameters
// that change without reloading.
func New(t *testing.T, values types.ParamValues) *TestSource {
	return &TestSource{values: values, t: t}
}

// TestSource is an application configuration source designed to be used on
// tests.
type TestSource struct {
	t       *testing.T
	values  types.ParamValues
	updater sources.Updater
}

var _ sources.Source = &TestSource{}

// Update changes a value on the test provider, allowing for test on
// hot-reloading of parameters.
func (r *TestSource) Update(setid, id string, value string) {
	set, ok := r.values[setid]
	if !ok {
		set = map[string]string{}
		r.values[setid] = set
	}

	set[id] = value

	r.updater.Update(r.values)
}

// Stop releases any resource being used by the TestSource.
func (r *TestSource) Stop() {
}

// Watch reads parameters from environment variables. Since environment
// variables never change, we only read once, and we don't have to watch
// for changes.
func (r *TestSource) Watch(
	paramIDs sources.Parameters,
	updater sources.Updater,
) (initial types.ParamValues, err error) {
	r.updater = updater
	return r.values, nil
}
