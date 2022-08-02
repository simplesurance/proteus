package proteus

import (
	"reflect"

	"github.com/simplesurance/proteus/types"
)

const invalidValue = "<invalid>"

// isXType will return true is the type supports hot-reloading the
// parameter value. This is determine by:
// - the value belonging to a type that implements the unmarshaller interface
// - the value belonging to a type that implements a method Value() T, where
//   the type T any supported value.
func isXType(ty reflect.Type) bool {
	tyUnmarshaler := reflect.TypeOf((*types.DynamicType)(nil)).Elem()
	if !ty.AssignableTo(tyUnmarshaler) {
		return false
	}

	valueMethod, ok := ty.MethodByName("Value")
	if !ok {
		return false
	}

	if valueMethod.Type.NumIn() != 1 {
		return false
	}

	return true
}

func dynamicTypeName(val reflect.Value) string {
	// give the opportunity for the value to describe itself by
	// implementing the dynamicTypeDescriber interface.
	tyDescriber := reflect.TypeOf((*types.DynamicTypeDescriber)(nil)).Elem()
	if val.Type().AssignableTo(tyDescriber) {
		return val.Interface().(types.DynamicTypeDescriber).DescribeDynamicType()
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

func toDynamic(val reflect.Value) types.DynamicType {
	return val.Interface().(types.DynamicType)
}

func toRedactor(val reflect.Value) types.Redactor {
	if ret, ok := val.Interface().(types.Redactor); ok {
		return ret
	}

	return nil
}
