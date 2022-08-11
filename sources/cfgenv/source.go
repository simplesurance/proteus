// Package cfgenv is a parameter provider that reads values from environment
// variables.
//
// In order to be able to perform the strongest possible validation of the
// provided environment variables, it is expected that all variables used
// to configure an application have common prefix, and all variables with
// that prefix must match a configuration parameter. This allows detecting
// small mistakes like typos when specifying the name of variable.
//
// The rule for determining the environment variable that configures a
// parameter is as follows:
//
//	{prefix}__{paramname}            (if parameter is not on a set)
//	{prefix}__{setname}__{paramname} (if parameter is on a set)
//	replace "-" by "_"
//	uppercase
//
// For example, if the prefix is "cfg":
//
//	param=test1-parameter           => env=CFG__TEST1_PARAMETER
//	param=test2_parameter           => env=CFG__TEST2_PARAMETER
//	param=address          set=http => env=CFG__HTTP__ADDRESS
//
// Note that both "-" and "_" are mapped to "_". For this reason, if one
// application has two parameters that are differentiated only by this
// character, it can't be configured using this configuration provider.
package cfgenv

import (
	"fmt"
	"os"
	"strings"

	"github.com/simplesurance/proteus/sources"
	"github.com/simplesurance/proteus/types"
)

// New creates a new provider that allows configuring parameters using
// environment variables. See package description for details.
func New(prefix string) sources.Provider {
	return &envVarProvider{prefix: prefix}
}

type envVarProvider struct {
	prefix string
}

func (r *envVarProvider) IsCommandLineFlag() bool {
	return false
}

func (r *envVarProvider) Stop() {
}

func (r *envVarProvider) Watch(
	paramIDs sources.Parameters,
	updater sources.Updater,
) (initial types.ParamValues, _ error) {
	return parse(updater.Log, r.prefix+"__", paramIDs)
}

func parse(
	logger func(string),
	prefix string,
	paramIDs sources.Parameters,
) (types.ParamValues, error) {
	env := readEnvVarsWithPrefix(logger, prefix)

	ret := types.ParamValues{}
	for setName, set := range paramIDs {
		for paramName := range set {
			envName := envVarName(setName, paramName, prefix)
			value, ok := env[envName]
			if !ok {
				continue
			}

			set, ok := ret[setName]
			if !ok {
				set = map[string]string{}
				ret[setName] = set
			}

			set[paramName] = value

			delete(env, envName)
		}
	}

	var violations types.ErrViolations
	for envName := range env {
		violations = append(violations, types.Violation{
			Message: fmt.Sprintf(
				"Environment variable %q has the %q prefix, but is does not match any expected application parameter",
				envName, prefix),
		})
	}

	if len(violations) > 0 {
		return ret, violations
	}

	return ret, nil
}

func readEnvVarsWithPrefix(logger logger, prefix string) map[string]string {
	envSlice := os.Environ()

	ret := map[string]string{}
	for _, env := range envSlice {
		envName, envVal, _ := strings.Cut(env, "=")
		if !strings.HasPrefix(envName, prefix) {
			continue
		}

		ret[envName] = envVal
	}

	return ret
}

// envVarName produces the name of the environment variable that should be used
// to configure a parameter. This function is not reversible.
func envVarName(setName, valueName, prefix string) string {
	var ret string
	if setName == "" {
		ret = prefix + valueName
	} else {
		ret = prefix + setName + "__" + valueName
	}

	return strings.ToUpper(strings.ReplaceAll(ret, "-", "_"))
}

type logger func(string)
