// Package sources defines the interface for configuration source providers.
package sources

import "github.com/simplesurance/proteus/types"

// Provider parses parameters from a medium. It can optionally watch the
// value of the parameter and report changes to it.
type Provider interface {
	Stop()

	// Watch parses the configuration source and returns all found parameters.
	// Optionally, the provider can start watching for changes in the source
	// and notify about changes via the updater.
	Watch(
		paramIDs Parameters,
		updater Updater,
	) (initial types.ParamValues, err error)

	// IsCommandLineFlag indicates if this provider reads from command-line
	// flags. Some "special parameters", like "--help" can only be handled
	// by command-line flags.
	IsCommandLineFlag() bool
}

// Updater is an interface that allow providers to notify about changes on
// configuration.
//
// When multiple providers are configured, one updater is created for each
// one of them, and it is passed to the respective provider as a parameter of
// `Watch`. Calling the Update() method on one of the Updaters will result in
// updating the copy of the parameter values for that provider.
//
// In case of issues reading or updating values, a log function is also
// available.
type Updater interface {
	// Update notify about a change in parameter values.
	// Useful only for providers that support hot-updating values.
	Update(types.ParamValues)

	// Log allows the provider to use the logger from the parser.
	// All log entries will be identified with the class name of the
	// provider.
	Log(string)

	// Peek reads the raw parameter value from the providers registered
	// before the provider associated to this updater. This allow one
	// provider to be configured by values received by other providers.
	Peek(setName, paramName string) (*string, error)
}

// Parameters contains information about what parameters the application
// expect, as long as some basic information about them.
type Parameters map[string]map[string]ParameterInfo

// Get returns the parameter information with the given set and parameter name.
func (p Parameters) Get(setName, paramName string) (ret ParameterInfo, found bool) {
	if set, ok := p[setName]; ok {
		if info, ok := set[paramName]; ok {
			return info, true
		}
	}

	return ret, false
}

// ParameterInfo holds information about a configuration parameter.
type ParameterInfo struct {
	IsBool bool
}
