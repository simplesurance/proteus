package proteus

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/simplesurance/proteus/types"
)

const redactedPlaceholder = "<redacted>"

// Parsed holds information about all parameters supported by the application,
// and their options, allowing interacting with them.
type Parsed struct {
	settings      settings
	inferedConfig config
	valuesMutex   sync.Mutex
	values        []types.ParamValues
}

// Usage print usage information to the provider writer.
func (p *Parsed) Usage(w io.Writer) {
	p.usage(w, nil)
}

// ErrUsage is a specialized version of Usage(), that is intended
// to be used when parsing the configuration failed and
// will show to the user what is wrong with the provided parameters.
// It does not terminate the application.
func (p *Parsed) ErrUsage(w io.Writer, err error) {
	fmt.Fprintln(w, "Invalid configuration parameters for application")
	if err != nil {
		fmt.Fprintln(w, err.Error())
	}

	p.usage(w, err)
}

func (p *Parsed) usage(w io.Writer, err error) {
	setKeys := make([]string, 0, len(p.inferedConfig))
	for k := range p.inferedConfig {
		setKeys = append(setKeys, k)
	}

	sort.Strings(setKeys)

	details := strings.Builder{}
	cmdLine := []string{binaryName()}

	lastSet := ""
	for _, setName := range setKeys {
		set := p.inferedConfig[setName]

		if lastSet != setName {
			lastSet = setName
			cmdLine = append(cmdLine, setName)
		}

		if len(set.fields) == 0 {
			continue
		}

		fmt.Fprintln(&details)
		if setName == "" {
			fmt.Fprintln(&details, "PARAMETERS")
		} else {
			fmt.Fprintln(&details, "PARAMETER SET: "+strings.ToUpper(setName))
			if set.desc != "" {
				fmt.Fprintln(&details, set.desc)
			}
		}

		// sort by parameter names
		paramNames := make([]string, 0, len(set.fields))
		for k := range set.fields {
			paramNames = append(paramNames, k)
		}

		sort.Slice(paramNames, func(i, j int) bool {
			p1 := set.fields[paramNames[i]]
			p2 := set.fields[paramNames[j]]

			// mandatory fields first
			if p1.optional != p2.optional {
				return p2.optional
			}

			// then lexicographic order
			return paramNames[i] < paramNames[j]
		})

		// describe parameter name, type and options
		for _, name := range paramNames {
			field := set.fields[name]

			cmdLine = append(cmdLine, "  "+formatCmdLineParam(name, field))

			opts := []string{fmt.Sprintf("- %s:%s", name, field.typ)}
			if field.optional {
				opts = append(opts, "default="+field.redactedDefaultValue())
			}

			fmt.Fprintln(&details, strings.Join(opts, " "))

			if field.desc != "" {
				fmt.Fprintf(&details, "  %s\n", field.desc)
			}
		}
	}

	fmt.Fprintln(w, "Syntax:")
	fmt.Fprintln(w, strings.Join(cmdLine, " \\\n  "))
	fmt.Fprintln(w, details.String())
}

func binaryName() string {
	_, ret := filepath.Split(os.Args[0])
	return "./" + ret
}

func formatCmdLineParam(cmd string, field paramSetField) string {
	content := fmt.Sprintf("-%s %s", cmd, field.typ)
	if field.boolean {
		content = "-" + cmd
	}

	if field.optional {
		return fmt.Sprintf("[%s]", content)
	}

	return fmt.Sprintf("<%s>", content)
}

// Dump prints the names and values of the parameters.
func (p *Parsed) Dump(w io.Writer) {
	fmt.Fprintf(w, "Parameter values:\n")
	merged := p.mergeValues()
	for _, setName := range mapKeysSorted(merged) {
		if setName != "" {
			fmt.Fprintf(w, "\nPARAMETER SET %s:\n", strings.ToUpper(setName))
		}

		set := merged[setName]
		for _, paramName := range mapKeysSorted(set) {
			value := set[paramName]
			redacted := p.inferedConfig[setName].fields[paramName].redactedValue(&value)()
			fmt.Fprintf(w, "- %s = %q\n", paramName, redacted)
		}
	}
}

func mapKeysSorted[T any](v map[string]T) []string {
	ret := make([]string, 0, len(v))
	for k := range v {
		ret = append(ret, k)
	}

	sort.Strings(ret)
	return ret
}

