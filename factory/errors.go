package factory

import "errors"

// ErrNilDelegatedListFactory signal that a nil delegated list handler factory has been provided
var ErrNilDelegatedListFactory = errors.New("nil delegated list factory has been provided")

// ErrNilDirectStakedListFactory signal that a nil direct staked list handler factory has been provided
var ErrNilDirectStakedListFactory = errors.New("nil direct staked list factory has been provided")
