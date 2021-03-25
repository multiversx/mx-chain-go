package epochproviders

import "errors"

// ErrNilRequestHandler signals that a nil request handler has been provided
var ErrNilRequestHandler = errors.New("nil request handler")

// ErrNilEpochStartMetaBlockInterceptor signals that a nil epoch start meta block interceptor has been provided
var ErrNilEpochStartMetaBlockInterceptor = errors.New("nil epoch start meta block interceptor")

// ErrNilMessenger signals that a nil messenger has been provided
var ErrNilMessenger = errors.New("nil messenger")

// ErrCannotGetLatestEpochStartMetaBlock signals that the latest epoch start meta block cannot be fetched from the network
var ErrCannotGetLatestEpochStartMetaBlock = errors.New("cannot fetch the latest epoch start meta block from the network: timeout")

// ErrComponentClosing signals that the latest epoch start meta block cannot be fetched from network because the component is closing
var ErrComponentClosing = errors.New("cannot fetch the latest epoch start meta block from the network: context closing")

// ErrInvalidRoundsPerEpoch signals that the provided RoundsPerEpoch value is incorrect
var ErrInvalidRoundsPerEpoch = errors.New("invalid rounds per epoch value")

// ErrInvalidRoundTimeInMilliseconds signals that the provided RoundTimeInMilliseconds value is incorrect
var ErrInvalidRoundTimeInMilliseconds = errors.New("invalid round time in milliseconds value")

// ErrInvalidStartTime signals that the provided start time is incorrect
var ErrInvalidStartTime = errors.New("invalid start time value")
