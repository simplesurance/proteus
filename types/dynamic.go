package types

// DynamicType is a configuration parameter that can be updated without
// without reloading the application.
//
// Dynamic types must also implement a
// function "Value() T", that takes no parameters and returns an arbitrary
// type T. This method will be used by the application to read the value
// of the parameter, and used by the configuration package to show how to
// use the package. If one dynamic type implements a function
//
//     Value() *url.URL
//
// Then when the user ask for usage information (for example by providing
// --help), the application will show:
//
//     - {fieldName}:*url.URL
//
// Helping the user know that he is expected to provide a URL for the parameter
// named {fieldName}. If for some case the resulting usage information is
// not appropriated, the type can also implement DynamicTypeDescriber to
// change that.
type DynamicType interface {
	UnmarshalParam(*string) error

	// GetDefaultValue reads a string representation of the value. This
	// string representation must be such that, if provided as the
	// parameter value, will give back the original value. If the
	// default value is not valid, details about the error are returned.
	GetDefaultValue() (string, error)

	ValueValid(string) error
}

// Redactor allows part of a value to be redacted.
type Redactor interface {
	// RedactValue redacts part of the value, if it contains sensitive
	// information.
	RedactValue(string) string
}

// DynamicTypeDescriber describes what kind of value must be provided by
// a configuration parameter. Implementation is optional, and should only
// be done when the default generated usage information is not appropriated.
type DynamicTypeDescriber interface {
	DescribeDynamicType() string
}
