package proteus

import (
	"github.com/simplesurance/proteus/sources"
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
