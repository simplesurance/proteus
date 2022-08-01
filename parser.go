package proteus

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/simplesurance/proteus/specialflags"
	"github.com/simplesurance/proteus/types"
)

// ParamNameRE is the regular expression used to valid parameter and flagset
// names.
var ParamNameRE = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// MustParse creates a new parser for configuration. It panics if there is
// a coding error. If no coding error is found, the parameters are parsed
// from the provided sources and made available on "flags". If the
// parameters are invalid and error is returned. Parsed is always returned,
// and can be used to get details about the violation that caused the
// parameter parsing to fail.
func MustParse(flags any, options ...Option) (*Parsed, error) {
	opts := settings{
		loggerFn: func(msg string, depth int) {}, // nop logger
	}
	opts.apply(options...)

	appConfig, err := mustInferConfigFromValue(flags, opts)
	if err != nil {
		panic(fmt.Errorf("INVALID CONFIGURATION STRUCT: %v", err))
	}

	if len(opts.sources) == 0 {
		panic("NO CONFIGURATION SOURCE WAS PROVIDED")
	}

	ret := Parsed{
		settings:      opts,
		inferedConfig: appConfig,
		values:        make([]types.ParamValues, len(opts.sources)),
	}

	if err := addSpecialFlags(appConfig, &ret, opts); err != nil {
		return &ret, err
	}

	// start watching each configuration item on each provider
	for ix, source := range opts.sources {
		updater := &updater{
			parsed:        &ret,
			providerIndex: ix,
			providerName:  fmt.Sprintf("%T", source)}

		initial, err := source.Watch(appConfig.configIDs(), updater)
		if err != nil {
			return &ret, err
		}

		// callbacks are not invoked on initial load, but only when
		// values are updated
		updater.update(initial, false)
	}

	processSpecialFlags(appConfig, &ret, opts)

	if err := ret.valid(); err != nil {
		return &ret, err
	}

	ret.refresh(true) // update dynamic and standard parameters

	return &ret, nil
}

func mustInferConfigFromValue(value any, opts settings) (config, error) {
	if reflect.ValueOf(value).Kind() != reflect.Ptr {
		return nil, errors.New("configuration struct must be a pointer")
	}

	val := reflect.ValueOf(value)
	if val.IsNil() {
		return nil, errors.New("provided configuration struct is nil")
	}

	val = val.Elem()

	ret := config{"": flagSet{fields: map[string]flagSetField{}}}

	// each member of the configuration struct can be either:
	// - ignored: identified with: param:"-"
	// - parameter set: meaning that is a structure that contains more
	//   parameter. This is identified by: param:",flagset"
	members, err := flatWalk("", "", val)
	if err != nil {
		return nil, fmt.Errorf("walking root fields of the configuration struct: %w", err)
	}

	var violations ErrViolations
	for _, member := range members {
		name, tag, err := parseParam(member.field, member.value)
		if err != nil {
			var paramViolations ErrViolations
			if errors.As(err, &paramViolations) {
				violations = append(violations, paramViolations...)
				continue
			}

			violations = append(violations, Violation{
				Path:    member.Path,
				Message: fmt.Sprintf("error reading struct tag: %v", err),
			})
			continue
		}

		if name == "-" {
			continue
		}

		if !ParamNameRE.MatchString(name) {
			violations = append(violations, Violation{
				Path: member.Path,
				Message: fmt.Sprintf("Name %q is invalid for parameter or set (valid: %s)",
					name, ParamNameRE)})
		}

		if tag.flagSet {
			// is a set or parameters
			d, err := parseParamSet(name, member.Path, member.value)
			if err != nil {
				var setViolations ErrViolations
				if errors.As(err, &setViolations) {
					violations = append(violations, setViolations...)
					continue
				}

				violations = append(violations, Violation{
					Path:    member.Path,
					SetName: name,
					Message: fmt.Sprintf("parsing set: %v", err),
				})
				continue
			}

			d.desc = tag.desc
			ret[name] = d
			continue
		}

		// is a parameter, add to root set
		ret[""].fields[name] = tag
	}

	if len(violations) > 0 {
		return nil, violations
	}

	return ret, nil
}

