package blake2b

import "errors"

// ErrInvalidHashSize signals that an invalid hash size has been provided
var ErrInvalidHashSize = errors.New("invalid hash size provided")
