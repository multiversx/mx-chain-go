package reflectcommon

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/check"
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
func AdaptStructureValueBasedOnPath(structure interface{}, path string, newValue string) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if check.IfNilReflect(structure) {
		return errNilStructure
	}
	if len(path) == 0 {
		return errEmptyPath
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

	return trySetTheNewValue(&value, newValue)
}

func trySetTheNewValue(value *reflect.Value, newValue string) error {
	valueKind := value.Kind()

	errFunc := func() error {
		return fmt.Errorf("cannot cast field <%s> to kind <%s>", newValue, valueKind)
	}

	switch valueKind {
	case reflect.Invalid:
		return errFunc()
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(newValue)
		if err != nil {
			return fmt.Errorf("%w: %s", errFunc(), err.Error())
		}

		value.Set(reflect.ValueOf(boolVal))
	case reflect.Int:
		intVal, err := strconv.ParseInt(newValue, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %s", errFunc(), err.Error())
		}

		value.Set(reflect.ValueOf(int(intVal)))
	case reflect.Int32:
		int32Val, err := strconv.ParseInt(newValue, 10, 32)
		if err != nil {
			return fmt.Errorf("%w: %s", errFunc(), err.Error())
		}

		value.Set(reflect.ValueOf(int32(int32Val)))
	case reflect.Int64:
		int64Val, err := strconv.ParseInt(newValue, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %s", errFunc(), err.Error())
		}

		value.Set(reflect.ValueOf(int64Val))
	case reflect.Uint32:
		uint32Val, err := strconv.ParseUint(newValue, 10, 32)
		if err != nil {
			return fmt.Errorf("%w: %s", errFunc(), err.Error())
		}

		value.Set(reflect.ValueOf(uint32(uint32Val)))
	case reflect.Uint64:
		uint64Val, err := strconv.ParseUint(newValue, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %s", errFunc(), err.Error())
		}

		value.Set(reflect.ValueOf(uint64Val))
	case reflect.Float32:
		float32Val, err := strconv.ParseFloat(newValue, 32)
		if err != nil {
			return fmt.Errorf("%w: %s", errFunc(), err.Error())
		}

		value.Set(reflect.ValueOf(float32(float32Val)))
	case reflect.Float64:
		float64Val, err := strconv.ParseFloat(newValue, 32)
		if err != nil {
			return fmt.Errorf("%w: %s", errFunc(), err.Error())
		}

		value.Set(reflect.ValueOf(float64Val))
	case reflect.String:
		value.Set(reflect.ValueOf(newValue))
	default:
		return fmt.Errorf("unsupported type <%s> when trying to set the value <%s>", valueKind, newValue)
	}
	return nil
}
