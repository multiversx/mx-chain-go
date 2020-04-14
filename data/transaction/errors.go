package transaction

import "errors"

// ErrNilEncoder signals that a nil encoder has been provided
var ErrNilEncoder = errors.New("nil encoder")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")
