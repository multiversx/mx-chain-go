package reflectcommon

import "errors"

var (
	errNilStructure               = errors.New("nil structure to update")
	errEmptyPath                  = errors.New("empty path to update")
	errCannotUpdateValueStructure = errors.New("cannot update structures that are not passed by pointer")
	errCannotSetValue             = errors.New("cannot set value for field. it or it's structure might be unexported")
	errInvalidNestedStructure     = errors.New("invalid structure name")
)
