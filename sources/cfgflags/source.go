// Package cfgflags implements a configuration reader that reads from
// command-line flags.
//
// Examples of parameters:
//
// Providing two parameters:
//
//	./binary -param1 value1 -param2 value2
//
// Values can optionally be provided in a key=value format. The two styles can
// be interchanged freely:
//
//	./binary -param1=value1 -param2 value2
//
// Boolean flags are a special case. They can be provided in one of two ways:
//
//	./binary -flag
//	./binary -flag=<true|false>
//
// Boolean flags CANNOT be provided using "-flag <true|false>".
//
// Parameter sets can be provided. For example, to provide a set of parameters
// for "http" and a set of parameters for "grpc", use:
//
//	./binary http -addr :8080 -max-connections 64 -enabled grpc -addr :6800 -enabled=true
package cfgflags

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/simplesurance/proteus/internal/consts"
	"github.com/simplesurance/proteus/sources"
	"github.com/simplesurance/proteus/types"
)

var (
	errIsSetName = errors.New("is a set name, not a parameter")
)

// New creates a new provider that reads from command-line flags.
// See package documentation for details on how to provide command-line
// parameters.
func New() sources.Provider {
	return &flagProvider{}
}

type flagProvider struct{}

func (r *flagProvider) IsCommandLineFlag() bool {
	return true
}

func (r *flagProvider) Stop() {
}

func (r *flagProvider) Watch(
	paramIDs sources.Parameters,
	updater sources.Updater,
) (initial types.ParamValues, err error) {
	ret := types.ParamValues{}

	var ix = 1
	set := map[string]string{}
	var setName string
	for {
		token, ok := readToken(&ix)
		if !ok {
			break
		}

		isBoolFn := func(paramName string) (isBool, ok bool) {
			if set, ok := paramIDs[setName]; ok {
				if paramInfo, ok := set[paramName]; ok {
					return paramInfo.IsBool, true
				}
			}
			return false, false
		}

		paramName, paramValue, err := readParam(&ix, token, isBoolFn)
		if err != nil {
			if errors.Is(err, errIsSetName) {
				newSetName := token
				// store all data from previous flagset. Validate it
				// first: flag sets must have attributes; attributes
				// without flagset have a setName="", and can
				// be empty.
				if setName != "" && len(set) == 0 {
					return nil, fmt.Errorf("flagset %q has no parameters", setName)
				}

				ret[setName] = set
				set = map[string]string{}

				setName = newSetName
				continue
			}

			if setName == "" {
				return nil, err
			}

			return nil, fmt.Errorf("parsing flagset %q: %w", setName, err)
		}

		set[paramName] = paramValue
	}

	if setName != "" && len(set) == 0 {
		return nil, fmt.Errorf("flagset %q has no parameters", setName)
	}

	if len(set) > 0 {
		ret[setName] = set
	}

	return ret, nil
}

// readParam reads the parameter key and value from token, possibly reading
// more tokens from the parameters. If it finds a flagset, returns
// errIsSetName
func readParam(
	ix *int,
	token string,
	isBoolFn func(paramName string) (isBool, ok bool),
) (key, value string, _ error) {
	if !strings.HasPrefix(token, "-") {
		// is a flagset
		if !consts.ParamNameRE.MatchString(token) {
			return "", "", fmt.Errorf("not a parameter, not a valid flagset name: %q", token)
		}

		return "", "", errIsSetName
	}

	token = strings.TrimPrefix(token, "-")
	token = strings.TrimPrefix(token, "-") // "--" handled the same as "-"

	// try to parse as key=value
	paramName, value, ok := strings.Cut(token, "=")
	if ok {
		// format of token is key=value
		if !consts.ParamNameRE.MatchString(paramName) {
			return "", "", fmt.Errorf(
				"%q is not valid for a parameter or flagset name (valid=%s)",
				paramName, consts.ParamNameRE)
		}

		return paramName, value, nil
	}

	paramName = token

	if !consts.ParamNameRE.MatchString(paramName) {
		return "", "", fmt.Errorf(
			"%q is not valid for a parameter or flagset name (valid=%s)",
			paramName, consts.ParamNameRE)
	}

	isBool, found := isBoolFn(paramName)
	if !found {
		// without knowing if the parameter is boolean or not, it is not
		// possible to determine how to process the remaining
		// flags. For example, the command line:
		//
		//    ./bin -a b -c d
		//
		// if "a" is bool, then:
		// - a=true
		// - b is flagset:
		//   - c=d
		//
		// but if "a" is not bool, then:
		// - a=b
		// - c=d
		//
		// For this reason, flags won't be processed further.
		return "", "", fmt.Errorf(
			"provided parameter $%d=%q is not expected by the application; parameters after this position will not be processed",
			*ix-1, paramName)
	}

	if isBool {
		return paramName, "true", nil
	}

	// token has only param name, must read value from another
	// token to get the value
	value, ok = readToken(ix)
	if !ok {
		return "", "", fmt.Errorf("parameter %q has no value", paramName)
	}

	return paramName, value, nil
}

func readToken(ix *int) (string, bool) {
	if *ix >= len(os.Args) {
		return "", false
	}

	ret := os.Args[*ix]
	*ix++
	return ret, true
}
