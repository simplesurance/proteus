package xtypes

import (
	"encoding/json"
	"sync"

	"github.com/simplesurance/proteus/types"
)

// RawJSON is a xtype for raw json messages.
type RawJSON struct {
	DefaultValue json.RawMessage
	UpdateFn     func(*json.RawMessage)
	content      struct {
		value json.RawMessage
		mutex sync.Mutex
	}
}

var _ types.XType = &RawJSON{}

// UnmarshalParam parses the input as a string.
func (d *RawJSON) UnmarshalParam(in *string) error {
	var j json.RawMessage
	if in != nil {
		err := json.Unmarshal([]byte(*in), &j)
		if err != nil {
			return err
		}
	}

	d.content.mutex.Lock()
	d.content.value = j
	d.content.mutex.Unlock()

	if d.UpdateFn != nil {
		cp := d.Value()
		d.UpdateFn(&cp)
	}

	return nil
}

// Value reads the current updated value, taking the default value into
// consideration. If the parameter is not marked as optional, this is
// guaranteed to be not nil.
func (d *RawJSON) Value() json.RawMessage {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.value == nil {
		// return a copy
		cp := make(json.RawMessage, len(d.DefaultValue))
		copy(cp, d.DefaultValue)
		return cp
	}

	cp := make(json.RawMessage, len(d.content.value))
	copy(cp, d.content.value)
	return cp
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *RawJSON) ValueValid(s string) error {
	var j json.RawMessage
	return json.Unmarshal([]byte(s), &j)
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *RawJSON) GetDefaultValue() (string, error) {
	if d.DefaultValue == nil {
		return "", nil
	}

	return string(d.DefaultValue), nil
}
