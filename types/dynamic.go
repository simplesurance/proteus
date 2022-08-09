package types

// XType is a configuration parameter parameter that supports being updated
// without restarting the application, as well as other features each type
// might decide to support.
//
// XTypes types must not only implement this interface, but also implement a
// function "Value() T", that takes no parameters and returns an arbitrary
// type T. Example of a Value function:
//
//	Value() *url.URL
//
// The returned value, when not immutable, will be a copy of the parameter
// value. The value returned by this function might change, if its value
// was read from a source that supports updating values without application
// restart, like file.
//
// When proteus responds to "--help" by generating usage instructions it
// will tell the user the _type_ of the parameter. For an XType, by default
// the return type of Value() is used. For example:
//
//	params: struct{
//		Server *xtype.URL
//	}{}
//
// will be displayed on help usage as:
//
//	-server <url.URL>
//
// Because Value() for xtype.URL returns an URL. If this is not appropriated
// in some use-case, the xtype may also implement TypeDescriber, that allow
// providing an arbitrary string to be used.
//
// XTypes may optionally implement the Redactor interface.
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

// TypeDescriber overrides the "type" of a parameter when showing usage
// information to the user. When an xtype does not implement this interface,
// the return type of the Value() function is used. Implementing this function
// allows providing an arbitrary string to be used instead. One example is the
// xtype.OneOf type:
//
//	params := struct{
//			Region *xtypes.OneOf
//		}{
//			Region: &xtypes.OneOf{
//				Choices: []string{"EU", "US"},
//			},
//	}
//
// The OneOf xtype implement this method to show the actual list of choices,
// instead of just showing "string":
//
//	-region <EU|US>
type TypeDescriber interface {
	DescribeType() string
}
