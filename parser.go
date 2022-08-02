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

// ParamNameRE is the regular expression used to valid parameter and parameter
// set names.
var ParamNameRE = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// MustParse reads the parameters into the provided structure reference. The
// provided parameters struct can be annotated with some parameter tags to
// configure how the configuration is read.
//
// There are support for sub-parameters and for getting updates about
// changes in value without the need to restart the application. See godoc
// examples for usage.
//
// A Parsed object is guaranteed to be always returned, even in case of error,
// allowing to get usage information.
func MustParse(parameters any, options ...Option) (*Parsed, error) {
	opts := settings{
		loggerFn: func(msg string, depth int) {}, // nop logger
	}
	opts.apply(options...)

	appConfig, err := mustInferConfigFromValue(parameters, opts)
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

	ret := config{"": paramSet{fields: map[string]paramSetField{}}}

	// each member of the configuration struct can be either:
	// - parameter: meaning that values must be loaded into it
	// - set of parameters: meaning that is a structure that contains more
	//   parameter.
	// - ignored: identified with: param:"-"
	members, err := flatWalk("", "", val)
	if err != nil {
		return nil, fmt.Errorf("walking root fields of the configuration struct: %w", err)
	}

	var violations types.ErrViolations
	for _, member := range members {
		name, tag, err := parseParam(member.field, member.value)
		if err != nil {
			var paramViolations types.ErrViolations
			if errors.As(err, &paramViolations) {
				violations = append(violations, paramViolations...)
				continue
			}

			violations = append(violations, types.Violation{
				Path:    member.Path,
				Message: fmt.Sprintf("error reading struct tag: %v", err),
			})
			continue
		}

		if name == "-" {
			continue
		}

		if !ParamNameRE.MatchString(name) {
			violations = append(violations, types.Violation{
				Path: member.Path,
				Message: fmt.Sprintf("Name %q is invalid for parameter or set (valid: %s)",
					name, ParamNameRE)})
		}

		if tag.paramSet {
			// is a set or parameters
			d, err := parseParamSet(name, member.Path, member.value)
			if err != nil {
				var setViolations types.ErrViolations
				if errors.As(err, &setViolations) {
					violations = append(violations, setViolations...)
					continue
				}

				violations = append(violations, types.Violation{
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

func parseParamSet(setName, setPath string, val reflect.Value) (paramSet, error) {
	members, err := flatWalk(setName, setPath, val)
	if err != nil {
		return paramSet{}, err
	}

	ret := paramSet{
		fields: make(map[string]paramSetField, len(members)),
	}

	violations := types.ErrViolations{}
	for _, member := range members {
		paramName, tag, err := parseParam(member.field, member.value)
		if err != nil {
			violations = append(violations, types.Violation{
				Path:    member.Path,
				Message: err.Error(),
			})
			continue
		}

		if paramName == "-" || tag.paramSet {
			continue
		}

		if !ParamNameRE.MatchString(paramName) {
			violations = append(violations, types.Violation{
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
	_ paramSetField,
	_ error,
) {
	tagParam := structField.Tag.Get("param")
	tagParamParts := strings.Split(tagParam, ",")
	paramName = tagParamParts[0]

	ret := paramSetField{
		typ:  describeType(fieldVal),
		desc: structField.Tag.Get("param_desc"),
	}

	for _, tagOption := range tagParamParts[1:] {
		switch tagOption {
		case "optional":
			ret.optional = true
		case "secret":
			ret.secret = true
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

	// try to configure it as a "basic type"
	err := configStandardCallbacks(&ret, fieldVal)
	if err == nil {
		return paramName, ret, nil
	}

	// try to configure it as an "xtype"
	if isXType(structField.Type) {
		if fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
		}

		ret.boolean = describeType(fieldVal) == "bool"

		ret.validFn = toDynamic(fieldVal).ValueValid
		ret.setValueFn = toDynamic(fieldVal).UnmarshalParam
		ret.getDefaultFn = toDynamic(fieldVal).GetDefaultValue

		// some types know how to redact themselves (for example,
		// xtype.URL know how to redact the password)
		if redactor := toRedactor(fieldVal); redactor != nil {
			ret.redactFn = redactor.RedactValue
		} else {
			ret.redactFn = func(s string) string { return s }
		}

		return paramName, ret, nil
	}

	// if is a struct, assume it to be a parameter set
	if fieldVal.Kind() == reflect.Struct {
		ret.paramSet = true

		// parameter sets have no value, and the callback functions should
		// not be called; install handlers to help debug in case of a mistake.
		panicMessage := fmt.Sprintf("%q is a paramset, it have no value", paramName)
		ret.validFn = func(v string) error { panic(panicMessage) }
		ret.setValueFn = func(v *string) error { panic(panicMessage) }
		ret.getDefaultFn = func() (string, error) { panic(panicMessage) }
		ret.redactFn = func(s string) string { panic(panicMessage) }

		return paramName, ret, nil
	}

	return paramName, ret, fmt.Errorf("struct member %q is unsupported", paramName)
}

func addSpecialFlags(appConfig config, parsed *Parsed, opts settings) error {
	// support special flags
	var violations types.ErrViolations
	if opts.autoUsageExitFn != nil {
		if conflictingParam, ok := appConfig[""].fields[specialflags.Help.Name]; ok {
			violations = append(violations, types.Violation{
				ParamName: specialflags.Help.Name,
				Path:      conflictingParam.path,
				Message:   "The help parameter cannot be used with the auto-usage is requested",
			})
		}

		appConfig[""].fields[specialflags.Help.Name] = paramSetField{
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
	if isXType(t) {
		return dynamicTypeName(val)
	}

	return t.Name()
}
