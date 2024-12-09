package broadcast

import "errors"

// ErrNilKeysHandler signals that a nil keys handler was provided
var ErrNilKeysHandler = errors.New("nil keys handler")

var errNilDelayedShardBroadCaster = errors.New("nil delayed shard broadcaster provided")
