package types

// XType is a configuration parameter parameter that supports being updated
// without restarting the application, as well as other features each type
// might decide to support.
//
// XTypes types must not only implement this interface, but also implement a
// function "Value() T", that takes no parameters and returns an arbitrary
// type T. This method will be used by the application to read the value
// of the parameter, and used by the configuration package to show how to
// use the package. Example of a Value function:
//
//     Value() *url.URL
//
// The returned value, when not immutable, will be a copy of the parameter
// value. The value returned by this function might change, if its value
// was read from a source that supports updating values without application
// restart, like file.
//
// Values of this type may optionally implement the Redactor or the
// TypeDescriber inteface (or both).
type XType interface {
	UnmarshalParam(*string) error

	// GetDefaultValue reads a string representation of the value. This
	// string representation must be such that, if provided as the
	// parameter value, will give back the original value. If the
	// default value is not valid, details about the error are returned.
	GetDefaultValue() (string, error)

	ValueValid(string) error
}

// Redactor allows part of a value to be redacted. One example of a use for
// a redactor is URL, which can automatically redact the password part of the
// URL.
type Redactor interface {
	// RedactValue redacts part of the value, if it contains sensitive
	// information.
	RedactValue(string) string
}

// TypeDescriber describes what kind of value must be provided by
// a configuration parameter. Implementation is optional, and should only
// be implemented when the default generated usage information is not
// appropriated.
type TypeDescriber interface {
	DescribeType() string
}
