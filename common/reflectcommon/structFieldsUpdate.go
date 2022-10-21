package reflectcommon

import (
	"fmt"
	"reflect"
	"strings"
)

func getReflectValue(original reflect.Value, fieldName string) (value reflect.Value, err error) {
	defer func() {
		// should never panic because of the validity check, but can remain for extra safety
		r := recover()
		if r != nil {
			err = fmt.Errorf("%v", r)
			return
		}
	}()

	value = original.FieldByName(fieldName)
	if !value.IsValid() {
		value = reflect.Value{}
		err = fmt.Errorf("%w: %s", errInvalidNestedStructure, fieldName)
		return
	}

	return
}

// AdaptStructureValueBasedOnPath will try to update the value specified at the given path in any structure
// the structure must be of type pointer, otherwise an error will be returned. All the fields or inner structures MUST be exported
// the path must be in the form of InnerStruct.InnerStruct2.Field
// newValue must have the same type as the old value, otherwise an error will be returned. Currently, this function does not support slices or maps
func AdaptStructureValueBasedOnPath(structure interface{}, path string, newValue interface{}) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if structure == nil {
		return errNilStructure
	}

	splitStr := strings.Split(path, ".")

	value := reflect.ValueOf(structure)
	if value.Kind() != reflect.Ptr {
		return errCannotUpdateValueStructure
	}

	value = value.Elem().FieldByName(splitStr[0])

	for i := 1; i < len(splitStr); i++ {

		value, err = getReflectValue(value, splitStr[i])
		if err != nil {
			return err
		}
	}

	if !value.CanSet() {
		err = fmt.Errorf("%w. field name=%s", errCannotSetValue, splitStr[len(splitStr)-1])
		return
	}

	value.Set(reflect.ValueOf(newValue))
	return nil
}
