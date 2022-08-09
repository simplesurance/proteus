package types

import (
	"fmt"
	"sort"
	"strings"
)

// ErrViolations is an error that represent violations to expectations on
// application configuration.
type ErrViolations []Violation

var _ error = ErrViolations{}

func (e ErrViolations) Error() string {
	if len(e) == 1 {
		return e[0].String()
	}

	sort.Slice(e, func(i, j int) bool {
		if c := strings.Compare(e[i].Path, e[j].Path); c != 0 {
			return c < 0
		}

		if c := strings.Compare(e[i].SetName, e[j].SetName); c != 0 {
			return c < 0
		}

		if c := strings.Compare(e[i].ParamName, e[j].ParamName); c != 0 {
			return c < 0
		}

		return e[i].Message < e[j].Message
	})

	ret := strings.Builder{}
	ret.WriteString("multiple invalid parameters:\n")
	for _, violation := range e {
		vStr := violation.String()
		ret.WriteString("- " + vStr + "\n")
	}

	return ret.String()
}

// Violation holds a single violation of the requirements for parameters.
type Violation struct {
	// Path identifies the location of a violation on the configuration
	// struct.
	Path string

	// SetName is the name of the paramset with a violation.
	SetName string

	// ParamName is the name of the configuration parameter affected
	// by the violation.
	ParamName string

	// ValueFn is a function that allows reading the provided parameter
	// value that caused the violation. It automatically redacts
	// secret values.
	ValueFn func() string

	// Message is a description of the violation.
	Message string
}

func (v Violation) String() string {
	var id string

	if v.Path != "" {
		id = v.Path
	} else if v.SetName != "" {
		id = v.SetName + "." + v.ParamName
	} else {
		id = v.ParamName
	}

	if v.ValueFn == nil {
		return fmt.Sprintf("%q: %s", id, v.Message)
	}

	return fmt.Sprintf("%q: %s (parsing %q)", id, v.Message, v.ValueFn())
}
