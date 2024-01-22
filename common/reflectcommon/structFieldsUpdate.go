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
		return fmt.Errorf("unable to cast value '%v' of type <%s> to type <%s>", newValue, reflect.TypeOf(newValue), valueKind)
	}

	switch valueKind {
	case reflect.Invalid:
		return errFunc()
	case reflect.Bool:
		boolVal, ok := newValue.(bool)
		if !ok {
			return errFunc()
		}

		value.SetBool(boolVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, ok := convertToSignedInteger(value, newValue)
		if !ok {
			return errFunc()
		}

		value.Set(*intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, ok := convertToUnsignedInteger(value, newValue)
		if !ok {
			return errFunc()
		}

		value.Set(*uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, ok := convertToFloat(value, newValue)
		if !ok {
			return errFunc()
		}

		value.Set(*floatVal)
	case reflect.String:
		strVal, ok := newValue.(string)
		if !ok {
			return errFunc()
		}

		value.SetString(strVal)
	case reflect.Slice:
		return trySetSliceValue(value, newValue)
	case reflect.Struct:
		structVal := reflect.ValueOf(newValue)

		return trySetStructValue(value, structVal)
	default:
		return fmt.Errorf("unsupported type <%s> when trying to set the value '%v' of type <%s>", valueKind, newValue, reflect.TypeOf(newValue))
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
		return fmt.Errorf("invalid new value kind")
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

func convertToSignedInteger(value *reflect.Value, newValue interface{}) (*reflect.Value, bool) {
	var reflectVal = reflect.ValueOf(newValue)

	if !isIntegerType(reflectVal.Type()) {
		return nil, false
	}

	if !fitsWithinSignedIntegerRange(reflectVal, value.Type()) {
		return nil, false
	}

	convertedVal := reflectVal.Convert(value.Type())
	return &convertedVal, true
}

func convertToUnsignedInteger(value *reflect.Value, newValue interface{}) (*reflect.Value, bool) {
	var reflectVal = reflect.ValueOf(newValue)

	if !isIntegerType(reflectVal.Type()) {
		return nil, false
	}

	if !fitsWithinUnsignedIntegerRange(reflectVal, value.Type()) {
		return nil, false
	}

	convertedVal := reflectVal.Convert(value.Type())
	return &convertedVal, true
}

func isIntegerType(value reflect.Type) bool {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func fitsWithinSignedIntegerRange(value reflect.Value, targetType reflect.Type) bool {
	min, err := getMinInt(targetType)
	if err != nil {
		return false
	}
	max, err := getMaxInt(targetType)
	if err != nil {
		return false
	}

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() >= min && value.Int() <= max
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() <= uint64(max)
	}

	return false
}

func fitsWithinUnsignedIntegerRange(value reflect.Value, targetType reflect.Type) bool {
	max, err := getMaxUint(targetType)
	if err != nil {
		return false
	}

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() >= 0 && uint64(value.Int()) <= max
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() <= max
	}

	return false
}

func convertToFloat(value *reflect.Value, newValue interface{}) (*reflect.Value, bool) {
	var reflectVal = reflect.ValueOf(newValue)

	if !isFloatType(reflectVal.Type()) {
		return nil, false
	}

	if !fitsWithinFloatRange(reflectVal, value.Type()) {
		return nil, false
	}

	convertedVal := reflectVal.Convert(value.Type())
	return &convertedVal, true
}

func isFloatType(value reflect.Type) bool {
	switch value.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func fitsWithinFloatRange(value reflect.Value, targetType reflect.Type) bool {
	min, err := getMinFloat(targetType)
	if err != nil {
		return false
	}
	max, err := getMaxFloat(targetType)
	if err != nil {
		return false
	}

	return value.Float() >= min && value.Float() <= max
}

func getMinInt(targetType reflect.Type) (int64, error) {
	switch targetType.Kind() {
	case reflect.Int:
		return math.MinInt, nil
	case reflect.Int64:
		return math.MinInt64, nil
	case reflect.Int8:
		return math.MinInt8, nil
	case reflect.Int16:
		return math.MinInt16, nil
	case reflect.Int32:
		return math.MinInt32, nil
	default:
		return 0, fmt.Errorf("target type is not integer")
	}
}

func getMaxInt(targetType reflect.Type) (int64, error) {
	switch targetType.Kind() {
	case reflect.Int:
		return math.MaxInt, nil
	case reflect.Int64:
		return math.MaxInt64, nil
	case reflect.Int8:
		return math.MaxInt8, nil
	case reflect.Int16:
		return math.MaxInt16, nil
	case reflect.Int32:
		return math.MaxInt32, nil
	default:
		return 0, fmt.Errorf("target type is not integer")
	}
}

func getMaxUint(targetType reflect.Type) (uint64, error) {
	switch targetType.Kind() {
	case reflect.Uint:
		return math.MaxUint, nil
	case reflect.Uint64:
		return math.MaxUint64, nil
	case reflect.Uint8:
		return math.MaxUint8, nil
	case reflect.Uint16:
		return math.MaxUint16, nil
	case reflect.Uint32:
		return math.MaxUint32, nil
	default:
		return 0, fmt.Errorf("taget type is not unsigned integer")
	}
}

func getMinFloat(targetType reflect.Type) (float64, error) {
	switch targetType.Kind() {
	case reflect.Float32:
		return -math.MaxFloat32, nil
	case reflect.Float64:
		return -math.MaxFloat64, nil
	default:
		return 0, fmt.Errorf("target type is not float")
	}
}

func getMaxFloat(targetType reflect.Type) (float64, error) {
	switch targetType.Kind() {
	case reflect.Float32:
		return math.MaxFloat32, nil
	case reflect.Float64:
		return math.MaxFloat64, nil
	default:
		return 0, fmt.Errorf("target type is not float")
	}
}
