package broadcast

import "errors"

// ErrNilKeysHandler signals that a nil keys handler was provided
var ErrNilKeysHandler = errors.New("nil keys handler")

// ErrNilDelayedBroadcaster signals that a nil delayed broadcaster was provided
var ErrNilDelayedBroadcaster = errors.New("nil delayed broadcaster")
