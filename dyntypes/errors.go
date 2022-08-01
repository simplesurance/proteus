package dyntypes

import "errors"

// ErrRequired indicates that a required parameter was not provided.
var ErrRequired = errors.New("required")
