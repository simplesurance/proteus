// Package cfgtest is a configuration provider to be used on tests.
package cfgtest

import (
	"sync"

	"github.com/simplesurance/proteus/sources"
	"github.com/simplesurance/proteus/types"
)

// New creates a new provider that can be used on tests. An initial
// value for the parameters can be provided on construction. The returned
// object can be used to change the values, allowing for tests on parameters
// that change without reloading.
func New(values types.ParamValues) *TestProvider {
	ret := &TestProvider{}

	ret.protected.mutex.Lock()
	ret.protected.values = values.Copy()
	ret.protected.mutex.Unlock()

	return ret
}

// TestProvider is an application configuration provider designed to be used on
// tests.
type TestProvider struct {
	protected struct {
		mutex  sync.Mutex
		values types.ParamValues
	}
	updater sources.Updater
}

var _ sources.Provider = &TestProvider{}

// IsCommandLineFlag returns true, to allow tests to handle special flags that
// only command-line flags are allowed to process.
func (r *TestProvider) IsCommandLineFlag() bool {
	return true
}

// Update changes a value on the test provider, allowing for test on
// hot-reloading of parameters.
func (r *TestProvider) Update(setid, id string, value string) {
	valuesCopy := func() types.ParamValues {
		r.protected.mutex.Lock()
		defer r.protected.mutex.Unlock()

		set, ok := r.protected.values[setid]
		if !ok {
			set = map[string]string{}
			r.protected.values[setid] = set
		}

		set[id] = value

		return r.protected.values.Copy()
	}()

	r.updater.Update(valuesCopy)
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

	r.protected.mutex.Lock()
	defer r.protected.mutex.Unlock()

	return r.protected.values.Copy(), nil
}
