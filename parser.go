package proteus

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/simplesurance/proteus/internal/consts"
	"github.com/simplesurance/proteus/sources"
	"github.com/simplesurance/proteus/sources/cfgenv"
	"github.com/simplesurance/proteus/sources/cfgflags"
	"github.com/simplesurance/proteus/types"
)

// MustParse receives on "config" a pointer to a struct that defines the
// expected application parameters and loads the parameters values into it.
// An example of a configuration is struct is as follows:
//
//	params := struct{
//		Name      string                                       // simple parameter
//		IsEnabled bool   `param:"is_enabled"`                  // rename parameter
//		Password  string `param:"pwd,secret"`                  // rename and mark parameter as secret
//		Port      uint16 `param:",optional"`                   // keep the name, mark as optional
//		LogLevel  string `param_desc:"Cut-off level for logs"` // describes the parameter
//		X         string `param:"-"`                           // ignore this field
//	}{
//		Port: 8080, // default value for optional parameter
//	}
//
// The tag "param" has the format "name[,option]*", where name is either empty,
// "-" or a lowercase arbitrary string containing a-z, 0-9, _ or -, starting with a-z and
// terminating not with - or _.
// The value "-" for the name result in the field being ignored. The empty
// string value indicates to infer the parameter name from the struct name. The
// inferred parameter name is the struct name in lowercase.
// Option can be either "secret" or "optional". An option can be provided
// without providing the name of the parameter by using an empty value for the
// name, resulting in the "param" tag starting with ",".
//
// The tag "param_desc" is an arbitrary string describing what the parameter
// is for. This will be shown to the user when usage information is requested.
//
// The provided struct can have any level of embedded structs. Embedded
// structs are handled as if they were "flat":
//
//	type httpParams struct {
//		Server string
//		Port   uint16
//	}
//
//	parmas := struct{
//		httpParams
//		LogLevel string
//	}{}
//
// Is the same as:
//
//	params := struct {
//		Server   string
//		Port     uint16
//		LogLevel string
//	}{}
//
// Configuration structs can also have "xtypes". Xtypes provide support for
// getting updates when parameter values change and other types-specific
// optons.
//
//	params := struct{
//		LogLevel *xtypes.OneOf
//	}{
//		OneOf: &xtypes.OneOf{
//			Choices: []string{"debug", "info", "error"},
//			Default: "info",
//			UpdateFn: func(newVal string) {
//				fmt.Printf("new log level: %s\n", newVal)
//			}
//		}
//	}
//
// The "options" parameter provides further customization. The option
// WithProviders() must be specified to define from what sources the parameters
// must be read.
//
// The configuration struct can have named sub-structs (in opposition to
// named, or embedded sub-structs, already mentioned above). The sub-structs
// can be up to 1 level deep, and can be used to represent "parameter sets".
// Two parameters can have the same name, as long as they belong to different
// parameter sets. Example:
//
//	params := struct{
//		Database struct {
//			Host     string
//			Username string
//			Password string `param:,secret`
//		}
//		Tracing struct {
//			Host     string
//			Username string
//			Password string `param:,secret`
//		}
//	}{}
//
// Complete usage example:
//
//	func main() {
//		params := struct {
//			X int
//		}{}
//
//		parsed, err := proteus.MustParse(&params,
//			proteus.WithAutoUsage(os.Stdout, "My Application", func() { os.Exit(0) }),
//			proteus.WithProviders(
//				cfgflags.New(),
//				cfgenv.New("CFG"),
//			))
//		if err != nil {
//			parsed.ErrUsage(os.Stderr, err)
//			os.Exit(1)
//		}
//
//		// "parsed" now have the parameter values
//	}
//
// See godoc for more examples.
//
// A Parsed object is guaranteed to be always returned, even in case of error,
// allowing the creation of useful error messages.
func MustParse(config any, options ...Option) (*Parsed, error) {
	opts := settings{
		providers: []sources.Provider{
			cfgflags.New(),
			cfgenv.New("CFG"),
		},
		loggerFn:        func(msg string, depth int) {}, // nop logger
		autoUsageExitFn: func() { os.Exit(0) },
		autoUsageWriter: os.Stdout,
	}
	opts.apply(options...)

	appConfig, err := inferConfigFromValue(config, opts)
	if err != nil {
		panic(fmt.Errorf("INVALID CONFIGURATION STRUCT: %v", err))
	}

	if len(opts.providers) == 0 {
		panic(fmt.Errorf("NO CONFIGURATION PROVIDER WAS PROVIDED"))
	}

	ret := Parsed{
		settings:      opts,
		inferedConfig: appConfig,
	}

	ret.protected.values = make([]types.ParamValues, len(opts.providers))

	if err := addSpecialFlags(appConfig, &ret, opts); err != nil {
		return &ret, err
	}

	// start watching each configuration item on each provider
	updaters := make([]*updater, len(opts.providers))
	for ix, provider := range opts.providers {
		updater := &updater{
			parsed:         &ret,
			providerIndex:  ix,
			providerName:   fmt.Sprintf("%T", provider),
			updatesEnabled: make(chan struct{})}

		updaters[ix] = updater

		initial, err := provider.Watch(
			appConfig.paramInfo(provider.IsCommandLineFlag()),
			updater)
		if err != nil {
			return &ret, err
		}

		// use the updater to store the initial values; do NOT update the
		// "config" struct yet
		updater.update(initial, false)
	}

	if err := ret.valid(); err != nil {
		return &ret, err
	}

	// send values back to the user by updating the fields on the
	// "config" parameter
	ret.refresh(true)

	// allow all sources to provide updates
	for _, updater := range updaters {
		close(updater.updatesEnabled)
	}

	return &ret, nil
}

