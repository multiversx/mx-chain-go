package wrapper

import "errors"

// ErrNilRouter signals that a nil router has been provided
var ErrNilRouter = errors.New("nil router")
