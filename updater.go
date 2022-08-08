package proteus

import (
	"fmt"

	"github.com/simplesurance/proteus/sources"
	"github.com/simplesurance/proteus/types"
)

// updater mediates the communication between a parameter source and the
// parsed values. See sources.Updater for details.
type updater struct {
	parsed *Parsed

	// providerIndex identifies the position on Parsed where values from this
	// provider should go. Parsed holds values for all providers in a
	// slice, this is the providerIndex for that slice.
	providerIndex int
	providerName  string
}

var _ sources.Updater = &updater{}

func (u *updater) Update(v types.ParamValues) {
	u.update(v, true)
}

func (u *updater) Log(msg string) {
	u.parsed.settings.loggerFn(u.providerName+": "+msg, 2)
}

func (u *updater) update(v types.ParamValues, refresh bool) {
	u.mustBeOnValidIDs(v)
	u.validateValues(v)

	u.parsed.valuesMutex.Lock()
	u.parsed.values[u.providerIndex] = v
	u.parsed.valuesMutex.Unlock()

	if refresh {
		u.parsed.refresh(false) // update only dynamic parameters
	}
}

func (u *updater) validateValues(v types.ParamValues) {
	for setName, set := range v {
		for paramName, value := range set {
			err := u.parsed.validValue(setName, paramName, &value)
			if err != nil {
				u.parsed.settings.loggerFn(fmt.Sprintf(
					"provider %q update: parameter %s.%s: %v",
					u.providerName, setName, paramName, err), 2)
			}
		}
	}
}

// mustBeOnValidIDs panics if one of the values provided by the provided is
// not on the list of ids the application registered.
func (u *updater) mustBeOnValidIDs(v types.ParamValues) {
	for setName, set := range v {
		if _, ok := u.parsed.inferedConfig[setName]; !ok {
			panic(fmt.Errorf("parameter source is providing value for unsolicited paramset %q", setName))
		}

		for paramName := range set {
			if _, ok := u.parsed.inferedConfig[setName].fields[paramName]; !ok {
				panic(fmt.Errorf("parameter source is providing value for unsolicited parameter %s.%s", setName, paramName))
			}
		}
	}
}
