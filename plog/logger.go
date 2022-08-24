// Package plog has the types and code used for logs on proteus. Design goals:
//
// - no external libraries
// - as simple as possible
// - allow extending in the future without breaking code
package plog

// D creates a log entry with severity "debug".
func (logger Logger) D(msg string, opts ...Option) {
	o := applyOptions(opts...)
	logger(Entry{
		Severity: SevDebug,
		Message:  msg,
		Caller:   ReadCaller(o.skipCallers),
	})
}

// I creates a log entry with severity "info".
func (logger Logger) I(msg string, opts ...Option) {
	o := applyOptions(opts...)
	logger(Entry{
		Severity: SevInfo,
		Message:  msg,
		Caller:   ReadCaller(o.skipCallers),
	})
}

// E creates a log entry with severity "error".
func (logger Logger) E(msg string, opts ...Option) {
	o := applyOptions(opts...)
	logger(Entry{
		Severity: SevError,
		Message:  msg,
		Caller:   ReadCaller(o.skipCallers),
	})
}