func inferConfigFromValue(value any, opts settings) (config, error) {
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

		if !consts.ParamNameRE.MatchString(name) {
			violations = append(violations, types.Violation{
				Path: member.Path,
				Message: fmt.Sprintf("Name %q is invalid for parameter or set (valid: %s)",
					name, consts.ParamNameRE)})
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

		if !consts.ParamNameRE.MatchString(paramName) {
			violations = append(violations, types.Violation{
				Path:    member.Path,
				SetName: setName,
				Message: fmt.Sprintf("Name %q is invalid for parameter or set (valid: %s)",
					paramName, consts.ParamNameRE)})
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
				"option '%s' is invalid for tag 'param' in '%s'; valid options are optional|secret",
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
	ok, err := isXType(structField.Type)
	if err != nil {
		return paramName, ret, err
	}

	if ok {
		ret.isXtype = true

		if fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
		}

		ret.boolean = describeType(fieldVal) == "bool"

		ret.validFn = toXType(fieldVal).ValueValid
		ret.setValueFn = toXType(fieldVal).UnmarshalParam
		ret.getDefaultFn = toXType(fieldVal).GetDefaultValue

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

// addSpecialFlags register flags like "--help" that the caller might have
// requested, that can only be provided by command-line flags, and that have
// to be handled in a special way by proteus.
func addSpecialFlags(appConfig config, parsed *Parsed, opts settings) error {
	var violations types.ErrViolations

	// --help
	if opts.autoUsageExitFn != nil {
		helpFlagName := "help"
		helpFlagDescription := "Prints information about how to use this application"

		if conflictingParam, exists := appConfig.getParam("", helpFlagName); exists {
			violations = append(violations, types.Violation{
				ParamName: helpFlagName,
				Path:      conflictingParam.path,
				Message:   "The help parameter cannot be used when the auto-usage is requested",
			})
		} else {
			appConfig[""].fields[helpFlagName] = paramSetField{
				typ:       "bool",
				optional:  true,
				desc:      helpFlagDescription,
				boolean:   true,
				isSpecial: true,

				// when the --help flag is provided, the parsed object will
				// try to determine if the value is valid. Generate the
				// help usage instead of terminate the application.
				validFn: func(v string) error {
					parsed.Usage(opts.autoUsageWriter)
					parsed.settings.autoUsageExitFn()

					fmt.Fprintln(opts.autoUsageWriter, "WARNING: the provided termination function did not terminated the application")
					os.Exit(0)
					return nil
				},
				setValueFn:   func(_ *string) error { return nil },
				getDefaultFn: func() (string, error) { return "false", nil },
				redactFn:     func(s string) string { return s },
			}
		}
	}

	// TODO: support --dry-mode

	if len(violations) > 0 {
		return nil
	}

	return nil
}

func describeType(val reflect.Value) string {
	t := val.Type()
	if ok, _ := isXType(t); ok {
		return describeXType(val)
	}

	return t.Name()
}
