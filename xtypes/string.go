package xtypes

import (
	"sync"

	"github.com/simplesurance/proteus/types"
)

// String is an XType for strings.
type String struct {
	DefaultValue string
	UpdateFn     func(string)
	content      struct {
		value *string
		mutex sync.Mutex
	}
}

var _ types.XType = &String{}

// UnmarshalParam parses the input as a string.
func (d *String) UnmarshalParam(in *string) error {
	var ptrStr *string
	if in != nil {
		strValue := *in // copy
		ptrStr = &strValue
	}

	d.content.mutex.Lock()
	d.content.value = ptrStr
	d.content.mutex.Unlock()

	if d.UpdateFn != nil {
		d.UpdateFn(d.Value())
	}

	return nil
}

// Value reads the current updated value, taking the default value into
// consideration.
func (d *String) Value() string {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.value == nil {
		return d.DefaultValue
	}

	return *d.content.value
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *String) ValueValid(s string) error {
	return nil
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *String) GetDefaultValue() (string, error) {
	return d.DefaultValue, nil
}
