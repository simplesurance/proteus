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
//
// The list of choices is mandatory, and must be provided in the "Choices"
// field.
//
// If the parameter is not optional it is mandatory to provide a default
// value that matches one of the choices. For mandatory parameters, a
// default value don't have to be provided, and is ignored.
//
// Example:
//
//	params := struct{
//		Region: *xtypes.OneOf
//	}{
//		Region: &xtypes.OneOf{
//			Choices:      []string{"EU", "US"},
//			DefaultValue: "EU",
//		},
//	}
type OneOf struct {
	Choices      []string
	DefaultValue string
	IgnoreCase   bool
	UpdateFn     func(string)
	content      struct {
		value *string
		mutex sync.Mutex
	}
}

var _ types.XType = &OneOf{}
var _ types.TypeDescriber = &OneOf{}

// UnmarshalParam is a custom parser for a string parameter. This will always
// run on brand new instance of string, so no synchronization is necessary.
func (d *OneOf) UnmarshalParam(in *string) error {
	var newValue *string
	if in != nil {
		ok := false
		for _, opt := range d.Choices {
			if d.compare(opt, *in) {
				ok = true
				cpin := *in
				newValue = &cpin
				break
			}
		}

		if !ok {
			return errors.New("value must be one of: " +
				strings.Join(d.Choices, "|"))
		}
	}

	d.content.mutex.Lock()
	d.content.value = newValue
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

	if d.content.value == nil {
		return d.DefaultValue
	}

	return *d.content.value
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
		"default value for OneOf is not in the list of choices [%s]",
		strings.Join(d.Choices, "|"))
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
