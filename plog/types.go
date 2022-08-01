package plog

import (
	"runtime"
)

// Logger is the function used to output human-readable diagnostics
// information. Depth can optionally be used to determine the real caller of
// the log function, by skipping the correct number of intermediate frames
// in the stacktrace.
type Logger func(Entry)

// Entry is a log entry.
type Entry struct {
	Severity Severity `json:"severity,omitempty"`
	Message  string   `json:"message,omitempty"`
	Caller   *Caller  `json:"caller,omitempty"`
}

// ReadCaller reads information about the caller, skipping the provided
// number of callers. Skip=1 means the immediate caller, skip=2 is the
// caller of the caller, and so on.
func ReadCaller(skip int) *Caller {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	return &Caller{
		File:       file,
		LineNumber: line,
	}
}

// Caller holds information about the caller who created the log message.
type Caller struct {
	File       string
	LineNumber int
}
