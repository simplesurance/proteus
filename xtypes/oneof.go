package xtypes

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/simplesurance/proteus/internal/slices"
	"github.com/simplesurance/proteus/types"
)

// OneOf is a parameter configuration that can hold a value from a list of
// valid options. The list of options must be provided. An UpdateFn can be
// provided to be notified about changes to the value.
type OneOf struct {
	Choices      []string
	DefaultValue string
	IgnoreCase   bool
	UpdateFn     func(string)
	content      struct {
		valueIx *int
		mutex   sync.Mutex
	}
}

var _ types.XType = &OneOf{}
var _ types.TypeDescriber = &OneOf{}

// UnmarshalParam is a custom parser for a string parameter. This will always
// run on brand new instance of string, so no synchronization is necessary.
func (d *OneOf) UnmarshalParam(in *string) error {
	var valueIxPtr *int
	if in != nil {
		for ix, opt := range d.Choices {
			if d.compare(opt, *in) {
				cpix := ix
				valueIxPtr = &cpix
				break
			}
		}
	}

	if valueIxPtr == nil {
		return errors.New("value must be one of: " +
			strings.Join(d.Choices, "|"))
	}

	d.content.mutex.Lock()
	d.content.valueIx = valueIxPtr
	d.content.mutex.Unlock()

	if d.UpdateFn != nil {
		d.UpdateFn(d.Value())
	}

	return nil
}

// Value reads the current updated value.
func (d *OneOf) Value() string {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.valueIx == nil {
		return d.DefaultValue
	}

	return d.Choices[*d.content.valueIx]
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *OneOf) ValueValid(s string) error {
	if !slices.Contains(d.Choices, s, d.compare) {
		return fmt.Errorf("value must be one of %s",
			strings.Join(d.Choices, "|"))
	}

	return nil
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *OneOf) GetDefaultValue() (string, error) {
	for ix := range d.Choices {
		if d.Choices[ix] == d.DefaultValue {
			return d.DefaultValue, nil
		}
	}

	return "", fmt.Errorf(
		"provided default value is not on the list of choices [%s]",
		strings.Join(d.Choices, ", "))
}

// DescribeType changes how usage information is shown for parameters
// of this type. Instead of showing the default:
//
//	{paramName}:string
//
// this method will make it show as
//
//	{paramName}:(option1|option2|...)
func (d *OneOf) DescribeType() string {
	return "" + strings.Join(d.Choices, "|")
}

func (d *OneOf) compare(v1, v2 string) bool {
	if d.IgnoreCase {
		return strings.EqualFold(v1, v2)
	}

	return v1 == v2
}