func parseParamSet(setName, setPath string, val reflect.Value) (flagSet, error) {
	members, err := flatWalk(setName, setPath, val)
	if err != nil {
		return flagSet{}, err
	}

	ret := flagSet{
		fields: make(map[string]flagSetField, len(members)),
	}

	violations := ErrViolations{}
	for _, member := range members {
		paramName, tag, err := parseParam(member.field, member.value)
		if err != nil {
			violations = append(violations, Violation{
				Path:    member.Path,
				Message: err.Error(),
			})
			continue
		}

		if paramName == "-" || tag.flagSet {
			continue
		}

		if !ParamNameRE.MatchString(paramName) {
			violations = append(violations, Violation{
				Path:    member.Path,
				SetName: setName,
				Message: fmt.Sprintf("Name %q is invalid for parameter or set (valid: %s)",
					paramName, ParamNameRE)})
		}

		tag.path = member.Path
		ret.fields[paramName] = tag
	}

	if len(violations) > 0 {
		return ret, violations
	}

	return ret, nil
}

func parseParam(structField reflect.StructField, fieldVal reflect.Value) (
	paramName string,
	_ flagSetField,
	_ error,
) {
	tagParam := structField.Tag.Get("param")
	tagParamParts := strings.Split(tagParam, ",")
	paramName = tagParamParts[0]

	ret := flagSetField{
		typ:  describeType(fieldVal),
		desc: structField.Tag.Get("param_desc"),
	}

	for _, tagOption := range tagParamParts[1:] {
		switch tagOption {
		case "optional":
			ret.optional = true
		case "secret":
			ret.secret = true
		case "flagset":
			ret.flagSet = true
		default:
			return paramName, ret, fmt.Errorf(
				"option '%s' is invalid for tag 'param' in '%s'",
				tagOption,
				tagParam)
		}
	}

	// if the parameter name is not provided in the "param" tag field then
	// the name of the struct member in lowercase is used as parameter
	// name.
	if paramName == "" {
		paramName = strings.ToLower(structField.Name)
	}

	if isDynamicType(structField.Type) {
		if fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
		}

		ret.boolean = describeType(fieldVal) == "bool"

		ret.validFn = toDynamic(fieldVal).ValueValid
		ret.setValueFn = toDynamic(fieldVal).UnmarshalParam
		ret.getDefaultFn = toDynamic(fieldVal).GetDefaultValue
		if redactor := toRedactor(fieldVal); redactor != nil {
			ret.redactFn = redactor.RedactValue
		} else {
			ret.redactFn = func(s string) string { return s }
		}

	} else if !ret.flagSet {
		err := configStandardCallbacks(&ret, fieldVal)
		if err != nil {
			return paramName, ret, err
		}
	} else {
		// callbacks are not expected to be called on flagsets
		msg := fmt.Sprintf("%q is a flagset, it have no value", paramName)
		ret.validFn = func(v string) error { panic(msg) }
		ret.setValueFn = func(v *string) error { panic(msg) }
		ret.getDefaultFn = func() (string, error) { panic(msg) }
		ret.redactFn = func(s string) string { panic(msg) }
	}

	return paramName, ret, nil
}

func addSpecialFlags(appConfig config, parsed *Parsed, opts settings) error {
	// support special flags
	var violations ErrViolations
	if opts.autoUsageExitFn != nil {
		if conflictingParam, ok := appConfig[""].fields[specialflags.Help.Name]; ok {
			violations = append(violations, Violation{
				ParamName: specialflags.Help.Name,
				Path:      conflictingParam.path,
				Message:   "The help parameter cannot be used with the auto-usage is requested",
			})
		}

		appConfig[""].fields[specialflags.Help.Name] = flagSetField{
			typ:      "bool",
			optional: true,
			desc:     specialflags.Help.Description,
			boolean:  true,

			setValueFn: func(v *string) error {
				parsed.Usage(os.Stdout)
				opts.autoUsageExitFn()
				return nil
			},

			validFn:      func(v string) error { return nil },
			getDefaultFn: func() (string, error) { return "false", nil },
			redactFn:     func(s string) string { return s },
		}
	}

	if len(violations) > 0 {
		return nil
	}

	return nil
}

func processSpecialFlags(appConfig config, parsed *Parsed, opts settings) {
	if parsed.settings.autoUsageExitFn == nil || parsed.readValue("", specialflags.Help.Name) == nil {
		return
	}

	fmt.Fprintln(opts.autoUsageWriter, opts.autoUsageHeadline)
	parsed.Usage(opts.autoUsageWriter)
	parsed.settings.autoUsageExitFn()
	panic("Auto usage termination callback function did not terminated the application")
}

func panicOnNil(v *string) {
	if v == nil {
		panic("bug: tried to set non-dynamic parameter to nil")
	}
}

func describeType(val reflect.Value) string {
	t := val.Type()
	if isDynamicType(t) {
		return dynamicTypeName(val)
	}

	return t.Name()
}
