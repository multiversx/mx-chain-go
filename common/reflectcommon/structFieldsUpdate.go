package reflectcommon

import (
	"fmt"
	"math"
	"reflect"
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
func AdaptStructureValueBasedOnPath(structure interface{}, path string, newValue interface{}) (err error) {
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

func trySetTheNewValue(value *reflect.Value, newValue interface{}) error {
	valueKind := value.Kind()

	errFunc := func() error {
		return fmt.Errorf("cannot cast value '%s' of type <%s> to kind <%s>", newValue, reflect.TypeOf(newValue), valueKind)
	}

	switch valueKind {
	case reflect.Invalid:
		return errFunc()
	case reflect.Bool:
		boolVal, err := newValue.(bool)
		if !err {
			return errFunc()
		}

		value.SetBool(boolVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		reflectVal := reflect.ValueOf(newValue)
		if !reflectVal.Type().ConvertibleTo(value.Type()) {
			return errFunc()
		}
		//Check if the newValue fits inside the signed int value
		if !fitsWithinSignedIntegerRange(reflectVal, value.Type()) {
			return fmt.Errorf("value '%s' does not fit within the range of <%s>", reflectVal, value.Type())
		}

		convertedValue := reflectVal.Convert(value.Type())
		value.Set(convertedValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		reflectVal := reflect.ValueOf(newValue)
		if !reflectVal.Type().ConvertibleTo(value.Type()) {
			return errFunc()
		}
		//Check if the newValue fits inside the unsigned int value
		if !fitsWithinUnsignedIntegerRange(reflectVal, value.Type()) {
			return fmt.Errorf("value '%s' does not fit within the range of %s", reflectVal, value.Type())
		}

		convertedValue := reflectVal.Convert(value.Type())
		value.Set(convertedValue)
	case reflect.Float32, reflect.Float64:
		reflectVal := reflect.ValueOf(newValue)
		if !reflectVal.Type().ConvertibleTo(value.Type()) {
			return errFunc()
		}
		//Check if the newValue fits inside the unsigned int value
		if !fitsWithinFloatRange(reflectVal, value.Type()) {
			return fmt.Errorf("value '%s' does not fit within the range of %s", reflectVal, value.Type())
		}

		convertedValue := reflectVal.Convert(value.Type())
		value.Set(convertedValue)
	case reflect.String:
		strVal, err := newValue.(string)
		if !err {
			return errFunc()
		}

		value.SetString(strVal)
	case reflect.Slice:
		return trySetSliceValue(value, newValue)
	case reflect.Struct:
		structVal := reflect.ValueOf(newValue)

		return trySetStructValue(value, structVal)
	default:
		return fmt.Errorf("unsupported type <%s> when trying to set the value <%s>", valueKind, newValue)
	}
	return nil
}

func trySetSliceValue(value *reflect.Value, newValue interface{}) error {
	sliceVal := reflect.ValueOf(newValue)
	newSlice := reflect.MakeSlice(value.Type(), sliceVal.Len(), sliceVal.Len())

	for i := 0; i < sliceVal.Len(); i++ {
		item := sliceVal.Index(i)
		newItem := reflect.New(value.Type().Elem()).Elem()

		err := trySetStructValue(&newItem, item)
		if err != nil {
			return err
		}

		newSlice.Index(i).Set(newItem)
	}

	value.Set(newSlice)

	return nil
}

func trySetStructValue(value *reflect.Value, newValue reflect.Value) error {
	switch newValue.Kind() {
	case reflect.Invalid:
		return fmt.Errorf("invalid newValue kind <%s>", newValue.Kind())
	case reflect.Map: // overwrite with value read from toml file
		return updateStructFromMap(value, newValue)
	case reflect.Struct: // overwrite with go struct
		return updateStructFromStruct(value, newValue)
	default:
		return fmt.Errorf("unsupported type <%s> when trying to set the value of type <%s>", newValue.Kind(), value.Kind())
	}
}

func updateStructFromMap(value *reflect.Value, newValue reflect.Value) error {
	for _, key := range newValue.MapKeys() {
		fieldName := key.String()
		field := value.FieldByName(fieldName)

		if field.IsValid() && field.CanSet() {
			err := trySetTheNewValue(&field, newValue.MapIndex(key).Interface())
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("field <%s> not found or cannot be set", fieldName)
		}
	}

	return nil
}

func updateStructFromStruct(value *reflect.Value, newValue reflect.Value) error {
	for i := 0; i < newValue.NumField(); i++ {
		fieldName := newValue.Type().Field(i).Name
		field := value.FieldByName(fieldName)

		if field.IsValid() && field.CanSet() {
			err := trySetTheNewValue(&field, newValue.Field(i).Interface())
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("field <%s> not found or cannot be set", fieldName)
		}
	}

	return nil
}

func fitsWithinSignedIntegerRange(value reflect.Value, targetType reflect.Type) bool {
	min := getMinInt(targetType)
	max := getMaxInt(targetType)

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() >= min && value.Int() <= max
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() <= uint64(max)
	default:
		return false
	}
}

func fitsWithinUnsignedIntegerRange(value reflect.Value, targetType reflect.Type) bool {
	max := getMaxUint(targetType)

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() >= 0 && uint64(value.Int()) <= max
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() <= math.MaxUint
	default:
		return false
	}
}

func fitsWithinFloatRange(value reflect.Value, targetType reflect.Type) bool {
	min := getMinFloat(targetType)
	max := getMaxFloat(targetType)

	return value.Float() >= min && value.Float() <= max
}

func getMinInt(targetType reflect.Type) int64 {
	switch targetType.Kind() {
	case reflect.Int, reflect.Int64:
		return math.MinInt64
	case reflect.Int8:
		return int64(math.MinInt8)
	case reflect.Int16:
		return int64(math.MinInt16)
	case reflect.Int32:
		return int64(math.MinInt32)
	default:
		return 0
	}
}

func getMaxInt(targetType reflect.Type) int64 {
	switch targetType.Kind() {
	case reflect.Int, reflect.Int64:
		return math.MaxInt64
	case reflect.Int8:
		return int64(math.MaxInt8)
	case reflect.Int16:
		return int64(math.MaxInt16)
	case reflect.Int32:
		return int64(math.MaxInt32)
	default:
		return 0
	}
}

func getMaxUint(targetType reflect.Type) uint64 {
	switch targetType.Kind() {
	case reflect.Uint, reflect.Uint64:
		return math.MaxUint64
	case reflect.Uint8:
		return uint64(math.MaxUint8)
	case reflect.Uint16:
		return uint64(math.MaxUint16)
	case reflect.Uint32:
		return uint64(math.MaxUint32)
	default:
		return 0
	}
}

func getMinFloat(targetType reflect.Type) float64 {
	switch targetType.Kind() {
	case reflect.Float32:
		return math.SmallestNonzeroFloat32
	case reflect.Float64:
		return math.SmallestNonzeroFloat64
	default:
		return 0
	}
}

func getMaxFloat(targetType reflect.Type) float64 {
	switch targetType.Kind() {
	case reflect.Float32:
		return math.MaxFloat32
	case reflect.Float64:
		return math.MaxFloat64
	default:
		return 0
	}
}
