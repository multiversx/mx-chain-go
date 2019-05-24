package core

import (
	"errors"
)

// ErrInvalidShardId signals that the shard id is invalid
var ErrInvalidShardId = errors.New("invalid shard id")

// ErrInvalidRoundDuration signals that an invalid round duration was provided
var ErrInvalidRoundDuration = errors.New("invalid round duration")

// ErrNilMarshalizer signals that a new marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer provided")

// ErrInvalidValue signals that a nil value has been provided
var ErrInvalidValue = errors.New("invalid value provided")

// ErrNilSendHandler signals that a nil send handler pointer has been provided
var ErrNilSendHandler = errors.New("nil send handler provided")
