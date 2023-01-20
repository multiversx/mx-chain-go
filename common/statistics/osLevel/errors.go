package osLevel

import "errors"

var (
	errWrongNumberOfComponents = errors.New("wrong number of components")
	errUnparsableData          = errors.New("unparsable data, no field was set")
)
