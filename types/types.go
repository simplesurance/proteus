// Package types defines types used by proteus and by code using it.
//
//nolint:revive
package types

import "errors"

// ErrNoValue can be returned by xtypes on their ValueValid method to indicate
// that the provided value should be considered as if no value was provided at
// all. This allows, for example, to handle empty strings as "no value"
var ErrNoValue = errors.New("no value provided")
