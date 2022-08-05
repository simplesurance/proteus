// Package cfgtest is a configuration provider to be used on tests.
package cfgtest

import (
	"github.com/simplesurance/proteus/sources"
	"github.com/simplesurance/proteus/types"
)

// New creates a new provider that can be used on tests. An initial
// value for the parameters can be provided on construction. The returned
// object can be used to change the values, allowing for tests on parameters
// that change without reloading.
func New(values types.ParamValues) *TestProvider {
	return &TestProvider{values: values}
}

// TestProvider is an application configuration provider designed to be used on
// tests.
type TestProvider struct {
	values  types.ParamValues
	updater sources.Updater
}

var _ sources.Provider = &TestProvider{}

// Update changes a value on the test provider, allowing for test on
// hot-reloading of parameters.
func (r *TestProvider) Update(setid, id string, value string) {
	set, ok := r.values[setid]
	if !ok {
		set = map[string]string{}
		r.values[setid] = set
	}

	set[id] = value

	r.updater.Update(r.values)
}

// Stop does nothing.
func (r *TestProvider) Stop() {
}

// Watch reads parameters from environment variables. Since environment
// variables never change, we only read once, and we don't have to watch
// for changes.
func (r *TestProvider) Watch(
	paramIDs sources.Parameters,
	updater sources.Updater,
) (initial types.ParamValues, err error) {
	r.updater = updater
	return r.values, nil
}
