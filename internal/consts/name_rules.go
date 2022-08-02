package consts

import "regexp"

const (
	// RedactedPlaceholder is a string that is shown instead of a secret, to
	// avoid leaking it.
	RedactedPlaceholder = "REDACTED"
)

// ParamNameRE is the regular expression used to valid parameter and parameter
// set names. A name must:
//
// 1. contains only a-z or 0-9 or "_" or "-"
// 2. start with a-z
// 3. not end with "-" or "_" (can have other characters on 1)
// 4. not have two letters in a row being "_" or "-"
var ParamNameRE = regexp.MustCompile(`^[a-z](?:[_-]?[a-z0-9])*$`)
