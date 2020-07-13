package middleware

import "errors"

// ErrInvalidMaxNumConcurrentRequests signals that a provided number of concurrent requests is invalid
var ErrInvalidMaxNumConcurrentRequests = errors.New("max number of concurrent requests value is invalid")

// ErrInvalidMaxNumRequests signals that a provided number of requests is invalid
var ErrInvalidMaxNumRequests = errors.New("max number of requests value is invalid")

// ErrTooManyRequests signals that too many requests were simultaneously received
var ErrTooManyRequests = errors.New("too many requests")

// ErrNilFacade signals that nil facade provided is nil
var ErrNilFacade = errors.New("nil facade handler")
