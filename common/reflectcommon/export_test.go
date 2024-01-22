package reflectcommon

import "reflect"

func FitsWithinSignedIntegerRange(value reflect.Value, targetType reflect.Type) bool {
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

func FitsWithinUnsignedIntegerRange(value reflect.Value, targetType reflect.Type) bool {
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

func FitsWithinFloatRange(value reflect.Value, targetType reflect.Type) bool {
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
