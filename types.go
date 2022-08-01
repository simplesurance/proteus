package proteus

import (
	"github.com/simplesurance/proteus/sources"
	"github.com/simplesurance/proteus/specialflags"
)

type config map[string]flagSet

// configIDs create configuration information that can be used by a
// configuration source.
func (c config) configIDs() sources.Parameters {
	ret := make(sources.Parameters, len(c))

	for fsName, fs := range c {
		paramIDs := make(map[string]sources.ParameterInfo, len(fs.fields))
		for paramName, info := range fs.fields {
			if fsName == "" && (paramName == specialflags.Help.Name || paramName == specialflags.DryMode.Name) {
				continue
			}

			paramIDs[paramName] = sources.ParameterInfo{
				IsBool: info.boolean,
			}
		}

		ret[fsName] = paramIDs
	}

	return ret
}

type flagSet struct {
	desc   string
	fields map[string]flagSetField
}

type flagSetField struct {
	typ      string
	optional bool
	secret   bool
	flagSet  bool
	desc     string
	boolean  bool
	path     string

	isDynamic    bool
	setValueFn   func(v *string) error
	validFn      func(v string) error
	getDefaultFn func() (string, error)
	redactFn     func(string) string
}

func (f flagSetField) redactedValue(v *string) func() string {
	return func() string {
		if f.secret {
			return redactedPlaceholder
		}

		if v == nil {
			return "<missing>"
		}

		return f.redactFn(*v)
	}
}

func (f flagSetField) redactedDefaultValue() string {
	if f.secret {
		return redactedPlaceholder
	}

	ret, err := f.getDefaultFn()
	if err != nil {
		return "<" + err.Error() + ">"
	}

	return f.redactFn(ret)
}
