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
//
// The stack trace can optionally be included. Since this is an expensive
// operation, this should not be liberally used.
func ReadCaller(skip int, withStackTrace bool) *Caller {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	ret := &Caller{
		File:       file,
		LineNumber: line,
	}

	if withStackTrace {
		buffer := make([]byte, 4*1024)
		read := runtime.Stack(buffer, false)
		ret.Stacktrace = string(buffer[:read])
	}

	return ret
}

// Caller holds information about the caller who created the log message.
type Caller struct {
	File       string `json:"file,omitempty"`
	LineNumber int    `json:"line,omitempty"`
	Stacktrace string `json:"stacktrace,omitempty"`
}
