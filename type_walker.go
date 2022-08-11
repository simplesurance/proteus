package proteus

import (
	"reflect"
	"strings"

	"github.com/simplesurance/proteus/types"
)

// flatWalk visits all fields of a struct, including embedded values,
// collect information about found fields and returns it.
func flatWalk(setName, setPath string, val reflect.Value) (map[string]fieldAndValue, error) {
	foundFields := map[string]fieldAndValue{}

	if val.Type().Kind() != reflect.Struct {
		return nil, types.ErrViolations([]types.Violation{
			{
				SetName: setName,
				Message: "only structs can be paramsets",
			},
		})
	}

	var violations types.ErrViolations
	// recursive function to walk on fields, including the ones on embedded
	// structs
	var walker func(reflect.Value, string)
	walker = func(val reflect.Value, path string) {
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			fieldValue := val.Field(i)
			path := path + "/" + field.Name

			if field.Type.Kind() == reflect.Struct && field.Anonymous {
				walker(fieldValue, path)
				continue
			}

			// duplicated fields are invalid
			normalizedName := strings.ToLower(field.Name)
			if _, ok := foundFields[normalizedName]; ok {
				violations = append(violations, types.Violation{
					SetName: setName,
					Path:    path,
					Message: "Duplicated field",
				})
				continue
			}

			foundFields[normalizedName] = fieldAndValue{
				field: field,
				value: fieldValue,
				Path:  path,
			}
		}
	}

	walker(val, setPath)

	if len(violations) > 0 {
		return nil, violations
	}

	return foundFields, nil
}

type fieldAndValue struct {
	field reflect.StructField
	value reflect.Value
	Path  string
}
