package xtypes

import (
	"errors"
	"net/url"
	"sync"

	"github.com/simplesurance/proteus/types"
)

// URL is a xtype for values of type *url.URL.
type URL struct {
	DefaultValue *url.URL
	UpdateFn     func(*url.URL)
	ValidateFn   func(*url.URL) error
	content      struct {
		value *url.URL
		mutex sync.Mutex
	}
}

var _ types.XType = &URL{}
var _ types.Redactor = &URL{}

// UnmarshalParam parses the input as a string.
func (d *URL) UnmarshalParam(in *string) error {
	var url *url.URL
	if in != nil {
		var err error
		url, err = parseURL(*in, d.ValidateFn)
		if err != nil {
			return err
		}
	}

	d.content.mutex.Lock()
	d.content.value = url
	d.content.mutex.Unlock()

	if d.UpdateFn != nil {
		d.UpdateFn(d.Value())
	}

	return nil
}

// Value reads the current updated value, taking the default value into
// consideration. If the parameter is not marked as optional, this is
// guaranteed to be not nil.
func (d *URL) Value() *url.URL {
	d.content.mutex.Lock()
	defer d.content.mutex.Unlock()

	if d.content.value == nil {
		return d.DefaultValue
	}

	return d.content.value
}

// ValueValid test if the provided parameter value is valid. Has no side
// effects.
func (d *URL) ValueValid(s string) error {
	_, err := parseURL(s, d.ValidateFn)
	return err
}

// GetDefaultValue will be used to read the default value when showing usage
// information.
func (d *URL) GetDefaultValue() (string, error) {
	if d.DefaultValue == nil {
		return "", nil
	}

	return formatURL(d.DefaultValue, true), nil
}

// RedactValue is used for URLs of this type to redact themselves, avoiding
// leaking secrets when dumping parameter values.
func (d *URL) RedactValue(in string) string {
	url, err := parseURL(in, d.ValidateFn)
	if err != nil {
		return in
	}

	return formatURL(url, true)
}

func parseURL(in string, validateFn urlValidateFn) (*url.URL, error) {
	parsed, err := url.Parse(in)
	if err != nil {
		return nil, err
	}

	if validateFn == nil {
		validateFn = defaultURLValidateFn
	}

	if err := validateFn(parsed); err != nil {
		return nil, err
	}

	return parsed, nil
}

func formatURL(in *url.URL, redact bool) string {
	if redact {
		return in.Redacted()
	}

	return in.String()
}

type urlValidateFn func(*url.URL) error

func defaultURLValidateFn(u *url.URL) error {
	if u.Scheme == "" {
		return errors.New("scheme of the URL is required")
	}

	if u.Host == "" {
		return errors.New("host of the URL is required")
	}

	return nil
}

// MustParseURL parses the URL. If parsing fails, it panics. It is being
// provided to make it straightforward to provide a default value to
// xtypes.URL, since the "url" package has only a parsing function that returns
// URL and error, resulting in cumbersome code, specially when needing to
// specify multiple default URLs.
func MustParseURL(v string) *url.URL {
	ret, err := url.Parse(v)
	if err != nil {
		panic(err)
	}

	return ret
}
