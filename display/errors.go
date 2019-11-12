package display

import (
	"errors"
)

// ErrNilHeader signals that a nil header slice has been provided
var ErrNilHeader = errors.New("nil header")

// ErrNilDataLines signals that a nil data lines slice has been provided
var ErrNilDataLines = errors.New("nil lineData slice")

// ErrEmptySlices signals that empty slices has been provided
var ErrEmptySlices = errors.New("empty slices")

// ErrNilLineDataInSlice signals that a nil line data element was found in slice
var ErrNilLineDataInSlice = errors.New("nil line data element found in slice")

// ErrNilValuesOfLineDataInSlice signals that a line data element has nil values
var ErrNilValuesOfLineDataInSlice = errors.New("nil line data values slice found")

// ErrNilDisplayByteSliceHandler signals that a nil display byte slice handler has been provided
var ErrNilDisplayByteSliceHandler = errors.New("nil display byte slice handler")
