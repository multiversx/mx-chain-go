package check

import (
	"fmt"
	"math"
	"reflect"
)

// ForZeroUintFields checks if fields are uint64 and whether are greater than 0
func ForZeroUintFields(arg interface{}) error {
	v := reflect.ValueOf(arg)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() != reflect.Uint64 &&
			field.Kind() != reflect.Uint32 &&
			field.Kind() != reflect.Uint {
			continue
		}
		if field.Uint() == 0 {
			name := v.Type().Field(i).Name
			return fmt.Errorf("gas cost for operation %s has been set to 0 or is not set", name)
		}
	}

	return nil
}

// IsZeroFloat64 returns true if the provided absolute value is less than provided absolute epsilon
func IsZeroFloat64(value float64, epsilon float64) bool {
	return math.Abs(value) <= math.Abs(epsilon)
}
