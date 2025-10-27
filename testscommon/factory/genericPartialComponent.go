package factory

import (
	"fmt"
	"reflect"
	"unsafe"
)

// ConstructPartialComponentForTest initializes the given component by setting its subcomponents
func ConstructPartialComponentForTest(component interface{}, subcomponents map[string]interface{}) error {

	// initialize nil pointer fields
	rv := reflect.ValueOf(component).Elem()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		ft := rv.Type().Field(i)

		// handle embedded *structs that are nil
		if ft.Anonymous && field.Kind() == reflect.Ptr && field.IsNil() && field.Type().Elem().Kind() == reflect.Struct {
			newVal := reflect.New(field.Type().Elem())
			if !field.CanSet() {
				field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			}
			field.Set(newVal)
		}
	}

	// set subcomponents
	setField := func(target any, name string, component any) error {
		rv := reflect.ValueOf(target).Elem()
		field := rv.FieldByName(name)
		if !field.IsValid() {
			component = nil
			return fmt.Errorf("invalid field: %s", name)
		}
		if !field.CanSet() {
			// bypass export check (ok in tests, same package)
			field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		}

		val := reflect.ValueOf(component)
		switch {
		case val.Type().AssignableTo(field.Type()):
			field.Set(val)
		case val.Type().ConvertibleTo(field.Type()):
			field.Set(val.Convert(field.Type()))
		case val.Kind() != reflect.Ptr && field.Kind() == reflect.Ptr && val.Type().AssignableTo(field.Type().Elem()):
			ptr := reflect.New(val.Type())
			ptr.Elem().Set(val)
			field.Set(ptr)
		case val.Kind() != reflect.Ptr && field.Kind() == reflect.Ptr && val.Type().ConvertibleTo(field.Type().Elem()):
			ptr := reflect.New(field.Type().Elem())
			ptr.Elem().Set(val.Convert(field.Type().Elem()))
			field.Set(ptr)
		default:
			return fmt.Errorf("cannot set field %s (got %s, expected %s)", name, val.Type(), field.Type())
		}
		return nil
	}

	for name, subComponent := range subcomponents {
		if err := setField(component, name, subComponent); err != nil {
			return err
		}
	}
	return nil
}
