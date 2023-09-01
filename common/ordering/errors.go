package ordering

import "errors"

// ErrItemNotFound is returned when an item is not found in the ordered collection
var ErrItemNotFound = errors.New("item not found")

// ErrIndexOutOfBounds is returned when an index is out of bounds
var ErrIndexOutOfBounds = errors.New("index out of bounds")
