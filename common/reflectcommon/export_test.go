package reflectcommon

import "reflect"

// FitsWithinSignedIntegerRange -
func FitsWithinSignedIntegerRange(value reflect.Value, targetType reflect.Type) bool {
	return fitsWithinSignedIntegerRange(value, targetType)
}

// FitsWithinUnsignedIntegerRange -
func FitsWithinUnsignedIntegerRange(value reflect.Value, targetType reflect.Type) bool {
	return fitsWithinUnsignedIntegerRange(value, targetType)
}

// FitsWithinFloatRange -
func FitsWithinFloatRange(value reflect.Value, targetType reflect.Type) bool {
	return fitsWithinFloatRange(value, targetType)
}
