package proteus

import (
	"github.com/simplesurance/proteus/sources"
	"github.com/simplesurance/proteus/types"
)

// add appends information about a parameter. If the item already exists
// "ok" will return false.
//
// This is not implemented as a method of "Parameters" because only the
// proteus package should add items to parameters. Packages are only expected
// to read them.
func addToParamInfo(
	p sources.Parameters,
	setName, paramName string,
	info sources.ParameterInfo,
) (ok bool) {
	set := p[setName]
	if set == nil {
		set = map[string]sources.ParameterInfo{}
		p[setName] = set
	}

	if _, ok := set[paramName]; ok {
		return false
	}

	set[paramName] = info
	return true
}

// AddAll adds all items on the second parameter to the first. If items already
// exist an error of type types.ErrViolation is returned.
//
// This is not implemented as a method of "Parameters" because only the
// proteus package should add items to parameters. Packages are only expected
// to read them.
func addAllToParamInfo(
	destination, source sources.Parameters,
	setName, paramName string,
	info sources.ParameterInfo,
) error {
	var violations types.ErrViolations

	for setName, set := range source {
		for paramName, info := range set {
			if ok := addToParamInfo(destination, setName, paramName, info); !ok {
				violations = append(violations, types.Violation{
					SetName:   setName,
					ParamName: paramName,
					Message:   "Trying to register the same parameter twice",
				})
			}
		}
	}

	if len(violations) > 0 {
		return violations
	}

	return nil
}
