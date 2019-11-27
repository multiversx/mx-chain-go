package provider

import "errors"

// ErrTypeAssertionFailed signals that a value for a metric is not an accepted one
var ErrTypeAssertionFailed = errors.New("type assertion is not possible")

// ErrInvalidAddressLength signals that an invalid length has been provided for the node's address
var ErrInvalidAddressLength = errors.New("invalid length for the node address")

// ErrInvalidFetchInterval signals that the duration in seconds between fetches is invalid
var ErrInvalidFetchInterval = errors.New("invalid fetch interval has been provided")

// ErrNilTermuiPresenter signals that that a nil termui presenter has been provided
var ErrNilTermuiPresenter = errors.New("nil termui presenter")
