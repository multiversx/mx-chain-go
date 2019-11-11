package endOfEpoch

import "errors"

// ErrNilArgsNewMetaEndOfEpochTrigger signals that nil arguments were provided
var ErrNilArgsNewMetaEndOfEpochTrigger = errors.New("nil arguments for meta end of epoch trigger")

// ErrNilRounder signals that nil round was provided
var ErrNilRounder = errors.New("nil rounder")

// ErrNilEndOfEpochSettings signals that nil end of epoch settings has been provided
var ErrNilEndOfEpochSettings = errors.New("nil end of epoch settings")

// ErrInvalidSettingsForEndOfEpochTrigger signals that settings for end of epoch trigger are invalid
var ErrInvalidSettingsForEndOfEpochTrigger = errors.New("invalid end of epoch trigger settings")

// ErrNilSyncTimer signals that sync timer is nil
var ErrNilSyncTimer = errors.New("nil sync timer")

// ErrNilArgsNewShardEndOfEpochTrigger signals that nil arguments for shard epoch trigger has been provided
var ErrNilArgsNewShardEndOfEpochTrigger = errors.New("nil arguments for shard end of epoch trigger")

// ErrNotEnoughRoundsBetweenEpochs signals that not enough rounds has passed since last epoch start
var ErrNotEnoughRoundsBetweenEpochs = errors.New("tried to force end of epoch before passing of enough rounds")

// ErrForceEndOfEpochCanBeCalledOnNewRound signals that force end of epoch was called on wrong round
var ErrForceEndOfEpochCanBeCalledOnNewRound = errors.New("invalid time to call force end of epoch, possible only on new round")

// ErrSavedRoundIsHigherThanSaved signals that input round was wrong
var ErrSavedRoundIsHigherThanSaved = errors.New("saved round is higher than input round")

// ErrWrongTypeAssertion signals wrong type assertion
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilMarshalizer signals that nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrNilStorage signals that nil storage has been provided
var ErrNilStorage = errors.New("nil storage")

// ErrNilHeaderHandler signal that a nil header handler has been provided
var ErrNilHeaderHandler = errors.New("nil header handler")

// ErrNilArgsPendingMiniblocks signals that nil argument was passed
var ErrNilArgsPendingMiniblocks = errors.New("nil arguments for pending miniblock object")
