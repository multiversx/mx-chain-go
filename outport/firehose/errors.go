package firehose

import "errors"

var errNilWriter = errors.New("nil writer provided")

var errNilHeader = errors.New("received nil header")

var errInvalidHeaderType = errors.New("received invalid/unknown header type")
