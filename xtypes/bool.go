package xtypes

import (
	"errors"
	"strconv"
	"sync"

	"github.com/simplesurance/proteus/types"
)

// Bool is a boolean parameter that can be updated without restarting
// the application. An UpdateFn must be provided to get updates for the value.
type Bool struct {
	DefaultValue bool
	UpdateFn     func(bool)
	content      struct {
		value *bool
		mutex sync.Mutex
	}
}

var _ types.XType = &Bool{}

// UnmarshalParam parses the input as a boolean.
func (d *Bool) UnmarshalParam(in *string) error {
	var ptrBool *bool
	if in != nil {
		boolValue, err := strconv.ParseBool(*in)
		if err != nil {
			return errors.New("not a valid boolean")
		}

		ptrBool = &boolValue
	}

	d.content.mutex.Lock()
	d.content.value = ptrBool
	d.content.mutex.Unlock()

	if d.UpdateFn != nil {
		d.UpdateFn(d.Value())
	}

	return nil
}

// Value reads the current updated value, taking the default value into
// consideration.
func (d *Bool) Value() bool {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.value == nil {
		return d.DefaultValue
	}

	return *d.content.value
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *Bool) ValueValid(s string) error {
	_, err := strconv.ParseBool(s)
	return err
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *Bool) GetDefaultValue() (string, error) {
	return strconv.FormatBool(d.DefaultValue), nil
}
