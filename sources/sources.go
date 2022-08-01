// Package sources defines the interface for configuration source providers.
package sources

import (
	"github.com/simplesurance/proteus/plog"
	"github.com/simplesurance/proteus/types"
)

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
	// flags. Proteus will call this method on a provider to determine if
	// "special parameters", like "--help" can be provided by it. Proteus will
	// only read special parameters from providers that return true on
	// this method. This is to avoid having situations where for example
	// an environment variable would inadvertently cause the application to
	// show help instead of run.
	IsCommandLineFlag() bool
}

// Updater is an interface that has as its primary use allowing providers to
// notify proteus about changes in parameter values.
//
// When multiple providers are configured, one updater is created for each
// one of them, and it is passed to the respective provider as a parameter of
// `Watch`. When the provider calls the Update() method on its own updater,
// that will result in updating the copy of the parameter values for that
// provider.
//
// This updater also allows providers to produce log messages.
type Updater interface {
	// Update notify about a change in parameter values.
	// Useful only for providers that support hot-updating values.
	Update(types.ParamValues)

	// Log allows the provider to use the logger from the parser.
	// All log entries will be identified with the class name of the
	// provider.
	Log(plog.Entry)

	// Peek reads the raw parameter value from the providers registered
	// before the provider associated to this updater. This allow one
	// provider to be configured by values received by other providers.
	Peek(setName, paramName string) (*string, error)
}

// Parameters contains information that proteus makes available to providers
// about what parameters the application expects.
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

// ParameterInfo holds information about a configuration parameter that is
// made available to a configuration provider.
type ParameterInfo struct {
	IsBool bool
}
