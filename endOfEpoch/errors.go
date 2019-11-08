package endOfEpoch

import "errors"

var ErrNilArgsNewMetaEndOfEpochTrigger = errors.New("text")
var ErrNilRounder = errors.New("text")
var ErrNilSettingsHandler = errors.New("text")
var ErrInvalidSettingsForEndOfEpochTrigger = errors.New("text")
var ErrNilSyncTimer = errors.New("text")
var ErrNilArgsNewShardEndOfEpochTrigger = errors.New("text")
var ErrNilArgsPendingMiniblocks = errors.New("text")

// ErrWongTypeAssertion signals wrong type assertion
var ErrWongTypeAssertion = errors.New("wrong type assertion")

// ErrNilMarshalizer signals that nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilStorage signals that nil storage has been provided
var ErrNilStorage = errors.New("nil storage")
