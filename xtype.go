package proteus

import (
	"fmt"
	"reflect"

	"github.com/simplesurance/proteus/types"
)

const invalidValue = "<invalid>"

// isXType will return true is the type supports hot-reloading the
// parameter value. This is determine by:
//   - the value belonging to a type that implements the unmarshaller interface
//   - the value belonging to a type that implements a method Value() T, where
//     the type T any supported value.
func isXType(ty reflect.Type) (bool, error) {
	tyUnmarshaler := reflect.TypeOf((*types.XType)(nil)).Elem()
	if !ty.AssignableTo(tyUnmarshaler) {
		return false, nil
	}

	valueMethod, ok := ty.MethodByName("Value")
	if !ok {
		return false, fmt.Errorf("provided XType is incorrectly implemented: missing 'Value() T' method")
	}

	if valueMethod.Type.NumIn() != 1 {
		return false, fmt.Errorf("provided XType is incorrectly implemented: 'Value() method must have 0 input parameters")
	}

	if valueMethod.Type.NumOut() != 1 {
		return false, fmt.Errorf("provided XType is incorrectly implemented: 'Value() method must return 1 value")
	}

	return true, nil
}

// describeXType creates a short description of what kind of parameter
// this refers to. It will be inferred from the return types of Value(),
// and the type may override this by implementing the types.XTypeDescriber
// interface.
func describeXType(val reflect.Value) string {
	// give the opportunity for the value to describe itself by
	// implementing the  interface.
	tyDescriber := reflect.TypeOf((*types.TypeDescriber)(nil)).Elem()
	if val.Type().AssignableTo(tyDescriber) {
		return val.Interface().(types.TypeDescriber).DescribeType()
	}

	// use the return type of the Value() method
	ty := val.Type()

	valueMethod, ok := ty.MethodByName("Value")
	if !ok {
		return invalidValue
	}

	if valueMethod.Type.NumOut() != 1 {
		return invalidValue
	}

	return valueMethod.Type.Out(0).String()
}

func toXType(val reflect.Value) types.XType {
	return val.Interface().(types.XType)
}

func toRedactor(val reflect.Value) types.Redactor {
	if ret, ok := val.Interface().(types.Redactor); ok {
		return ret
	}

	return nil
}
