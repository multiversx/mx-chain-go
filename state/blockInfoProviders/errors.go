package blockInfoProviders

import "errors"

// ErrNilChainHandler signals that a nil chain handler was provided
var ErrNilChainHandler = errors.New("nil chain handler")
