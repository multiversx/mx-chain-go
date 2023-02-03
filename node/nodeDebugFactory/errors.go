package nodeDebugFactory

import "errors"

// ErrNilNodeWrapper signals that a nil node wrapper has been provided
var ErrNilNodeWrapper = errors.New("nil node wrapper")

// ErrNilInterceptorContainer signals that a nil interceptor container has been provided
var ErrNilInterceptorContainer = errors.New("nil interceptor container")

// ErrNilResolverContainer signals that a nil resolver container has been provided
var ErrNilResolverContainer = errors.New("nil resolver container")

// ErrNilRequestersContainer signals that a nil requesters container has been provided
var ErrNilRequestersContainer = errors.New("nil requesters container")
