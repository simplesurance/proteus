// Package sources holds types that are shared between packages.
package sources

import "github.com/simplesurance/proteus/types"

// Source can read parameters from one medium. It can optionally watch the
// value of the parameter and report changes to it.
type Source interface {
	Stop()

	// Watch will read and return all parameters it can find. Optionally,
	// the provider can start watching for changes in the values and
	// notify about changes in the values.
	Watch(
		paramIDs Parameters,
		updater Updater,
	) (initial types.ParamValues, err error)
}

// Updater mediates the communication between the "parsed" object and one
// specific parameter source. It also minimizes the interface between
// them to an absolute minimum.
//
// Responsibilities include:
// - offer a way for the sources to provide updates to a parameter. This is
//   only useful for providers that support hot-updating parameters.
// - make sure synchronized fields are only accessed after holding the mutex
// - provide a logger to a parameter source. This logger uses the "parsed"
//   logger and mark all messages it creates, clearly identifying the source
//   that generated it.
// - provides the source with a method of finding out some very basic
//   properties of a parameter, that might be necessary for reading their
//   values.
type Updater interface {
	// Update can be called by a parameter source to notify about a change
	// on the value of a parameter. Useful only for sources that support
	// hot-updating values.
	Update(types.ParamValues)

	// Log allows the parameter source to use the logger from the parser.
	// All log entries will be identified with the class name of the
	// source.
	Log(string)
}

// Parameters contains information about what parameters the application
// expect, as long as some basic information about them.
type Parameters map[string]map[string]ParameterInfo

// ParameterInfo holds basic information about one application parameter that
// can be used by a configuration source.
type ParameterInfo struct {
	IsBool bool

	// FlagOnly means that this parameter can only be provided using
	// command-line flags.
	FlagOnly bool
}
