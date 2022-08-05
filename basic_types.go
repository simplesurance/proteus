package proteus

import (
	"fmt"
	"reflect"
	"strconv"
)

func configStandardCallbacks(fieldData *paramSetField, val reflect.Value) error {
	// the redact function is to allow redacting part of a value, like
	// redacting the "password" part of an URL. For basic types use
	// the identity function.
	fieldData.redactFn = func(s string) string { return s }

	// non-xtype values are either left alone with whatever value they
	// had initially, or written once, with a value provided by a
	// configuration provider.
	switch val.Type().Kind() {
	case reflect.String:
		fieldData.validFn = func(str string) error {
			return nil
		}

		fieldData.setValueFn = func(str *string) error {
			panicOnNil(str)
			val.SetString(*str)
			return nil
		}

		fieldData.getDefaultFn = func() (string, error) {
			return val.String(), nil
		}

		return nil
	case reflect.Bool:
		fieldData.boolean = true
		fieldData.validFn = func(str string) error {
			_, err := strconv.ParseBool(str)
			return err
		}

		fieldData.setValueFn = func(str *string) error {
			panicOnNil(str)
			v, err := strconv.ParseBool(*str)
			if err != nil {
				return err
			}

			val.SetBool(v)
			return nil
		}

		fieldData.getDefaultFn = func() (string, error) {
			return strconv.FormatBool(val.Bool()), nil
		}

		return nil
	case reflect.Int:
		configAsInt(fieldData, val, 0)
		return nil
	case reflect.Int8:
		configAsInt(fieldData, val, 8)
		return nil
	case reflect.Int16:
		configAsInt(fieldData, val, 16)
		return nil
	case reflect.Int32:
		configAsInt(fieldData, val, 32)
		return nil
	case reflect.Int64:
		configAsInt(fieldData, val, 64)
		return nil
	case reflect.Uint:
		configAsUint(fieldData, val, 0)
		return nil
	case reflect.Uint8:
		configAsUint(fieldData, val, 8)
		return nil
	case reflect.Uint16:
		configAsUint(fieldData, val, 16)
		return nil
	case reflect.Uint32:
		configAsUint(fieldData, val, 32)
		return nil
	case reflect.Uint64:
		configAsUint(fieldData, val, 64)
		return nil
	default:
		return fmt.Errorf("unsupported type %+v", val.Type())
	}

}

func configAsInt(fieldData *paramSetField, val reflect.Value, bitSize int) {
	fieldData.validFn = func(str string) error {
		_, err := strconv.ParseInt(str, 10, bitSize)
		return err
	}

	fieldData.setValueFn = func(str *string) error {
		panicOnNil(str)
		v, err := strconv.ParseInt(*str, 10, bitSize)
		if err != nil {
			return err
		}

		val.SetInt(v)
		return nil
	}

	fieldData.getDefaultFn = func() (string, error) {
		return strconv.FormatInt(val.Int(), 10), nil
	}
}

func configAsUint(fieldData *paramSetField, val reflect.Value, bitSize int) {
	fieldData.validFn = func(str string) error {
		_, err := strconv.ParseUint(str, 10, bitSize)
		return err
	}

	fieldData.setValueFn = func(str *string) error {
		panicOnNil(str)
		v, err := strconv.ParseUint(*str, 10, bitSize)
		if err != nil {
			return err
		}

		val.SetUint(v)
		return nil
	}

	fieldData.getDefaultFn = func() (string, error) {
		return strconv.FormatUint(val.Uint(), 10), nil
	}
}

func panicOnNil(v *string) {
	if v == nil {
		panic("bug: tried to set non-xtype parameter to nil")
	}
}
