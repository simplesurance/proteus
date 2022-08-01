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
	protected     struct {
		valuesMutex sync.Mutex
		values      []types.ParamValues
	}
}

// ErrUsage is a specialized version of Usage(), that is intended
// to be used when configuration parsing failed. Additionally to the
// usage text, it also outputs the validation errors with the supplied
// parameters. It does not terminate the application.
func (p *Parsed) ErrUsage(w io.Writer, err error) {
	// TODO: the output here can be a lot more insightful
	fmt.Fprintf(w, "%s: %s\n", binaryName(), err.Error())
	p.Usage(w)
}

// Usage print usage information to the provided writer.
func (p *Parsed) Usage(w io.Writer) {
	setKeys := make([]string, 0, len(p.inferedConfig))
	for k := range p.inferedConfig {
		setKeys = append(setKeys, k)
	}

	sort.Strings(setKeys)

	paramDoc := strings.Builder{}
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

		fmt.Fprintln(&paramDoc)
		if setName == "" {
			fmt.Fprintln(&paramDoc, "PARAMETERS")
		} else {
			fmt.Fprintln(&paramDoc, "PARAMETER SET: "+strings.ToUpper(setName))
			if set.desc != "" {
				fmt.Fprintln(&paramDoc, set.desc)
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

			// special parameters come first
			if p1.isSpecial != p2.isSpecial {
				return p1.isSpecial
			}

			// then mandatory fields
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

			opts := []string{fmt.Sprintf("- %s", name)}
			if field.secret {
				opts = append(opts, "secret")
			}

			if field.optional {
				opts = append(opts, "default="+field.redactedDefaultValue())
			}

			fmt.Fprintln(&paramDoc, strings.Join(opts, " "))

			if field.desc != "" {
				fmt.Fprintf(&paramDoc, "  %s\n", field.desc)
			}
		}
	}

	if p.settings.onelineDesc != "" {
		fmt.Fprintln(w, p.settings.onelineDesc)
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, strings.Join(cmdLine, " \\\n  "))

	fmt.Fprintln(w, paramDoc.String())
}

// Dump prints the names and values of the parameters.
func (p *Parsed) Dump(w io.Writer) {
	p.protected.valuesMutex.Lock()
	defer p.protected.valuesMutex.Unlock()

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

// Valid allows determining if the provided application parameters are valid.
func (p *Parsed) Valid() error {
	p.protected.valuesMutex.Lock()
	defer p.protected.valuesMutex.Unlock()

	return p.valid()
}

// Stop release resources being used. Proteus itself does not use any
// resource that need to be released, but some providers might.
func (p *Parsed) Stop() {
	for _, p := range p.settings.providers {
		p.Stop()
	}
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
//
// Caller must hold the protected.mutex.
func (p *Parsed) mergeValues() types.ParamValues {
	ret := types.ParamValues{}
	for _, providerValues := range p.protected.values {
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
					Message:   "parameter is required but was not specified",
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
//
// Caller must hold the mutex.
func (p *Parsed) refresh(force bool) {
	if err := p.valid(); err != nil {
		p.settings.loggerFn.E(fmt.Sprintf(
			"Refusing to update values because configuration is invalid: %v",
			err.Error()))
		return
	}

	for setName, set := range p.inferedConfig {
		for paramName, paramConfig := range set.fields {
			if !paramConfig.isXtype && !force {
				p.settings.loggerFn.D(fmt.Sprintf(
					"Not updating %s.%s (xtype: %t, force: %t)",
					setName, paramName, paramConfig.isXtype, force))

				continue
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
				p.settings.loggerFn.E(fmt.Sprintf(
					"error updating %s.%s: %v",
					setName, paramName, err))
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
	for _, providerData := range p.protected.values {
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

func binaryName() string {
	_, ret := filepath.Split(os.Args[0])
	return ret
}

func formatCmdLineParam(cmd string, field paramSetField) string {
	content := fmt.Sprintf("-%s <%s>", cmd, field.typ)
	if field.boolean {
		content = fmt.Sprintf("-%s", cmd)
	}

	if field.optional {
		return fmt.Sprintf("[%s]", content)
	}

	return content
}

func mapKeysSorted[T any](v map[string]T) []string {
	ret := make([]string, 0, len(v))
	for k := range v {
		ret = append(ret, k)
	}

	sort.Strings(ret)
	return ret
}
