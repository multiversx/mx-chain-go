package reflectcommon

import "reflect"

func FitsWithinSignedIntegerRange(value reflect.Value, targetType reflect.Type) bool {
	return fitsWithinSignedIntegerRange(value, targetType)
}

func FitsWithinUnsignedIntegerRange(value reflect.Value, targetType reflect.Type) bool {
	return fitsWithinUnsignedIntegerRange(value, targetType)
}

func FitsWithinFloatRange(value reflect.Value, targetType reflect.Type) bool {
	return fitsWithinFloatRange(value, targetType)
}