// Valid allows determining if the provided application parameters are valid.
func (p *Parsed) Valid() error {
	p.valuesMutex.Lock()
	defer p.valuesMutex.Unlock()

	return p.valid()
}

// valid determines if the desired parameters are valid.
// Caller must hold the mutex.
func (p *Parsed) valid() error {
	mergedValues := p.mergeValues()

	violations := types.ErrViolations{}
	for setName, set := range p.inferedConfig {
		for paramName, paramConfig := range set.fields {
			// must validate when a value is present and when it
			// is missing (value=nil)
			var value *string
			if setValues, ok := mergedValues[setName]; ok {
				if paramValue, ok := setValues[paramName]; ok {
					value = &paramValue
				}
			}

			if err := p.validValue(setName, paramName, value); err != nil {
				var validViol types.ErrViolations
				if errors.As(err, &validViol) {
					violations = append(violations, validViol...)
					continue
				}

				violations = append(violations, types.Violation{
					SetName:   setName,
					ParamName: paramName,
					ValueFn:   paramConfig.redactedValue(value),
					Message:   err.Error(),
				})
			}
		}
	}

	if len(violations) > 0 {
		return violations
	}

	return nil
}

// mergeValues compute the configuration from all providers, taking provider
// priority into consideration.
// Caller must hold the mutex.
func (p *Parsed) mergeValues() types.ParamValues {
	ret := types.ParamValues{}
	for _, providerValues := range p.values {
		for setName, set := range providerValues {
			retSet, ok := ret[setName]
			if !ok {
				retSet = map[string]string{}
				ret[setName] = retSet
			}

			for paramName, value := range set {
				if _, ok := retSet[paramName]; !ok {
					retSet[paramName] = value
				}
			}
		}
	}

	return ret
}

// validValue test if a value is valid for a given parameter. It has no
// side effects.
func (p *Parsed) validValue(setName, paramName string, value *string) error {
	set, ok := p.inferedConfig[setName]
	if !ok {
		return fmt.Errorf("set %q does not exist", setName)
	}

	param, ok := set.fields[paramName]
	if !ok {
		return fmt.Errorf("param %s.%s does not exit", setName, paramName)
	}

	if value == nil {
		if !param.optional {
			return types.ErrViolations([]types.Violation{
				{
					SetName:   setName,
					ParamName: paramName,
					ValueFn:   param.redactedValue(nil),
					Message:   "Required",
				},
			})
		}

		return nil
	}

	err := param.validFn(*value)
	if err != nil {
		return types.ErrViolations([]types.Violation{
			{
				SetName:   setName,
				ParamName: paramName,
				ValueFn:   param.redactedValue(value),
				Message:   err.Error(),
			},
		})
	}

	return nil
}

// refresh reads the available parameter values that are stored on "parsed"
// and use them to update the configuration struct.
func (p *Parsed) refresh(force bool) {
	p.valuesMutex.Lock()
	defer p.valuesMutex.Unlock()

	if err := p.valid(); err != nil {
		p.settings.loggerFn(fmt.Sprintf(
			"Refusing to update values because configuration is invalid: %v",
			err.Error()), 1)
		return
	}

	for setName, set := range p.inferedConfig {
		for paramName, paramConfig := range set.fields {
			if !paramConfig.isXtype && !force {
				p.settings.loggerFn(fmt.Sprintf("Not updating %s.%s", setName, paramName), 1)
			}

			value := p.desiredValue(setName, paramName)

			if !paramConfig.isXtype && value == nil {
				// value=nil represents the default value. For
				// non-xtype the approach is not touching
				// whatever value is already present on the
				// configuration struct. For xtype values
				// only we may set the value to something, then
				// revert it back to the default value.
				continue
			}

			err := paramConfig.setValueFn(value)
			if err != nil {
				p.settings.loggerFn(fmt.Sprintf("error updating %s.%s: %v", setName, paramName, err), 1)
			}
		}
	}
}

// desiredValue returns the value for a parameter from one of the parameter
// providers, respecting priority. If the value is not provider by any of them,
// nil is returned.
// Caller must hold the mutex.
func (p *Parsed) desiredValue(setName, paramName string) *string {
	// the first provider with a value wins
	for _, providerData := range p.values {
		set, ok := providerData[setName]
		if !ok {
			continue
		}

		if value, ok := set[paramName]; ok {
			return &value
		}
	}

	return nil
}
