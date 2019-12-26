package middleware

import "errors"

// ErrInvalidMaxNumRequests signals that a provided number of requests is invalid
var ErrInvalidMaxNumRequests = errors.New("max number of requests value is invalid")
