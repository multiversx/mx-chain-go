package provider

import "errors"

var ErrTypeAssertionFailed = errors.New("type assertion is not possible")

var ErrInvalidAddressLength = errors.New("invalid length for the node address")

var ErrInvalidFetchInterval = errors.New("invalid fetch interval has been provided")
