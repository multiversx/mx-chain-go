package broadcast

import "errors"

// ErrNilKeysHandler signals that a nil keys handler was provided
var ErrNilKeysHandler = errors.New("nil keys handler")
